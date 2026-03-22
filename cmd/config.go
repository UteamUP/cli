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

		// Export JSON defaults: enabled for development profiles, disabled for production
		isDev := strings.Contains(strings.ToLower(name), "dev") ||
			strings.Contains(baseURL, "localhost")
		exportDefault := "N"
		if isDev {
			exportDefault = "Y"
		}
		fmt.Printf("Export JSON responses to file? [%s] (y/n): ", exportDefault)
		exportAnswer, _ := reader.ReadString('\n')
		exportAnswer = strings.TrimSpace(strings.ToLower(exportAnswer))
		exportJSON := isDev // default based on profile type
		if exportAnswer == "y" || exportAnswer == "yes" {
			exportJSON = true
		} else if exportAnswer == "n" || exportAnswer == "no" {
			exportJSON = false
		}

		exportDir := ""
		if exportJSON {
			fmt.Print("Export directory [~/.uteamup/exports]: ")
			exportDir, _ = reader.ReadString('\n')
			exportDir = strings.TrimSpace(exportDir)
		}

		fmt.Print("\n--- Gemini AI (Image Analysis) ---\n")
		fmt.Print("Gemini API key (press Enter to skip): ")
		geminiKey, _ := reader.ReadString('\n')
		geminiKey = strings.TrimSpace(geminiKey)

		fmt.Println("Available models: gemini-pro-latest, gemini-3.1-pro-preview, gemini-3.1-flash-lite-preview, gemini-2.5-pro, gemini-2.5-flash")
		fmt.Print("Gemini model [gemini-3.1-flash-lite-preview]: ")
		geminiModel, _ := reader.ReadString('\n')
		geminiModel = strings.TrimSpace(geminiModel)
		if geminiModel == "" {
			geminiModel = "gemini-3.1-flash-lite-preview"
		}

		profile := config.Profile{
			Name:           name,
			APIKey:         apiKey,
			Secret:         secret,
			BaseURL:        baseURL,
			LogLevel:       "INFO",
			RequestTimeout: 30000,
			MaxRetries:     3,
			ExportJSON:     exportJSON,
			ExportDir:      exportDir,
			GeminiAPIKey:   geminiKey,
			GeminiModel:    geminiModel,
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

Valid keys: baseUrl, apiKey, secret, logLevel, requestTimeout, maxRetries, exportJson, exportDir

Examples:
  uteamup config set baseUrl https://localhost:5002
  ut config set logLevel DEBUG
  ut config set exportJson true
  ut config set exportDir ~/my-exports`,
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
		case "exportJson", "exportjson":
			profile.ExportJSON = strings.ToLower(value) == "true" || value == "1" || strings.ToLower(value) == "yes"
		case "exportDir", "exportdir":
			profile.ExportDir = value
		case "geminiApiKey", "geminiapikey":
			profile.GeminiAPIKey = value
		case "geminiModel", "geminimodel":
			profile.GeminiModel = value
		case "googleMapsApiKey", "googlemapsapikey", "mapsApiKey", "mapsapikey":
			profile.GoogleMapsAPIKey = value
		default:
			return fmt.Errorf("unknown config key %q — valid keys: baseUrl, apiKey, secret, logLevel, requestTimeout, maxRetries, exportJson, exportDir, geminiApiKey, geminiModel, googleMapsApiKey", key)
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

var configAPIKeyCmd = &cobra.Command{
	Use:   "apikey [key]",
	Short: "Get or set the Gemini API key",
	Long: `Get or set the Gemini API key for image analysis.

Examples:
  ut config apikey                     # Show current key (redacted)
  ut config apikey AIzaSy...           # Set the key
  uteamup config apikey=AIzaSy...      # Also works with = syntax`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		profile, err := cfg.ActiveProfileConfig()
		if err != nil {
			return err
		}

		// Parse key=value syntax from command name
		value := ""
		if len(args) == 1 {
			value = args[0]
			// Handle apikey=value syntax
			if strings.Contains(value, "=") {
				value = strings.SplitN(value, "=", 2)[1]
			}
		}

		if value == "" {
			// Show current key
			display := "(not set)"
			if profile.GeminiAPIKey != "" {
				if len(profile.GeminiAPIKey) > 8 {
					display = profile.GeminiAPIKey[:8] + "..." + profile.GeminiAPIKey[len(profile.GeminiAPIKey)-4:]
				} else {
					display = "***"
				}
			}
			fmt.Printf("Gemini API Key: %s\n", display)
			return nil
		}

		profile.GeminiAPIKey = value
		cfg.Profiles[cfg.ActiveProfile] = *profile
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Gemini API key updated (profile: %s)\n", cfg.ActiveProfile)
		return nil
	},
}

var configModelCmd = &cobra.Command{
	Use:   "model [name]",
	Short: "Get or set the default Gemini model",
	Long: `Get or set the default Gemini model for image analysis.

Examples:
  ut config model                              # Show current model
  ut config model gemini-3.1-pro-preview       # Set default model
  ut config model list                         # List available models
  uteamup config model=gemini-2.5-flash        # Also works with = syntax`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		profile, err := cfg.ActiveProfileConfig()
		if err != nil {
			return err
		}

		value := ""
		if len(args) == 1 {
			value = args[0]
			if strings.Contains(value, "=") {
				value = strings.SplitN(value, "=", 2)[1]
			}
		}

		if value == "" {
			// Show current model
			model := profile.GeminiModel
			if model == "" {
				model = "gemini-3.1-flash-lite-preview (default)"
			}
			fmt.Printf("Gemini Model: %s\n", model)
			return nil
		}

		if value == "list" {
			fmt.Println("Available Gemini models for image analysis:")
			fmt.Println()
			fmt.Println("  Pro models (higher accuracy, slower):")
			fmt.Println("    gemini-pro-latest              Always points to newest pro (rolling)")
			fmt.Println("    gemini-3.1-pro-preview         Latest explicit pro")
			fmt.Println("    gemini-3-pro-preview           Previous gen pro")
			fmt.Println("    gemini-2.5-pro                 Stable pro")
			fmt.Println()
			fmt.Println("  Flash models (faster, cheaper):")
			fmt.Println("    gemini-3.1-flash-lite-preview  Default — fastest and cheapest")
			fmt.Println("    gemini-3-flash-preview         Previous gen flash")
			fmt.Println("    gemini-2.5-flash               Stable flash")
			fmt.Println()
			current := profile.GeminiModel
			if current == "" {
				current = "gemini-3.1-flash-lite-preview"
			}
			fmt.Printf("  Current: %s\n", current)
			return nil
		}

		profile.GeminiModel = value
		cfg.Profiles[cfg.ActiveProfile] = *profile
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Gemini model set to %s (profile: %s)\n", value, cfg.ActiveProfile)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configProfileCmd)
	configCmd.AddCommand(configAPIKeyCmd)
	configCmd.AddCommand(configModelCmd)
}
