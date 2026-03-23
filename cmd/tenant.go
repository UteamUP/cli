package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/config"
)

var tenantCmd = &cobra.Command{
	Use:     "tenant",
	Aliases: []string{"tenants"},
	Short:   "Manage and view tenant information",
	Long: `View and manage the tenants you have access to.

Examples:
  uteamup tenant show          # List all tenants you have access to
  ut tenant show               # Same, using shortname`,
}

var tenantShowCmd = &cobra.Command{
	Use:     "show",
	Aliases: []string{"list", "ls"},
	Short:   "Show all tenants you have access to",
	Long: `Display a list of all tenants associated with your account,
including tenant name, GUID, plan, and active status.

The currently active tenant (from login) is marked with an asterisk (*).
If a tenantGuid is set in your config profile, it is marked with a caret (^).

Examples:
  uteamup tenant show
  ut tenant list
  ut tenant ls`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load token for auth.
		token, err := auth.LoadToken()
		if err != nil || token == nil || !token.IsValid() {
			fmt.Fprintln(os.Stderr, "Not authenticated. Run \"uteamup login\" first.")
			os.Exit(1)
		}

		// Load config for base URL and tenantGuid override.
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		profile, err := cfg.ActiveProfileConfig()
		if err != nil {
			return fmt.Errorf("loading profile: %w", err)
		}

		// Fetch all tenants.
		tenants, err := auth.FetchAllTenants(token.AccessToken, profile.BaseURL)
		if err != nil {
			return fmt.Errorf("fetching tenants: %w", err)
		}

		if len(tenants) == 0 {
			fmt.Println("No tenants found for your account.")
			return nil
		}

		// Print header.
		fmt.Printf("\nTenants for %s (%d total)\n", token.Email, len(tenants))
		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  \tNAME\tGUID\tPLAN\tSTATUS")
		fmt.Fprintln(w, "  \t----\t----\t----\t------")

		for _, t := range tenants {
			// Mark current tenant.
			marker := " "
			if t.Guid == token.TenantGuid {
				marker = "*"
			}
			if profile.TenantGuid != "" && t.Guid == profile.TenantGuid {
				marker = "^"
			}

			plan := "(no plan)"
			if t.HasPlan() {
				plan = t.PlanName
			}

			status := "inactive"
			if t.IsActive {
				status = "active"
			}

			fmt.Fprintf(w, "%s \t%s\t%s\t%s\t%s\n", marker, t.Name, t.Guid, plan, status)
		}
		w.Flush()

		fmt.Println()
		fmt.Println("  * = currently logged-in tenant")
		if profile.TenantGuid != "" {
			fmt.Println("  ^ = tenant set in config profile")
		}
		fmt.Println()
		fmt.Println("To switch tenants, set tenantGuid in your config:")
		fmt.Println("  ut config set tenantGuid <GUID>")
		fmt.Println()

		return nil
	},
}

var tenantSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Interactively select and save a tenant to your config",
	Long: `Fetches all tenants you have access to, presents a numbered list,
and saves the selected tenant's GUID to your config profile.

This tenant will be used for all subsequent commands that require a tenant context.

Examples:
  uteamup tenant select
  ut tenant select`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load token for auth.
		token, err := auth.LoadToken()
		if err != nil || token == nil || !token.IsValid() {
			fmt.Fprintln(os.Stderr, "Not authenticated. Run \"uteamup login\" first.")
			os.Exit(1)
		}

		// Load config.
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		profile, err := cfg.ActiveProfileConfig()
		if err != nil {
			return fmt.Errorf("loading profile: %w", err)
		}

		// Fetch all tenants.
		tenants, err := auth.FetchAllTenants(token.AccessToken, profile.BaseURL)
		if err != nil {
			return fmt.Errorf("fetching tenants: %w", err)
		}

		if len(tenants) == 0 {
			fmt.Println("No tenants found for your account.")
			return nil
		}

		// Display list.
		fmt.Printf("\nSelect a tenant for %s:\n\n", token.Email)
		for i, t := range tenants {
			current := " "
			if t.Guid == token.TenantGuid {
				current = "*"
			}
			plan := "(no plan)"
			if t.HasPlan() {
				plan = t.PlanName
			}
			fmt.Printf("  %s %d. %s [%s]\n", current, i+1, t.Name, plan)
		}

		// Prompt for selection.
		fmt.Printf("\nSelect tenant (1-%d): ", len(tenants))
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("no selection made")
		}
		choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || choice < 1 || choice > len(tenants) {
			return fmt.Errorf("invalid selection: choose 1-%d", len(tenants))
		}

		selected := tenants[choice-1]

		// Save tenant GUID to config profile.
		p := cfg.Profiles[cfg.ActiveProfile]
		p.TenantGuid = selected.Guid
		cfg.Profiles[cfg.ActiveProfile] = p

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		// Update the cached token so `ut auth status` reflects the active tenant.
		token.TenantID = selected.ID
		token.TenantGuid = selected.Guid
		token.TenantName = selected.Name
		if err := auth.SaveToken(token); err != nil {
			return fmt.Errorf("updating token: %w", err)
		}

		fmt.Printf("\nActive tenant set to: %s (%s)\n", selected.Name, selected.Guid)
		fmt.Printf("Saved to profile: %s\n", cfg.ActiveProfile)

		return nil
	},
}

func init() {
	tenantCmd.AddCommand(tenantShowCmd)
	tenantCmd.AddCommand(tenantSelectCmd)
}
