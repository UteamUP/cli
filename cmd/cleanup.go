package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/cleanup"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/config"
	"github.com/uteamup/cli/internal/logging"
)

var (
	cleanupRoot    string
	cleanupOut     string
	cleanupType    string
	cleanupMinDays int
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Find code that is never used at runtime (global admin only)",
	Long: `Scan the UteamUP monorepo for every code unit (backend endpoints, repositories, frontend
pages/components, mobile pages/components), pull runtime usage from the selected environment, and
write cleanup_report.md listing what exists in code but is never exercised.

Requires the Usage Verifier to be enabled by a global admin (uteamup config selects the environment;
GET /usage is global-admin only). Run from the monorepo root, or pass --root.

Examples:
  uteamup cleanup --insecure                       # localhost (self-signed cert)
  uteamup --profile production cleanup             # compare against production usage
  uteamup cleanup --type FrontendPage              # only one type
  uteamup cleanup --out reports/cleanup.md`,
	RunE: runCleanup,
}

func init() {
	cleanupCmd.Flags().StringVar(&cleanupRoot, "root", "", "Monorepo root (auto-detected if omitted)")
	cleanupCmd.Flags().StringVar(&cleanupOut, "out", "cleanup_report.md", "Path to write the report")
	cleanupCmd.Flags().StringVar(&cleanupType, "type", "", "Limit to one usage type (e.g. FrontendPage)")
	cleanupCmd.Flags().IntVar(&cleanupMinDays, "min-days", 14, "Minimum observation window before trusting 'unused'")
	rootCmd.AddCommand(cleanupCmd)
}

type usageStatRow struct {
	Type         string `json:"type"`
	Key          string `json:"key"`
	HitCount     int64  `json:"hitCount"`
	LastSeenDate string `json:"lastSeenDate"`
}

type usagePaged struct {
	Items       []usageStatRow `json:"items"`
	CurrentPage int            `json:"currentPage"`
	PageSize    int            `json:"pageSize"`
	TotalItems  int            `json:"totalItems"`
	TotalPages  int            `json:"totalPages"`
}

type usageStatus struct {
	Enabled      bool    `json:"enabled"`
	EnabledSince *string `json:"enabledSince"`
}

func runCleanup(cmd *cobra.Command, args []string) error {
	token, err := auth.LoadToken()
	if err != nil || token == nil || !token.IsValid() {
		fmt.Fprintln(os.Stderr, "Not authenticated. Run \"uteamup login\" first.")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	profile, err := cfg.ActiveProfileConfig()
	if err != nil {
		return fmt.Errorf("loading profile: %w", err)
	}

	root, err := resolveRoot(cleanupRoot)
	if err != nil {
		return err
	}

	timeout := time.Duration(profile.RequestTimeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	apiClient := client.NewAPIClient(profile.BaseURL, timeout, insecure, client.DefaultRetryOptions(), logging.New(logging.LevelError))
	ctx := context.Background()

	status, err := fetchStatus(ctx, apiClient)
	if err != nil {
		return friendlyAPIErr(err)
	}

	usage, err := fetchAllUsage(ctx, apiClient)
	if err != nil {
		return friendlyAPIErr(err)
	}

	fmt.Printf("Scanning monorepo at %s ...\n", root)
	catalog := cleanup.Scan(root)

	in := cleanup.ReportInput{
		Catalog:      catalog,
		Usage:        toUsageRows(usage),
		Env:          profile.BaseURL,
		EnabledSince: parseSince(status.EnabledSince),
		Now:          time.Now(),
		MinDays:      cleanupMinDays,
		FilterType:   cleanupType,
	}
	summaries := cleanup.BuildSummaries(in)

	if outputFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summaries)
	}

	if err := cleanup.WriteMarkdown(in, summaries, cleanupOut); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	totalUnused := 0
	for _, s := range summaries {
		totalUnused += s.Unused
		fmt.Printf("  %-18s in-code=%-5d used=%-5d unused=%-5d\n", s.Type, s.TotalInCode, s.Used, s.Unused)
	}
	fmt.Printf("\n%d unused unit(s). Report written to %s\n", totalUnused, cleanupOut)
	if !status.Enabled {
		fmt.Fprintln(os.Stderr, "\nNote: usage tracking is currently OFF for this environment — results reflect only past data.")
	}
	return nil
}

func fetchStatus(ctx context.Context, c *client.APIClient) (usageStatus, error) {
	var s usageStatus
	raw, err := c.CallREST(ctx, "GET", "/api/usage/status", map[string]any{}, nil, "cleanup-status")
	if err != nil {
		return s, err
	}
	_ = json.Unmarshal(raw, &s)
	return s, nil
}

func fetchAllUsage(ctx context.Context, c *client.APIClient) ([]usageStatRow, error) {
	var all []usageStatRow
	page := 1
	for {
		raw, err := c.CallREST(ctx, "GET", "/api/usage", map[string]any{"page": page, "pageSize": 200}, nil, "cleanup-usage")
		if err != nil {
			return nil, err
		}
		var pg usagePaged
		if err := json.Unmarshal(raw, &pg); err != nil {
			return nil, fmt.Errorf("parsing usage response: %w", err)
		}
		all = append(all, pg.Items...)
		if pg.CurrentPage >= pg.TotalPages || len(pg.Items) == 0 {
			break
		}
		page++
	}
	return all, nil
}

func toUsageRows(rows []usageStatRow) []cleanup.UsageRow {
	out := make([]cleanup.UsageRow, len(rows))
	for i, r := range rows {
		out[i] = cleanup.UsageRow{Type: r.Type, Key: r.Key, HitCount: r.HitCount, LastSeenDate: r.LastSeenDate}
	}
	return out
}

func parseSince(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, *s); err == nil {
		return &t
	}
	return nil
}

func friendlyAPIErr(err error) error {
	msg := err.Error()
	if strings.Contains(msg, "403") || strings.Contains(strings.ToLower(msg), "forbidden") {
		return fmt.Errorf("this command requires a global-admin account (GET /usage returned 403): %w", err)
	}
	return err
}

// resolveRoot returns the monorepo root: the explicit flag, or the nearest ancestor of the working
// directory that contains UteamUP_Backend/.
func resolveRoot(flag string) (string, error) {
	if flag != "" {
		return filepath.Abs(flag)
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if st, err := os.Stat(filepath.Join(dir, "UteamUP_Backend")); err == nil && st.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find the monorepo root (no UteamUP_Backend/ in any parent of the working directory); pass --root")
		}
		dir = parent
	}
}
