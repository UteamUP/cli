package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ActiveProfile != "production" {
		t.Errorf("expected active profile 'production', got %q", cfg.ActiveProfile)
	}

	p, ok := cfg.Profiles["production"]
	if !ok {
		t.Fatal("expected production profile to exist")
	}
	if p.BaseURL != "https://api.uteamup.com" {
		t.Errorf("unexpected base URL: %s", p.BaseURL)
	}
	if p.LogLevel != "INFO" {
		t.Errorf("unexpected log level: %s", p.LogLevel)
	}
	if p.RequestTimeout != 30000 {
		t.Errorf("unexpected timeout: %d", p.RequestTimeout)
	}
	if p.MaxRetries != 3 {
		t.Errorf("unexpected max retries: %d", p.MaxRetries)
	}
}

func TestValidateEmptyProfiles(t *testing.T) {
	cfg := &Config{
		ActiveProfile: "prod",
		Profiles:      map[string]Profile{},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected validation error for empty profiles")
	}
}

func TestValidateMissingActiveProfile(t *testing.T) {
	cfg := &Config{
		ActiveProfile: "nonexistent",
		Profiles: map[string]Profile{
			"prod": {Name: "Production", BaseURL: "https://api.uteamup.com"},
		},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected validation error for missing active profile")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use a temp directory instead of home
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".uteamup")
	os.MkdirAll(configDir, 0700)

	cfg := DefaultConfig()
	cfg.Profiles["production"] = Profile{
		Name:           "Test Production",
		APIKey:         "12345678901234567890123456789012",
		BaseURL:        "https://api.uteamup.com",
		LogLevel:       "DEBUG",
		RequestTimeout: 15000,
		MaxRetries:     5,
	}

	// Save to temp path
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Read back
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if loaded.ActiveProfile != "production" {
		t.Errorf("expected active profile 'production', got %q", loaded.ActiveProfile)
	}
	p := loaded.Profiles["production"]
	if p.LogLevel != "DEBUG" {
		t.Errorf("expected DEBUG log level, got %s", p.LogLevel)
	}
	if p.MaxRetries != 5 {
		t.Errorf("expected 5 retries, got %d", p.MaxRetries)
	}
}

func TestRedactedSummary(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Profiles["production"] = Profile{
		Name:           "Production",
		APIKey:         "12345678901234567890123456789012",
		Secret:         "supersecretvaluethatshouldneverbevisibleanywhereinoutput123456789012345",
		BaseURL:        "https://api.uteamup.com",
		LogLevel:       "INFO",
		RequestTimeout: 30000,
		MaxRetries:     3,
	}

	summary := cfg.RedactedSummary()

	if len(summary) == 0 {
		t.Fatal("summary should not be empty")
	}

	// API key should be partially shown
	if !contains(summary, "12345678...") {
		t.Error("API key should be partially redacted")
	}

	// Secret should be fully hidden
	if contains(summary, "supersecret") {
		t.Error("secret should not appear in summary")
	}
	if !contains(summary, "***") {
		t.Error("secret should be shown as ***")
	}
}

func TestEnvOverrides(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("UTEAMUP_API_KEY", "envkey12345678901234567890123456")
	t.Setenv("UTEAMUP_API_BASE_URL", "https://dev.uteamup.com")

	applyEnvOverrides(cfg)

	p := cfg.Profiles["production"]
	if p.APIKey != "envkey12345678901234567890123456" {
		t.Errorf("env override for API key not applied: got %s", p.APIKey)
	}
	if p.BaseURL != "https://dev.uteamup.com" {
		t.Errorf("env override for base URL not applied: got %s", p.BaseURL)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
