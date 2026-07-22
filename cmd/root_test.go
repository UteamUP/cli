package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/config"
	"github.com/uteamup/cli/internal/logging"
	"github.com/uteamup/cli/internal/registry"
)

func TestAuthStatusIsAvailableWithoutAuthentication(t *testing.T) {
	if !commandsExemptFromAuth["auth"] {
		t.Fatal("auth parent command must bypass the root authentication gate")
	}
}

func TestSelectedProfileConfigUsesRequestedProfileAndEnvironment(t *testing.T) {
	t.Setenv("UTEAMUP_API_BASE_URL", "https://override.example.com")
	t.Setenv("UTEAMUP_LOG_LEVEL", "DEBUG")
	cfg := &config.Config{
		ActiveProfile: "production",
		Profiles: map[string]config.Profile{
			"production": {BaseURL: "https://api.uteamup.com", LogLevel: "INFO"},
			"windows":    {BaseURL: "https://localhost:5002", LogLevel: "WARN"},
		},
	}

	profile, name, err := selectedProfileConfig(cfg, "windows")
	if err != nil {
		t.Fatal(err)
	}
	if name != "windows" {
		t.Fatalf("selected profile = %q, want windows", name)
	}
	if profile.BaseURL != "https://override.example.com" {
		t.Fatalf("base URL = %q, want environment override", profile.BaseURL)
	}
	if profile.LogLevel != "DEBUG" {
		t.Fatalf("log level = %q, want DEBUG", profile.LogLevel)
	}
}

func TestNewDomainAPIClientHonorsRuntimeInsecureFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("UTEAMUP_API_BASE_URL", "")
	t.Setenv("UTEAMUP_LOG_LEVEL", "")

	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/asset" {
			t.Errorf("request path = %q, want /api/asset", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("authorization header = %q", request.Header.Get("Authorization"))
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	cfg := config.DefaultConfig()
	profile := cfg.Profiles[cfg.ActiveProfile]
	profile.BaseURL = server.URL
	cfg.Profiles[cfg.ActiveProfile] = profile
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	if err := auth.SaveToken(&auth.TokenData{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "11111111-1111-4111-8111-111111111111",
	}); err != nil {
		t.Fatal(err)
	}

	previousProfileName, previousInsecure, previousVerbose := profileName, insecure, verbose
	profileName, insecure, verbose = "", true, true
	t.Cleanup(func() {
		profileName, insecure, verbose = previousProfileName, previousInsecure, previousVerbose
	})

	apiClient, err := newDomainAPIClient(
		logging.New(logging.LevelError),
		&registry.ExportConfig{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := apiClient.CallREST(
		context.Background(),
		http.MethodGet,
		"/api/asset",
		nil,
		nil,
		"list",
	); err != nil {
		t.Fatalf("self-signed TLS request failed despite --insecure: %v", err)
	}
}
