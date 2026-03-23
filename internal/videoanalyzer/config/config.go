// Package config provides configuration loading for the video analyzer tool.
// It reads YAML config files, applies environment variable overrides, and
// supports functional options for CLI flag overrides.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// GeminiConfig holds settings for the Gemini AI API.
type GeminiConfig struct {
	APIKey            string  `yaml:"api_key"`
	Model             string  `yaml:"model"`
	MaxOutputTokens   int     `yaml:"max_output_tokens"`
	Temperature       float64 `yaml:"temperature"`
	RequestsPerMinute int     `yaml:"requests_per_minute"`
	MaxRetries        int     `yaml:"max_retries"`
	TimeoutSeconds    int     `yaml:"timeout_seconds"`
	GoogleMapsAPIKey  string  `yaml:"google_maps_api_key"`
}

// VideoScanConfig holds video-specific scan settings.
type VideoScanConfig struct {
	VideoFolder  string `yaml:"video_folder"`
	OutputFolder string `yaml:"output_folder"`
	Recursive    bool   `yaml:"recursive"`
	MaxFileSizeMB int   `yaml:"max_file_size_mb"`
}

// ProcessingConfig holds settings for video processing behavior.
type ProcessingConfig struct {
	DryRun                      bool     `yaml:"dry_run"`
	MaxCost                     *float64 `yaml:"max_cost,omitempty"`
	GroupingSimilarityThreshold  float64  `yaml:"grouping_similarity_threshold"`
	ConfidenceThreshold         float64  `yaml:"confidence_threshold"`
	ProcessingTimeoutSec        int      `yaml:"processing_timeout_seconds"`
	PollIntervalSec             int      `yaml:"poll_interval_seconds"`
	TemporalDedupWindowSec      int      `yaml:"temporal_dedup_window_seconds"`
}

// AppConfig is the root video analyzer configuration.
type AppConfig struct {
	Gemini     GeminiConfig     `yaml:"gemini"`
	Scan       VideoScanConfig  `yaml:"scan"`
	Processing ProcessingConfig `yaml:"processing"`
}

// ConfigOption is a functional option for overriding config values.
type ConfigOption func(*AppConfig)

// DefaultConfig returns an AppConfig populated with sensible defaults.
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Gemini: GeminiConfig{
			Model:             "gemini-2.5-flash",
			MaxOutputTokens:   8192,
			Temperature:       0.1,
			RequestsPerMinute: 10,
			MaxRetries:        3,
			TimeoutSeconds:    60,
		},
		Scan: VideoScanConfig{
			OutputFolder:  "./Output",
			Recursive:     true,
			MaxFileSizeMB: 500,
		},
		Processing: ProcessingConfig{
			GroupingSimilarityThreshold: 0.75,
			ConfidenceThreshold:        0.5,
			ProcessingTimeoutSec:       600,
			PollIntervalSec:            5,
			TemporalDedupWindowSec:     30,
		},
	}
}

// WithFolderOverride overrides the video folder path.
func WithFolderOverride(path string) ConfigOption {
	return func(c *AppConfig) {
		if path != "" {
			c.Scan.VideoFolder = path
		}
	}
}

// WithOutputOverride overrides the output folder path.
func WithOutputOverride(path string) ConfigOption {
	return func(c *AppConfig) {
		if path != "" {
			c.Scan.OutputFolder = path
		}
	}
}

// WithDryRun sets the dry run mode.
func WithDryRun(dryRun bool) ConfigOption {
	return func(c *AppConfig) {
		c.Processing.DryRun = dryRun
	}
}

// WithAPIKey overrides the Gemini API key.
func WithAPIKey(key string) ConfigOption {
	return func(c *AppConfig) {
		if key != "" {
			c.Gemini.APIKey = key
		}
	}
}

// WithModel overrides the Gemini model name.
func WithModel(model string) ConfigOption {
	return func(c *AppConfig) {
		if model != "" {
			c.Gemini.Model = model
		}
	}
}

// WithGoogleMapsAPIKey overrides the Google Maps API key for reverse geocoding.
func WithGoogleMapsAPIKey(key string) ConfigOption {
	return func(c *AppConfig) {
		if key != "" {
			c.Gemini.GoogleMapsAPIKey = key
		}
	}
}

// WithMaxCost sets the maximum cost limit.
func WithMaxCost(maxCost *float64) ConfigOption {
	return func(c *AppConfig) {
		c.Processing.MaxCost = maxCost
	}
}

