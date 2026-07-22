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
	"auth":       true,
	"completion": true,
	"config":     true,
	"help":       true,
	"health":     true,
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
	rootCmd.AddCommand(videoCmd)
	rootCmd.AddCommand(tenantCmd)

	// Register domain commands
	registerDomainCommands()
}

func registerDomainCommands() {
	logger := logging.New(logging.LevelInfo)
	exportCfg := &registry.ExportConfig{}

	for _, cmd := range registry.DefaultRegistry.BuildCommands(func() (*client.APIClient, error) {
		return newDomainAPIClient(logger, exportCfg)
	}, logger, &outputFormat, exportCfg) {
		rootCmd.AddCommand(cmd)
	}
}

func selectedProfileConfig(cfg *config.Config, requested string) (*config.Profile, string, error) {
	name := requested
	if name == "" {
		name = cfg.ActiveProfile
	}

	profile, ok := cfg.Profiles[name]
	if !ok {
		return nil, "", fmt.Errorf("config profile %q not found", name)
	}

	// Environment values override whichever profile was selected, including a
	// non-active profile chosen with --profile.
	if value := os.Getenv("UTEAMUP_API_BASE_URL"); value != "" {
		profile.BaseURL = value
	}
	if value := os.Getenv("UTEAMUP_LOG_LEVEL"); value != "" {
		profile.LogLevel = value
	}

	return &profile, name, nil
}

func newDomainAPIClient(logger *logging.Logger, exportCfg *registry.ExportConfig) (*client.APIClient, error) {
	cfg, err := config.Load()
	if err != nil {
		if profileName != "" {
			return nil, fmt.Errorf("loading config profile %q: %w", profileName, err)
		}

		baseURL := os.Getenv("UTEAMUP_API_BASE_URL")
		if baseURL == "" {
			baseURL = "https://api.uteamup.com"
		}
		level := logging.LevelInfo
		if verbose {
			level = logging.LevelDebug
		}
		logger.SetLevel(level)
		exportCfg.Enabled = false
		exportCfg.Dir = ""
		return client.NewAPIClient(
			baseURL,
			30*time.Second,
			insecure,
			client.DefaultRetryOptions(),
			logger,
		), nil
	}

	profile, _, err := selectedProfileConfig(cfg, profileName)
	if err != nil {
		return nil, err
	}

	level := logging.ParseLevel(profile.LogLevel)
	if verbose {
		level = logging.LevelDebug
	}
	logger.SetLevel(level)
	exportCfg.Enabled = profile.ExportJSON
	exportCfg.Dir = profile.ExportDir

	timeout := time.Duration(profile.RequestTimeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	retryOpts := client.RetryOptions{
		MaxRetries: profile.MaxRetries,
		BaseDelay:  time.Second,
		MaxDelay:   10 * time.Second,
	}
	return client.NewAPIClient(profile.BaseURL, timeout, insecure, retryOpts, logger), nil
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
