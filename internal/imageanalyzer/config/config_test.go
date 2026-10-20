package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uteamup/cli/internal/security"
)

func TestDefaultConfigUsesBackendPhotoLimit(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Scan.MaxFileSizeMB != 15 {
		t.Fatalf("MaxFileSizeMB = %d, want 15", cfg.Scan.MaxFileSizeMB)
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

func TestValidateRejectsOversizedPhotoLimit(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Scan.ImageFolder = t.TempDir()
	cfg.Scan.OutputFolder = filepath.Join(t.TempDir(), "output")
	cfg.Scan.RenamedImagesFolder = filepath.Join(t.TempDir(), "renamed")
	cfg.Scan.MaxFileSizeMB = 16
	if errors := cfg.Validate(); len(errors) == 0 {
		t.Fatal("expected validation error for a file limit above 15 MB")
	}
}

func TestAmbientConfigCannotSelectAnUnownedCheckpoint(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	victim := filepath.Join(t.TempDir(), "unrelated.json")
	if err := os.WriteFile(victim, []byte(`{"keep":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(fmt.Sprintf("processing:\n  checkpoint_file: %q\n", victim)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("ambient arbitrary deletion path accepted")
	}
	data, _ := os.ReadFile(victim)
	if string(data) != `{"keep":true}` {
		t.Fatal("unowned JSON changed")
	}
}