// LoadConfig reads a YAML config file, applies environment variable overrides,
// and then applies any functional option overrides. If the config file does not
// exist, defaults are used.
func LoadConfig(configPath string, opts ...ConfigOption) (*AppConfig, error) {
	cfg := DefaultConfig()

	// Read YAML file if it exists.
	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file %s: %w", configPath, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading config file %s: %w", configPath, err)
	}

	// Ensure defaults are applied for zero-value fields after YAML unmarshal.
	defaults := DefaultConfig()
	if cfg.Gemini.Model == "" {
		cfg.Gemini.Model = defaults.Gemini.Model
	}
	if cfg.Gemini.MaxOutputTokens == 0 {
		cfg.Gemini.MaxOutputTokens = defaults.Gemini.MaxOutputTokens
	}
	if cfg.Gemini.RequestsPerMinute == 0 {
		cfg.Gemini.RequestsPerMinute = defaults.Gemini.RequestsPerMinute
	}
	if cfg.Gemini.MaxRetries == 0 {
		cfg.Gemini.MaxRetries = defaults.Gemini.MaxRetries
	}
	if cfg.Gemini.TimeoutSeconds == 0 {
		cfg.Gemini.TimeoutSeconds = defaults.Gemini.TimeoutSeconds
	}
	if cfg.Scan.OutputFolder == "" {
		cfg.Scan.OutputFolder = defaults.Scan.OutputFolder
	}
	if cfg.Scan.MaxFileSizeMB == 0 {
		cfg.Scan.MaxFileSizeMB = defaults.Scan.MaxFileSizeMB
	}
	if cfg.Processing.ProcessingTimeoutSec == 0 {
		cfg.Processing.ProcessingTimeoutSec = defaults.Processing.ProcessingTimeoutSec
	}
	if cfg.Processing.PollIntervalSec == 0 {
		cfg.Processing.PollIntervalSec = defaults.Processing.PollIntervalSec
	}
	if cfg.Processing.TemporalDedupWindowSec == 0 {
		cfg.Processing.TemporalDedupWindowSec = defaults.Processing.TemporalDedupWindowSec
	}

	// Environment variable overrides (env vars take precedence over YAML).
	if v := os.Getenv("GEMINI_API_KEY"); v != "" {
		cfg.Gemini.APIKey = v
	}
	if v := os.Getenv("GEMINI_MODEL"); v != "" {
		cfg.Gemini.Model = v
	}
	if v := os.Getenv("GOOGLE_MAPS_API_KEY"); v != "" {
		cfg.Gemini.GoogleMapsAPIKey = v
	}

	// Apply functional option overrides (CLI flags take highest precedence).
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg, nil
}

// Validate checks the configuration for errors and returns a list of
// human-readable error messages.
func (c *AppConfig) Validate() []string {
	var errors []string

	// API key required unless dry run.
	if c.Gemini.APIKey == "" && !c.Processing.DryRun {
		errors = append(errors, "GEMINI_API_KEY is required (set in .env, environment, or config file)")
	}

	// Model name required.
	if c.Gemini.Model == "" {
		errors = append(errors, "Gemini model name is required")
	}

	// Output folder required.
	if c.Scan.OutputFolder == "" {
		errors = append(errors, "Output folder is required")
	}

	// Requests per minute must be positive.
	if c.Gemini.RequestsPerMinute < 1 {
		errors = append(errors, fmt.Sprintf("requests_per_minute must be > 0, got %d", c.Gemini.RequestsPerMinute))
	}

	// Max file size must be positive.
	if c.Scan.MaxFileSizeMB < 1 {
		errors = append(errors, fmt.Sprintf("max_file_size_mb must be > 0, got %d", c.Scan.MaxFileSizeMB))
	}

	// Similarity threshold must be 0-1.
	if c.Processing.GroupingSimilarityThreshold < 0 || c.Processing.GroupingSimilarityThreshold > 1 {
		errors = append(errors, fmt.Sprintf("grouping_similarity_threshold must be 0-1, got %v", c.Processing.GroupingSimilarityThreshold))
	}

	// Confidence threshold must be 0-1.
	if c.Processing.ConfidenceThreshold < 0 || c.Processing.ConfidenceThreshold > 1 {
		errors = append(errors, fmt.Sprintf("confidence_threshold must be 0-1, got %v", c.Processing.ConfidenceThreshold))
	}

	return errors
}
