package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/uteamup/cli/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new config file interactively",
	Long: `Create ~/.uteamup/config.json with your connection settings.

You will be prompted for:
  - Profile name (default: production)
  - API base URL (default: https://api.uteamup.com)
  - API key and secret (optional, can be set later)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}

		// Check if config already exists
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Config file already exists at %s\n", path)
			fmt.Print("Overwrite? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(answer)) != "y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		reader := bufio.NewReader(os.Stdin)
		cfg := config.DefaultConfig()

		fmt.Print("Profile name [production]: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			name = "production"
		}

		fmt.Print("API base URL [https://api.uteamup.com]: ")
		baseURL, _ := reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			baseURL = "https://api.uteamup.com"
		}

		fmt.Print("API key (32 chars, press Enter to skip): ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)

		fmt.Print("API secret (64+ chars, press Enter to skip): ")
		secret, _ := reader.ReadString('\n')
		secret = strings.TrimSpace(secret)

		profile := config.Profile{
			Name:           name,
			APIKey:         apiKey,
			Secret:         secret,
			BaseURL:        baseURL,
			LogLevel:       "INFO",
			RequestTimeout: 30000,
			MaxRetries:     3,
		}

		profileKey := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		cfg.ActiveProfile = profileKey
		cfg.Profiles = map[string]config.Profile{profileKey: profile}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("\nConfig saved to %s\n", path)
		fmt.Printf("Active profile: %s\n", profileKey)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration (secrets redacted)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Println(cfg.RedactedSummary())
		fmt.Printf("\nProfiles: %s\n", strings.Join(cfg.ListProfiles(), ", "))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a value in the active profile's configuration.

Valid keys: baseUrl, apiKey, secret, logLevel, requestTimeout, maxRetries

Examples:
  uteamup config set baseUrl https://localhost:5002
  ut config set logLevel DEBUG
  ut config set maxRetries 5`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		profile, err := cfg.ActiveProfileConfig()
		if err != nil {
			return err
		}

		key, value := args[0], args[1]
		switch key {
		case "baseUrl", "baseurl":
			profile.BaseURL = value
		case "apiKey", "apikey":
			profile.APIKey = value
		case "secret":
			profile.Secret = value
		case "logLevel", "loglevel":
			profile.LogLevel = strings.ToUpper(value)
		case "requestTimeout":
			fmt.Sscanf(value, "%d", &profile.RequestTimeout)
		case "maxRetries":
			fmt.Sscanf(value, "%d", &profile.MaxRetries)
		case "name":
			profile.Name = value
		default:
			return fmt.Errorf("unknown config key %q — valid keys: baseUrl, apiKey, secret, logLevel, requestTimeout, maxRetries", key)
		}

		cfg.Profiles[cfg.ActiveProfile] = *profile
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Set %s = %s (profile: %s)\n", key, value, cfg.ActiveProfile)
		return nil
	},
}

var configProfileCmd = &cobra.Command{
	Use:   "profile <name>",
	Short: "Switch the active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if err := cfg.SetActiveProfile(args[0]); err != nil {
			return err
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Switched to profile: %s\n", args[0])
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configProfileCmd)
}
