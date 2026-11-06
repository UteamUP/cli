package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	clierrors "github.com/uteamup/cli/internal/errors"
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

func TestNormalizeTrustedBlobOriginsRequiresExactHTTPSOrigins(t *testing.T) {
	got, err := NormalizeTrustedBlobOrigins([]string{
		" https://Blob.Internal.Example:443/ ",
		"https://blob.internal.example",
		"https://storage.example:8443",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "https://blob.internal.example" || got[1] != "https://storage.example:8443" {
		t.Fatalf("normalized origins = %#v", got)
	}

	for _, invalid := range []string{
		"http://blob.internal.example",
		"https://*.internal.example",
		"https://blob.internal.example/path",
		"https://user:pass@blob.internal.example",
		"https://blob.internal.example?token=value",
	} {
		if _, err := NormalizeTrustedBlobOrigins([]string{invalid}); err == nil {
			t.Errorf("accepted non-origin trusted Blob value %q", invalid)
		}
	}
}

func TestValidateCanonicalizesTrustedBlobOriginsPerProfile(t *testing.T) {
	cfg := &Config{
		ActiveProfile: "production",
		Profiles: map[string]Profile{
			"production": {
				BaseURL:            "https://api.uteamup.com",
				TrustedBlobOrigins: []string{"https://Storage.Example:443/"},
			},
		},
	}
	if err := validate(cfg); err != nil {
		t.Fatal(err)
	}
	if got := cfg.Profiles["production"].TrustedBlobOrigins; len(got) != 1 || got[0] != "https://storage.example" {
		t.Fatalf("trusted Blob origins = %#v", got)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use a temp directory instead of home
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".uteamup")
	os.MkdirAll(configDir, 0700)

	cfg := DefaultConfig()
	cfg.Profiles["production"] = Profile{
		Name:               "Test Production",
		APIKey:             "12345678901234567890123456789012",
		BaseURL:            "https://api.uteamup.com",
		TrustedBlobOrigins: []string{"https://blob.internal.example:8443"},
		LogLevel:           "DEBUG",
		RequestTimeout:     15000,
		MaxRetries:         5,
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
	if len(p.TrustedBlobOrigins) != 1 || p.TrustedBlobOrigins[0] != "https://blob.internal.example:8443" {
		t.Errorf("unexpected trusted Blob origins: %#v", p.TrustedBlobOrigins)
	}
}

func TestConfigPersistenceUsesPrivatePermissionsAndRejectsPublicFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := Save(DefaultConfig()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".uteamup", "config.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("config permissions = %04o, want 0600", info.Mode().Perm())
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(); err == nil {
			t.Fatal("public config with credentials accepted")
		}
	}
}

func TestLoadMissingConfigReturnsSetupGuidance(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	_, err := Load()
	var configErr *clierrors.ConfigError
	if !errors.As(err, &configErr) || !strings.Contains(err.Error(), "uteamup config init") {
		t.Fatalf("Load error = %v, want the missing-config setup error", err)
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

	// API credentials should be fully hidden.
	if contains(summary, "12345678") {
		t.Error("API key should not appear in summary")
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

func TestLoadRemovesDeprecatedDirectProviderSettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	configPath, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatal(err)
	}
	legacy := `{
  "activeProfile": "production",
  "profiles": {
    "production": {
      "name": "Production",
      "baseUrl": "https://api.uteamup.com",
      "logLevel": "INFO",
      "requestTimeout": 30000,
      "maxRetries": 3,
      "geminiApiKey": "must-be-removed",
      "geminiModel": "must-be-removed",
      "googleMapsApiKey": "must-be-removed"
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	sanitized, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"geminiApiKey", "geminiModel", "googleMapsApiKey", "must-be-removed"} {
		if contains(string(sanitized), forbidden) {
			t.Fatalf("deprecated provider setting %q remained in config", forbidden)
		}
	}
}

func savedProductionProfile() Profile {
	return Profile{
		Name:           "Production",
		APIKey:         "diskkey1234567890123456789012345",
		Secret:         "disk-secret",
		BaseURL:        "https://api.uteamup.com",
		LogLevel:       "INFO",
		RequestTimeout: 15000,
		MaxRetries:     4,
		ExportDir:      "/exports",
		TenantGUID:     "11111111-1111-1111-1111-111111111111",
	}
}

// useConfigHome points the config at a temp home, clears every UTEAMUP_* override the
// developer's shell may export, and writes the given profiles as the saved config.
func useConfigHome(t *testing.T, active string, profiles map[string]Profile) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	for _, override := range envOverrides {
		t.Setenv(override.name, "")
	}
	if err := Save(&Config{ActiveProfile: active, Profiles: profiles}); err != nil {
		t.Fatal(err)
	}
	return home
}

func readSavedConfig(t *testing.T, home string) Config {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".uteamup", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	return saved
}

func TestSaveKeepsEnvOverridesOutOfConfigFile(t *testing.T) {
	home := useConfigHome(t, "production", map[string]Profile{"production": savedProductionProfile()})
	t.Setenv("UTEAMUP_API_KEY", "envkey12345678901234567890123456")
	t.Setenv("UTEAMUP_SECRET", "env-secret")
	t.Setenv("UTEAMUP_API_BASE_URL", "https://devback.uteamup.com")
	t.Setenv("UTEAMUP_LOG_LEVEL", "DEBUG")
	t.Setenv("UTEAMUP_TENANT_GUID", "22222222-2222-2222-2222-222222222222")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	inMemory := cfg.Profiles["production"]
	if inMemory.BaseURL != "https://devback.uteamup.com" || inMemory.LogLevel != "DEBUG" ||
		inMemory.TenantGUID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("env overrides not applied in memory: %+v", inMemory)
	}

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	want := savedProductionProfile()
	if got := readSavedConfig(t, home).Profiles["production"]; !reflect.DeepEqual(got, want) {
		t.Errorf("env overrides written to disk:\n got  %+v\n want %+v", got, want)
	}
}

func TestSaveWritesValuesSetExplicitlyWhileEnvOverridesAreActive(t *testing.T) {
	home := useConfigHome(t, "production", map[string]Profile{"production": savedProductionProfile()})
	t.Setenv("UTEAMUP_API_BASE_URL", "https://devback.uteamup.com")
	t.Setenv("UTEAMUP_LOG_LEVEL", "DEBUG")
	t.Setenv("UTEAMUP_TENANT_GUID", "22222222-2222-2222-2222-222222222222")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	p := cfg.Profiles["production"]
	p.BaseURL = "https://localhost:5002"
	p.LogLevel = "WARN"
	p.TenantGUID = "33333333-3333-3333-3333-333333333333"
	cfg.Profiles["production"] = p
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	got := readSavedConfig(t, home).Profiles["production"]
	if got.BaseURL != "https://localhost:5002" || got.LogLevel != "WARN" ||
		got.TenantGUID != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("explicitly set values not saved: %+v", got)
	}
	if got.APIKey != savedProductionProfile().APIKey || got.MaxRetries != 4 {
		t.Errorf("untouched file values changed: %+v", got)
	}
}

func TestSaveAfterProfileSwitchPreservesEveryProfilesFileValues(t *testing.T) {
	dev := Profile{Name: "Dev", BaseURL: "https://devback.uteamup.com", LogLevel: "DEBUG", RequestTimeout: 30000, MaxRetries: 3}
	home := useConfigHome(t, "production", map[string]Profile{"production": savedProductionProfile(), "dev": dev})
	t.Setenv("UTEAMUP_API_BASE_URL", "https://override.example.com")
	t.Setenv("UTEAMUP_LOG_LEVEL", "TRACE")
	t.Setenv("UTEAMUP_TENANT_GUID", "22222222-2222-2222-2222-222222222222")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetActiveProfile("dev"); err != nil {
		t.Fatal(err)
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	saved := readSavedConfig(t, home)
	if saved.ActiveProfile != "dev" {
		t.Errorf("active profile = %q, want dev", saved.ActiveProfile)
	}
	if got := saved.Profiles["production"]; !reflect.DeepEqual(got, savedProductionProfile()) {
		t.Errorf("production profile changed:\n got  %+v\n want %+v", got, savedProductionProfile())
	}
	if got := saved.Profiles["dev"]; !reflect.DeepEqual(got, dev) {
		t.Errorf("dev profile changed:\n got  %+v\n want %+v", got, dev)
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
