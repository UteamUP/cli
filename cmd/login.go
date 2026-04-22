package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/config"
	"github.com/uteamup/cli/internal/logging"
)

var (
	loginAPIKey    string
	loginAPISecret string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with UteamUP",
	Long: `Authenticate with UteamUP using email/password or API key.

Interactive login (email/password):
  uteamup login
  ut login

API key authentication:
  uteamup login --api-key=KEY --api-secret=SECRET
  ut login --api-key=KEY --api-secret=SECRET

The resulting JWT token is cached at ~/.uteamup/token.json and used
for all subsequent commands until it expires or you run "uteamup logout".`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().StringVar(&loginAPIKey, "api-key", "", "API key (32 characters) for OAuth 2.0 + PKCE auth")
	loginCmd.Flags().StringVar(&loginAPISecret, "api-secret", "", "API secret (64+ characters) for OAuth 2.0 + PKCE auth")
}

func runLogin(cmd *cobra.Command, args []string) error {
	logger := logging.Default()
	if verbose {
		logger.SetLevel(logging.LevelDebug)
	}

	// Determine base URL. Prefer an explicit UTEAMUP_API_BASE_URL env var
	// (so the uteamup-debug skill and CI can point the CLI at a non-prod
	// backend without needing `uteamup config init` first), then an active
	// config profile's BaseURL, and only fall back to the hardcoded prod URL
	// when neither is set.
	baseURL := os.Getenv("UTEAMUP_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.uteamup.com"
	}
	cfg, err := config.Load()
	if err == nil {
		if profile, profErr := cfg.ActiveProfileConfig(); profErr == nil && profile.BaseURL != "" {
			baseURL = profile.BaseURL
		}
	}

	authClient := auth.NewAuthClient(baseURL, insecure, logger)

	var token *auth.TokenData

	if loginAPIKey != "" || loginAPISecret != "" {
		// API Key auth flow
		apiKey := loginAPIKey
		secret := loginAPISecret

		// Prompt for missing values
		if apiKey == "" || secret == "" {
			prompted, promptedSecret, err := auth.PromptAPIKey()
			if err != nil {
				return fmt.Errorf("reading API key: %w", err)
			}
			if apiKey == "" {
				apiKey = prompted
			}
			if secret == "" {
				secret = promptedSecret
			}
		}

		token, err = authClient.LoginWithAPIKey(apiKey, secret)
		if err != nil {
			return fmt.Errorf("API key authentication failed: %w", err)
		}
	} else {
		// Interactive login flow
		email, password, err := auth.PromptCredentials()
		if err != nil {
			return fmt.Errorf("reading credentials: %w", err)
		}

		token, err = authClient.LoginWithCredentials(email, password)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
	}

	// Save active profile name to token
	if cfg != nil {
		token.Profile = cfg.ActiveProfile
	}

	if err := auth.SaveToken(token); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	method := "email/password"
	if token.AuthMethod == "apikey" {
		method = "API key"
	}
	fmt.Printf("Authenticated successfully via %s.\n", method)
	if token.Email != "" {
		fmt.Printf("Logged in as: %s\n", token.Email)
	}
	if token.TenantName != "" {
		fmt.Printf("Tenant: %s (ID: %d)\n", token.TenantName, token.TenantID)
	}
	fmt.Printf("Token expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05 UTC"))

	return nil
}
