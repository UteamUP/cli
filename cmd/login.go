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

	// Determine the target after Cobra has parsed persistent flags. This keeps
	// login on the same profile and backend as subsequent domain commands.
	baseURL := "https://api.uteamup.com"
	selectedProfile := ""
	cfg, err := config.Load()
	if err == nil {
		profile, name, profErr := selectedProfileConfig(cfg, profileName)
		if profErr != nil {
			return profErr
		}
		selectedProfile = name
		if profile.BaseURL != "" {
			baseURL = profile.BaseURL
		}
	} else if profileName != "" {
		return fmt.Errorf("loading config profile %q: %w", profileName, err)
	}
	if envBaseURL := os.Getenv("UTEAMUP_API_BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	}

	authClient := auth.NewClient(baseURL, insecure, logger)

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
	if selectedProfile != "" {
		token.Profile = selectedProfile
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
		// Print the GUID, not the int Id — internal database keys must not leak to
		// user-facing CLI output per the GUIDs-at-boundary rule.
		if token.TenantGUID != "" {
			fmt.Printf("Tenant: %s (%s)\n", token.TenantName, token.TenantGUID)
		} else {
			fmt.Printf("Tenant: %s\n", token.TenantName)
		}
	}
	fmt.Printf("Token expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05 UTC"))

	return nil
}
