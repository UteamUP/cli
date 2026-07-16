package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigUsesBackendUploadLimit(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Scan.MaxFileSizeMB != 100 {
		t.Fatalf("MaxFileSizeMB = %d, want 100", cfg.Scan.MaxFileSizeMB)
	}
	if cfg.Processing.TemporalDedupWindowSec != 30 {
		t.Fatalf("TemporalDedupWindowSec = %d, want 30", cfg.Processing.TemporalDedupWindowSec)
	}
}

func TestLoadConfigRejectsLegacyProviderCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte("gemini:\n  api_key: secret\n  model: direct-model\n")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(path)
	if err == nil || !strings.Contains(err.Error(), "field gemini not found") {
		t.Fatalf("expected legacy provider config rejection, got %v", err)
	}
}

func TestValidateRejectsOversizedVideoLimit(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Scan.MaxFileSizeMB = 101
	if errors := cfg.Validate(); len(errors) == 0 {
		t.Fatal("expected validation error for a file limit above 100 MB")
	}
}

func TestLoadConfigAppliesLocalProcessingOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte(`scan:
  output_folder: ./custom
  max_file_size_mb: 80
processing:
  grouping_similarity_threshold: 0.8
  confidence_threshold: 0.6
  temporal_dedup_window_seconds: 45
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Scan.MaxFileSizeMB != 80 || cfg.Processing.TemporalDedupWindowSec != 45 {
		t.Fatalf("unexpected loaded config: %+v", cfg)
	}
}
