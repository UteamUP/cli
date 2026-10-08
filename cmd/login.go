package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/config"
	"github.com/uteamup/cli/internal/logging"
	"github.com/uteamup/cli/internal/security"
)

var (
	loginAPIKey      string
	loginAPISecret   string
	loginSecretFile  string
	loginSecretStdin bool
	loginKeyFile     string
	loginKeyAuth     bool
	loginSAML        bool
	loginCompany     string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with UteamUP",
	Long: `Authenticate with UteamUP using email/password, company SSO, or API key.

Interactive login (email/password):
  uteamup login
  ut login

Company SSO (system browser):
  uteamup login --saml --company iteggs --profile dev

API key authentication:
  uteamup login --api-key-auth
  ut login --api-key-file key.txt --api-secret-file secret.txt

The resulting JWT token is cached at ~/.uteamup/token.json and used
for all subsequent commands until it expires or you run "uteamup logout".`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().StringVar(&loginAPIKey, "api-key", "", "API key (32 characters) for OAuth 2.0 + PKCE auth")
	loginCmd.Flags().StringVar(&loginAPISecret, "api-secret", "", "Rejected: use --api-secret-file or --api-secret-stdin")
	loginCmd.Flags().StringVar(&loginSecretFile, "api-secret-file", "", "Owner-only API secret file")
	loginCmd.Flags().BoolVar(&loginSecretStdin, "api-secret-stdin", false, "Read API secret from stdin")
	loginCmd.Flags().StringVar(&loginKeyFile, "api-key-file", "", "Owner-only API key file")
	loginCmd.Flags().BoolVar(&loginKeyAuth, "api-key-auth", false, "Prompt for API credentials without echo")
	loginCmd.Flags().BoolVar(&loginSAML, "saml", false, "Sign in through your company SAML identity provider")
	loginCmd.Flags().StringVar(&loginCompany, "company", "", "Company code for SAML sign-in")
}

func runLogin(cmd *cobra.Command, args []string) error {
	if loginAPISecret != "" || loginAPIKey != "" {
		return fmt.Errorf("credentials in argv are unsafe; use protected file/stdin inputs or --api-key-auth")
	}
	if loginSecretFile != "" && loginSecretStdin {
		return fmt.Errorf("choose one API secret input")
	}
	apiKeyLogin := loginKeyAuth || loginKeyFile != "" || loginSecretFile != "" || loginSecretStdin
	if loginSAML && apiKeyLogin {
		return fmt.Errorf("choose SAML login or API key login")
	}
	if loginCompany != "" && !loginSAML {
		return fmt.Errorf("--company requires --saml")
	}
	if loginSAML && loginCompany == "" {
		return fmt.Errorf("--saml requires --company")
	}
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

	origin, originErr := security.Origin(baseURL)
	if originErr != nil {
		return originErr
	}
	authClient := auth.NewClient(origin, insecure, logger)

	var token *auth.TokenData

	if loginSAML {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()
		token, err = authClient.LoginWithSAML(ctx, loginCompany)
		if err != nil {
			return err
		}
	} else if apiKeyLogin {
		// API Key auth flow
		apiKey, secret := "", ""
		if loginKeyFile != "" {
			apiKey, err = security.SecretFile(loginKeyFile)
			if err != nil {
				return err
			}
		}
		if loginSecretFile != "" {
			secret, err = security.SecretFile(loginSecretFile)
			if err != nil {
				return err
			}
		}
		if loginSecretStdin {
			secret, err = security.SecretStdin(os.Stdin)
			if err != nil {
				return err
			}
		}

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

	token.APIOrigin = origin
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
	} else if token.AuthMethod == "saml" {
		method = "company SSO"
	}
	fmt.Printf("Authenticated successfully via %s.\n", method)
	if token.Email != "" {
		fmt.Printf("Logged in as: %s\n", security.SafeText(token.Email))
	}
	if token.TenantName != "" {
		// Print the GUID, not the int Id — internal database keys must not leak to
		// user-facing CLI output per the GUIDs-at-boundary rule.
		if token.TenantGUID != "" {
			fmt.Printf("Tenant: %s (%s)\n", security.SafeText(token.TenantName), token.TenantGUID)
		} else {
			fmt.Printf("Tenant: %s\n", security.SafeText(token.TenantName))
		}
	}
	fmt.Printf("Token expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05 UTC"))

	return nil
}
