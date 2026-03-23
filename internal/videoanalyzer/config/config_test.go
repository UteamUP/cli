package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Gemini defaults.
	if cfg.Gemini.Model != "gemini-2.5-flash" {
		t.Errorf("Model = %q, want %q", cfg.Gemini.Model, "gemini-2.5-flash")
	}
	if cfg.Gemini.MaxOutputTokens != 8192 {
		t.Errorf("MaxOutputTokens = %d, want %d", cfg.Gemini.MaxOutputTokens, 8192)
	}
	if cfg.Gemini.Temperature != 0.1 {
		t.Errorf("Temperature = %v, want %v", cfg.Gemini.Temperature, 0.1)
	}
	if cfg.Gemini.RequestsPerMinute != 10 {
		t.Errorf("RequestsPerMinute = %d, want %d", cfg.Gemini.RequestsPerMinute, 10)
	}
	if cfg.Gemini.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want %d", cfg.Gemini.MaxRetries, 3)
	}
	if cfg.Gemini.TimeoutSeconds != 60 {
		t.Errorf("TimeoutSeconds = %d, want %d", cfg.Gemini.TimeoutSeconds, 60)
	}

	// Scan defaults.
	if cfg.Scan.OutputFolder != "./Output" {
		t.Errorf("OutputFolder = %q, want %q", cfg.Scan.OutputFolder, "./Output")
	}
	if !cfg.Scan.Recursive {
		t.Error("Recursive = false, want true")
	}
	if cfg.Scan.MaxFileSizeMB != 500 {
		t.Errorf("MaxFileSizeMB = %d, want %d", cfg.Scan.MaxFileSizeMB, 500)
	}

	// Processing defaults.
	if cfg.Processing.GroupingSimilarityThreshold != 0.75 {
		t.Errorf("GroupingSimilarityThreshold = %v, want %v", cfg.Processing.GroupingSimilarityThreshold, 0.75)
	}
	if cfg.Processing.ConfidenceThreshold != 0.5 {
		t.Errorf("ConfidenceThreshold = %v, want %v", cfg.Processing.ConfidenceThreshold, 0.5)
	}
	if cfg.Processing.ProcessingTimeoutSec != 600 {
		t.Errorf("ProcessingTimeoutSec = %d, want %d", cfg.Processing.ProcessingTimeoutSec, 600)
	}
	if cfg.Processing.PollIntervalSec != 5 {
		t.Errorf("PollIntervalSec = %d, want %d", cfg.Processing.PollIntervalSec, 5)
	}
	if cfg.Processing.TemporalDedupWindowSec != 30 {
		t.Errorf("TemporalDedupWindowSec = %d, want %d", cfg.Processing.TemporalDedupWindowSec, 30)
	}
	if cfg.Processing.DryRun {
		t.Error("DryRun = true, want false")
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg.Gemini.Model != "gemini-2.5-flash" {
		t.Errorf("Model = %q, want default %q", cfg.Gemini.Model, "gemini-2.5-flash")
	}
}

func TestLoadConfig_YAMLFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	yamlContent := `
gemini:
  api_key: "test-key-123"
  model: "gemini-pro"
  max_output_tokens: 4096
scan:
  video_folder: "/tmp/videos"
  output_folder: "/tmp/output"
  max_file_size_mb: 200
processing:
  confidence_threshold: 0.8
  processing_timeout_seconds: 300
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Gemini.APIKey != "test-key-123" {
		t.Errorf("APIKey = %q, want %q", cfg.Gemini.APIKey, "test-key-123")
	}
	if cfg.Gemini.Model != "gemini-pro" {
		t.Errorf("Model = %q, want %q", cfg.Gemini.Model, "gemini-pro")
	}
	if cfg.Gemini.MaxOutputTokens != 4096 {
		t.Errorf("MaxOutputTokens = %d, want %d", cfg.Gemini.MaxOutputTokens, 4096)
	}
	if cfg.Scan.VideoFolder != "/tmp/videos" {
		t.Errorf("VideoFolder = %q, want %q", cfg.Scan.VideoFolder, "/tmp/videos")
	}
	if cfg.Scan.OutputFolder != "/tmp/output" {
		t.Errorf("OutputFolder = %q, want %q", cfg.Scan.OutputFolder, "/tmp/output")
	}
	if cfg.Scan.MaxFileSizeMB != 200 {
		t.Errorf("MaxFileSizeMB = %d, want %d", cfg.Scan.MaxFileSizeMB, 200)
	}
	if cfg.Processing.ConfidenceThreshold != 0.8 {
		t.Errorf("ConfidenceThreshold = %v, want %v", cfg.Processing.ConfidenceThreshold, 0.8)
	}
	if cfg.Processing.ProcessingTimeoutSec != 300 {
		t.Errorf("ProcessingTimeoutSec = %d, want %d", cfg.Processing.ProcessingTimeoutSec, 300)
	}
	// Defaults should still be applied for unset fields.
	if cfg.Gemini.RequestsPerMinute != 10 {
		t.Errorf("RequestsPerMinute = %d, want default %d", cfg.Gemini.RequestsPerMinute, 10)
	}
}

func TestLoadConfig_EnvVarOverrides(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "env-key-456")

	cfg, err := LoadConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Gemini.APIKey != "env-key-456" {
		t.Errorf("APIKey = %q, want %q", cfg.Gemini.APIKey, "env-key-456")
	}
}

func TestLoadConfig_FunctionalOptions(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.yaml",
		WithAPIKey("option-key-789"),
		WithModel("gemini-custom"),
	)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Gemini.APIKey != "option-key-789" {
		t.Errorf("APIKey = %q, want %q", cfg.Gemini.APIKey, "option-key-789")
	}
	if cfg.Gemini.Model != "gemini-custom" {
		t.Errorf("Model = %q, want %q", cfg.Gemini.Model, "gemini-custom")
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Gemini.APIKey = "valid-key"

	errors := cfg.Validate()
	if len(errors) != 0 {
		t.Errorf("Validate() returned %d errors, want 0: %v", len(errors), errors)
	}
}

func TestValidate_MissingAPIKey(t *testing.T) {
	cfg := DefaultConfig()
	// APIKey is empty by default.

	errors := cfg.Validate()
	found := false
	for _, e := range errors {
		if e == "GEMINI_API_KEY is required (set in .env, environment, or config file)" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Validate() did not return API key error, got: %v", errors)
	}
}

func TestValidate_InvalidThresholds(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Gemini.APIKey = "valid-key"
	cfg.Processing.GroupingSimilarityThreshold = 1.5 // Invalid: > 1.0

	errors := cfg.Validate()
	found := false
	for _, e := range errors {
		if len(e) > 0 {
			// Check for threshold error message.
			if e == "grouping_similarity_threshold must be 0-1, got 1.5" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("Validate() did not return threshold error, got: %v", errors)
	}
}

func TestValidate_InvalidMaxFileSize(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Gemini.APIKey = "valid-key"
	cfg.Scan.MaxFileSizeMB = 0 // Invalid: must be > 0

	errors := cfg.Validate()
	found := false
	for _, e := range errors {
		if e == "max_file_size_mb must be > 0, got 0" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Validate() did not return max file size error, got: %v", errors)
	}
}
