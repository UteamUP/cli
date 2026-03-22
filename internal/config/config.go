package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	clierrors "github.com/uteamup/cli/internal/errors"
)

const (
	defaultBaseURL        = "https://api.uteamup.com"
	defaultLogLevel       = "INFO"
	defaultRequestTimeout = 30000
	defaultMaxRetries     = 3
	configDirName         = ".uteamup"
	configFileName        = "config.json"
)

// Config is the top-level config stored in ~/.uteamup/config.json.
type Config struct {
	ActiveProfile string             `json:"activeProfile"`
	Profiles      map[string]Profile `json:"profiles"`
}

// Profile holds connection settings for a single environment.
type Profile struct {
	Name           string `json:"name"`
	APIKey         string `json:"apiKey,omitempty"`
	Secret         string `json:"secret,omitempty"`
	BaseURL        string `json:"baseUrl"`
	LogLevel       string `json:"logLevel"`
	RequestTimeout int    `json:"requestTimeout"`
	MaxRetries     int    `json:"maxRetries"`
	ExportJSON     bool   `json:"exportJson"`
	ExportDir      string `json:"exportDir,omitempty"`
}

// ConfigDir returns ~/.uteamup.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, configDirName), nil
}

// ConfigPath returns ~/.uteamup/config.json.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// Load reads and validates the config file.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &clierrors.ConfigError{
				Message: fmt.Sprintf("config file not found at %s — run \"uteamup config init\" first", path),
			}
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, &clierrors.ConfigError{Message: fmt.Sprintf("invalid JSON in config: %v", err)}
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	applyEnvOverrides(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	path := filepath.Join(dir, configFileName)
	return os.WriteFile(path, data, 0600)
}

// ActiveProfile returns the currently active profile.
func (c *Config) ActiveProfileConfig() (*Profile, error) {
	if c.ActiveProfile == "" {
		return nil, &clierrors.ConfigError{Message: "no active profile set"}
	}
	p, ok := c.Profiles[c.ActiveProfile]
	if !ok {
		return nil, &clierrors.ConfigError{
			Field:   "activeProfile",
			Message: fmt.Sprintf("profile %q not found", c.ActiveProfile),
		}
	}
	return &p, nil
}

// DefaultConfig returns a config with a single production profile.
func DefaultConfig() *Config {
	return &Config{
		ActiveProfile: "production",
		Profiles: map[string]Profile{
			"production": {
				Name:           "Production",
				BaseURL:        defaultBaseURL,
				LogLevel:       defaultLogLevel,
				RequestTimeout: defaultRequestTimeout,
				MaxRetries:     defaultMaxRetries,
			},
		},
	}
}

// applyEnvOverrides applies UTEAMUP_* environment variable overrides to the active profile.
func applyEnvOverrides(cfg *Config) {
	if cfg.ActiveProfile == "" {
		return
	}
	p, ok := cfg.Profiles[cfg.ActiveProfile]
	if !ok {
		return
	}

	if v := os.Getenv("UTEAMUP_API_KEY"); v != "" {
		p.APIKey = v
	}
	if v := os.Getenv("UTEAMUP_SECRET"); v != "" {
		p.Secret = v
	}
	if v := os.Getenv("UTEAMUP_API_BASE_URL"); v != "" {
		p.BaseURL = v
	}
	if v := os.Getenv("UTEAMUP_LOG_LEVEL"); v != "" {
		p.LogLevel = v
	}

	cfg.Profiles[cfg.ActiveProfile] = p
}

// validate checks required fields.
func validate(cfg *Config) error {
	if len(cfg.Profiles) == 0 {
		return &clierrors.ConfigError{Message: "no profiles defined"}
	}
	if cfg.ActiveProfile == "" {
		return &clierrors.ConfigError{Field: "activeProfile", Message: "must not be empty"}
	}
	if _, ok := cfg.Profiles[cfg.ActiveProfile]; !ok {
		return &clierrors.ConfigError{
			Field:   "activeProfile",
			Message: fmt.Sprintf("profile %q not found in profiles", cfg.ActiveProfile),
		}
	}
	return nil
}

// RedactedSummary returns a display-safe summary of the active profile.
func (c *Config) RedactedSummary() string {
	p, err := c.ActiveProfileConfig()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	apiKey := "(not set)"
	if p.APIKey != "" {
		if len(p.APIKey) > 8 {
			apiKey = p.APIKey[:8] + "..."
		} else {
			apiKey = "***"
		}
	}

	secret := "(not set)"
	if p.Secret != "" {
		secret = "***"
	}

	exportDir := p.ExportDir
	if exportDir == "" {
		exportDir = "~/.uteamup/exports"
	}

	return fmt.Sprintf(`Active Profile: %s (%s)
  Base URL:        %s
  API Key:         %s
  Secret:          %s
  Log Level:       %s
  Request Timeout: %dms
  Max Retries:     %d
  Export JSON:     %v
  Export Dir:      %s`,
		c.ActiveProfile, p.Name,
		p.BaseURL, apiKey, secret,
		p.LogLevel, p.RequestTimeout, p.MaxRetries,
		p.ExportJSON, exportDir)
}
