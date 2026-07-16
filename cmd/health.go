package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/config"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check CLI authentication and backend health",
	Long: `Reports the environment the CLI is authenticated against and whether
the backend health endpoint is reachable. If no valid token is cached, outputs
"environment: none" and "status: none".`,
	RunE: runHealth,
}

func runHealth(cmd *cobra.Command, args []string) error {
	token, err := auth.LoadToken()
	if err != nil {
		return fmt.Errorf("reading token: %w", err)
	}

	if token == nil || !token.IsValid() {
		fmt.Println("environment: none")
		fmt.Println("status: none")
		return nil
	}

	baseURL := resolveHealthBaseURL()
	env := environmentFromBaseURL(baseURL)

	status := "unhealthy"
	if checkBackendHealth(baseURL) {
		status = "healthy"
	}

	fmt.Printf("environment: %s\n", env)
	fmt.Printf("status: %s\n", status)
	return nil
}

// resolveHealthBaseURL returns the API base URL the CLI would use for calls.
// Precedence: UTEAMUP_API_BASE_URL env var, active profile baseUrl, prod default.
func resolveHealthBaseURL() string {
	if url := os.Getenv("UTEAMUP_API_BASE_URL"); url != "" {
		return url
	}

	cfg, err := config.Load()
	if err == nil {
		if profile, profErr := cfg.ActiveProfileConfig(); profErr == nil && profile.BaseURL != "" {
			return profile.BaseURL
		}
	}

	return "https://api.uteamup.com"
}

// environmentFromBaseURL maps a backend URL to a human environment name.
func environmentFromBaseURL(baseURL string) string {
	lower := strings.ToLower(baseURL)
	switch {
	case strings.Contains(lower, "localhost"), strings.Contains(lower, "127.0.0.1"):
		return "localhost"
	case strings.Contains(lower, "pruf"), strings.Contains(lower, "staging"):
		return "staging"
	case strings.Contains(lower, "dev"):
		return "dev"
	default:
		return "production"
	}
}

// checkBackendHealth performs a best-effort GET against /health.
func checkBackendHealth(baseURL string) bool {
	url := strings.TrimRight(baseURL, "/") + "/health"
	client := newHealthClient()

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func newHealthClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: insecure, //nolint:gosec // user-requested dev flag
			},
		},
	}
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
