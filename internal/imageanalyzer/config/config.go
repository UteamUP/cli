// Package config provides configuration loading for the image analyzer tool.
// It reads YAML config files, applies environment variable overrides, and
// supports functional options for CLI flag overrides.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ScanConfig holds settings for image scanning and discovery.
type ScanConfig struct {
	ImageFolder         string   `yaml:"image_folder"`
	OutputFolder        string   `yaml:"output_folder"`
	RenamedImagesFolder string   `yaml:"renamed_images_folder"`
	Recursive           bool     `yaml:"recursive"`
	SupportedFormats    []string `yaml:"supported_formats"`
	MaxImageDimension   int      `yaml:"max_image_dimension"`
	MaxFileSizeMB       int      `yaml:"max_file_size_mb"`
}

// ProcessingConfig holds settings for image processing behavior.
type ProcessingConfig struct {
	DryRun                      bool    `yaml:"dry_run"`
	RenameImages                bool    `yaml:"rename_images"`
	RenamePattern               string  `yaml:"rename_pattern"`
	GroupingSimilarityThreshold float64 `yaml:"grouping_similarity_threshold"`
	ConfidenceThreshold         float64 `yaml:"confidence_threshold"`
	CheckpointFile              string  `yaml:"checkpoint_file"`
}

// AppConfig is the top-level configuration for the image analyzer.
type AppConfig struct {
	Scan       ScanConfig       `yaml:"scan"`
	Processing ProcessingConfig `yaml:"processing"`
}

// defaultCheckpointPath returns the checkpoint file path inside ~/.uteamup/.
func defaultCheckpointPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".checkpoint.json"
	}
	dir := filepath.Join(home, ".uteamup")
	_ = os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "image-checkpoint.json")
}

// DefaultConfig returns an AppConfig populated with all default values.
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Scan: ScanConfig{
			ImageFolder:         "./Images/Original",
			OutputFolder:        "./Output",
			RenamedImagesFolder: "./Images/Updated",
			Recursive:           true,
			SupportedFormats:    []string{".jpg", ".jpeg", ".png", ".webp", ".heic", ".heif", ".tiff", ".bmp"},
			MaxImageDimension:   2048,
			MaxFileSizeMB:       15,
		},
		Processing: ProcessingConfig{
			RenameImages:                true,
			RenamePattern:               "{entity_type}_{name}_{seq}_{date}.{ext}",
			GroupingSimilarityThreshold: 0.75,
			ConfidenceThreshold:         0.5,
			CheckpointFile:              defaultCheckpointPath(),
		},
	}
}

// Option overrides image-analysis configuration values.
type Option func(*AppConfig)

// WithFolderOverride overrides the image folder path.
func WithFolderOverride(folder string) Option {
	return func(c *AppConfig) {
		if folder != "" {
			c.Scan.ImageFolder = folder
		}
	}
}

// WithOutputOverride overrides the output folder path.
func WithOutputOverride(output string) Option {
	return func(c *AppConfig) {
		if output != "" {
			c.Scan.OutputFolder = output
		}
	}
}

// WithDryRun sets the dry run mode.
func WithDryRun(dryRun bool) Option {
	return func(c *AppConfig) {
		c.Processing.DryRun = dryRun
	}
}

// WithNoRename disables image renaming when true.
func WithNoRename(noRename bool) Option {
	return func(c *AppConfig) {
		if noRename {
			c.Processing.RenameImages = false
		}
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
	if cfg.Scan.ImageFolder == "" {
		cfg.Scan.ImageFolder = defaults.Scan.ImageFolder
	}
	if cfg.Scan.OutputFolder == "" {
		cfg.Scan.OutputFolder = defaults.Scan.OutputFolder
	}
	if cfg.Scan.RenamedImagesFolder == "" {
		cfg.Scan.RenamedImagesFolder = defaults.Scan.RenamedImagesFolder
	}
	if len(cfg.Scan.SupportedFormats) == 0 {
		cfg.Scan.SupportedFormats = defaults.Scan.SupportedFormats
	}
	if cfg.Scan.MaxImageDimension == 0 {
		cfg.Scan.MaxImageDimension = defaults.Scan.MaxImageDimension
	}
	if cfg.Scan.MaxFileSizeMB == 0 {
		cfg.Scan.MaxFileSizeMB = defaults.Scan.MaxFileSizeMB
	}
	if cfg.Processing.RenamePattern == "" {
		cfg.Processing.RenamePattern = defaults.Processing.RenamePattern
	}
	if cfg.Processing.CheckpointFile == "" {
		cfg.Processing.CheckpointFile = defaults.Processing.CheckpointFile
	}

	// Environment variable overrides (env vars take precedence over YAML).
	if v := os.Getenv("IMAGE_FOLDER"); v != "" {
		cfg.Scan.ImageFolder = v
	}
	if v := os.Getenv("OUTPUT_FOLDER"); v != "" {
		cfg.Scan.OutputFolder = v
	}
	if v := os.Getenv("RENAMED_IMAGES_FOLDER"); v != "" {
		cfg.Scan.RenamedImagesFolder = v
	}
	// Apply functional option overrides (CLI flags take highest precedence).
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg, nil
}

// Validate checks the configuration for errors and returns a list of
// human-readable error messages. It also creates output directories if they
// do not exist.
func (c *AppConfig) Validate() []string {
	var errors []string

	// Image folder must exist.
	imageFolder, _ := filepath.Abs(c.Scan.ImageFolder)
	if info, err := os.Stat(imageFolder); err != nil || !info.IsDir() {
		errors = append(errors, fmt.Sprintf("Image folder does not exist: %s", imageFolder))
	}

	// Create output folders if they don't exist.
	outputFolder, _ := filepath.Abs(c.Scan.OutputFolder)
	if err := os.MkdirAll(outputFolder, 0o700); err != nil {
		errors = append(errors, fmt.Sprintf("Cannot create output folder %s: %v", outputFolder, err))
	}

	renamedFolder, _ := filepath.Abs(c.Scan.RenamedImagesFolder)
	if err := os.MkdirAll(renamedFolder, 0o700); err != nil {
		errors = append(errors, fmt.Sprintf("Cannot create renamed images folder %s: %v", renamedFolder, err))
	}

	if c.Scan.MaxFileSizeMB < 1 || c.Scan.MaxFileSizeMB > 15 {
		errors = append(errors, fmt.Sprintf("max_file_size_mb must be between 1 and 15, got %d", c.Scan.MaxFileSizeMB))
	}
	if c.Scan.MaxImageDimension < 256 || c.Scan.MaxImageDimension > 4096 {
		errors = append(errors, fmt.Sprintf("max_image_dimension must be between 256 and 4096, got %d", c.Scan.MaxImageDimension))
	}
	if c.Processing.GroupingSimilarityThreshold < 0 || c.Processing.GroupingSimilarityThreshold > 1 {
		errors = append(errors, fmt.Sprintf("grouping_similarity_threshold must be 0-1, got %v", c.Processing.GroupingSimilarityThreshold))
	}
	if c.Processing.ConfidenceThreshold < 0 || c.Processing.ConfidenceThreshold > 1 {
		errors = append(errors, fmt.Sprintf("confidence_threshold must be 0-1, got %v", c.Processing.ConfidenceThreshold))
	}

	return errors
}
