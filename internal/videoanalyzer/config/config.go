// Package config provides configuration loading for the video analyzer tool.
// It reads YAML config files, applies environment variable overrides, and
// supports functional options for CLI flag overrides.
package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// VideoScanConfig holds video-specific scan settings.
type VideoScanConfig struct {
	VideoFolder   string `yaml:"video_folder"`
	OutputFolder  string `yaml:"output_folder"`
	Recursive     bool   `yaml:"recursive"`
	MaxFileSizeMB int    `yaml:"max_file_size_mb"`
}

// ProcessingConfig holds settings for video processing behavior.
type ProcessingConfig struct {
	DryRun                      bool    `yaml:"dry_run"`
	GroupingSimilarityThreshold float64 `yaml:"grouping_similarity_threshold"`
	ConfidenceThreshold         float64 `yaml:"confidence_threshold"`
	TemporalDedupWindowSec      int     `yaml:"temporal_dedup_window_seconds"`
}

// AppConfig is the root video analyzer configuration.
type AppConfig struct {
	Scan       VideoScanConfig  `yaml:"scan"`
	Processing ProcessingConfig `yaml:"processing"`
}

// Option overrides video-analysis configuration values.
type Option func(*AppConfig)

// DefaultConfig returns an AppConfig populated with sensible defaults.
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Scan: VideoScanConfig{
			OutputFolder:  "./Output",
			Recursive:     true,
			MaxFileSizeMB: 100,
		},
		Processing: ProcessingConfig{
			GroupingSimilarityThreshold: 0.75,
			ConfidenceThreshold:         0.5,
			TemporalDedupWindowSec:      30,
		},
	}
}

// WithFolderOverride overrides the video folder path.
func WithFolderOverride(path string) Option {
	return func(c *AppConfig) {
		if path != "" {
			c.Scan.VideoFolder = path
		}
	}
}

// WithOutputOverride overrides the output folder path.
func WithOutputOverride(path string) Option {
	return func(c *AppConfig) {
		if path != "" {
			c.Scan.OutputFolder = path
		}
	}
}

// WithDryRun sets the dry run mode.
func WithDryRun(dryRun bool) Option {
	return func(c *AppConfig) {
		c.Processing.DryRun = dryRun
	}
}

// LoadConfig reads a YAML config file, applies environment variable overrides,
// and then applies any functional option overrides. If the config file does not
// exist, defaults are used.
func LoadConfig(configPath string, opts ...Option) (*AppConfig, error) {
	cfg := DefaultConfig()

	// Read YAML file if it exists.
	data, err := os.ReadFile(configPath)
	if err == nil {
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		if err := decoder.Decode(cfg); err != nil {
			return nil, fmt.Errorf("parsing config file %s: %w", configPath, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading config file %s: %w", configPath, err)
	}

	// Ensure defaults are applied for zero-value fields after YAML unmarshal.
	defaults := DefaultConfig()
	if cfg.Scan.OutputFolder == "" {
		cfg.Scan.OutputFolder = defaults.Scan.OutputFolder
	}
	if cfg.Scan.MaxFileSizeMB == 0 {
		cfg.Scan.MaxFileSizeMB = defaults.Scan.MaxFileSizeMB
	}
	if cfg.Processing.TemporalDedupWindowSec == 0 {
		cfg.Processing.TemporalDedupWindowSec = defaults.Processing.TemporalDedupWindowSec
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

	// Output folder required.
	if c.Scan.OutputFolder == "" {
		errors = append(errors, "Output folder is required")
	}

	// Max file size must be positive.
	if c.Scan.MaxFileSizeMB < 1 || c.Scan.MaxFileSizeMB > 100 {
		errors = append(errors, fmt.Sprintf("max_file_size_mb must be between 1 and 100, got %d", c.Scan.MaxFileSizeMB))
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
