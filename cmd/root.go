package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/config"
	"github.com/uteamup/cli/internal/logging"
	"github.com/uteamup/cli/internal/registry"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"

	outputFormat string
	profileName  string
	verbose      bool
	insecure     bool
)

// SetBuildInfo sets the build metadata from main.go ldflags.
func SetBuildInfo(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
}

// commandsExemptFromAuth lists commands that don't require authentication.
var commandsExemptFromAuth = map[string]bool{
	"login":      true,
	"logout":     true,
	"version":    true,
	"completion": true,
	"config":     true,
	"help":       true,
}

var rootCmd = &cobra.Command{
	Use:   "uteamup",
	Short: "UteamUP CLI — command-line interface for the UteamUP platform",
	Long: `UteamUP CLI provides direct terminal access to the UteamUP platform.

Manage assets, work orders, users, and more from any terminal.
Authenticate with email/password or API key, then run commands.

Aliases: uteamup, ut

Examples:
  uteamup login                           # Interactive login
  ut login --api-key=KEY --api-secret=SEC # API key auth
  ut asset list                           # List assets
  ut workorder get 123 -o json            # Get work order as JSON`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip auth check for exempt commands
		if commandsExemptFromAuth[cmd.Name()] {
			return nil
		}
		// Check parent command too (e.g., "config init")
		if cmd.Parent() != nil && commandsExemptFromAuth[cmd.Parent().Name()] {
			return nil
		}

		token, err := auth.LoadToken()
		if err != nil {
			return fmt.Errorf("checking auth status: %w", err)
		}
		if token == nil || !token.IsValid() {
			fmt.Fprintln(os.Stderr, "Not authenticated. Run \"uteamup login\" or \"ut login\" first.")
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().StringVarP(&profileName, "profile", "P", "", "Config profile to use (overrides activeProfile)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose/debug logging")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (for dev)")

	// Register subcommands
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(imageCmd)

	// Register domain commands
	registerDomainCommands()
}

func registerDomainCommands() {
	logLevel := logging.LevelInfo
	if verbose {
		logLevel = logging.LevelDebug
	}
	logger := logging.New(logLevel)

	// Load config for API client (best-effort — domains register even without config)
	cfg, err := config.Load()
	var apiClient *client.APIClient
	if err == nil {
		profile, profErr := cfg.ActiveProfileConfig()
		if profErr == nil {
			timeout := time.Duration(profile.RequestTimeout) * time.Millisecond
			retryOpts := client.RetryOptions{
				MaxRetries: profile.MaxRetries,
				BaseDelay:  1 * time.Second,
				MaxDelay:   10 * time.Second,
			}
			apiClient = client.NewAPIClient(profile.BaseURL, timeout, insecure, retryOpts, logger)
		}
	}

	if apiClient == nil {
		// Fallback client for help/completions (won't be used for actual requests)
		apiClient = client.NewAPIClient("https://api.uteamup.com", 30*time.Second, false, client.DefaultRetryOptions(), logger)
	}

	// Build export config from active profile
	exportCfg := &registry.ExportConfig{}
	if cfg != nil {
		if profile, profErr := cfg.ActiveProfileConfig(); profErr == nil {
			exportCfg.Enabled = profile.ExportJSON
			exportCfg.Dir = profile.ExportDir
		}
	}

	for _, cmd := range registry.DefaultRegistry.BuildCommands(apiClient, logger, &outputFormat, exportCfg) {
		rootCmd.AddCommand(cmd)
	}
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
