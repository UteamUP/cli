package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uteamup/cli/internal/security"
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

func TestLoadConfigRejectsOversizedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := strings.Repeat("x", int(security.MaximumConfigFileBytes)+1)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("oversized configuration was accepted")
	}
}

func TestLoadConfigRejectsSymlink(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "target.yaml")
	if err := os.WriteFile(target, []byte("scan:\n  max_file_size_mb: 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(directory, "config.yaml")
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	if _, err := LoadConfig(link); err == nil {
		t.Fatal("symlink configuration was accepted")
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

func TestLoadConfigPreservesExplicitNonrecursiveScan(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("scan:\n  recursive: false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Scan.Recursive {
		t.Fatal("explicit recursion boundary was overridden")
	}
	if !DefaultConfig().Scan.Recursive {
		t.Fatal("default recursion changed")
	}
}
