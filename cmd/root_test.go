package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
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
		APIOrigin:   server.URL,
		Profile:     cfg.ActiveProfile,
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

func renewTestJWT(exp int64) string {
	enc := base64.RawURLEncoding
	return enc.EncodeToString([]byte(`{"alg":"HS256"}`)) + "." +
		enc.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d}`, exp))) + "." + enc.EncodeToString([]byte("sig"))
}

func expiredSessionFor(origin string) *auth.TokenData {
	return &auth.TokenData{
		APIOrigin:    origin,
		AccessToken:  "old-access",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(-time.Hour),
		AuthMethod:   "login",
		Profile:      "prod",
	}
}

// An expired login session is renewed and saved, so the next command (and "auth status")
// sees a valid session instead of "Not authenticated".
func TestRenewSessionRenewsAndSavesAnExpiredLoginSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	previous := insecure
	insecure = true
	t.Cleanup(func() { insecure = previous })

	newAccess := renewTestJWT(time.Now().Add(2 * time.Hour).Unix())
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"accessToken":%q,"refreshToken":"new-refresh"}`, newAccess)
	}))
	defer server.Close()

	got := renewSession(expiredSessionFor(server.URL))
	if !got.IsValid() || got.AccessToken != newAccess {
		t.Fatalf("renewSession returned %+v, want the renewed session", got)
	}
	saved, err := auth.LoadToken()
	if err != nil || saved == nil || saved.RefreshToken != "new-refresh" || saved.AccessToken != newAccess {
		t.Fatalf("saved session = %+v, %v; want the rotated tokens", saved, err)
	}
}

// Refresh tokens rotate, so when two commands renew at once only one wins. The loser must use
// the session the winner saved rather than telling the user to sign in again.
func TestRenewSessionUsesASessionAnotherProcessRenewed(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	previous := insecure
	insecure = true
	t.Cleanup(func() { insecure = previous })

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Invalid refresh token.", http.StatusUnauthorized)
	}))
	defer server.Close()

	winner := expiredSessionFor(server.URL)
	winner.AccessToken = "winner-access"
	winner.ExpiresAt = time.Now().Add(time.Hour)
	if err := auth.SaveToken(winner); err != nil {
		t.Fatal(err)
	}

	got := renewSession(expiredSessionFor(server.URL))
	if got.AccessToken != "winner-access" || !got.IsValid() {
		t.Fatalf("renewSession returned %+v, want the session the other process saved", got)
	}
}

func TestRenewSessionLeavesAPIKeySessionsAlone(t *testing.T) {
	token := expiredSessionFor("https://127.0.0.1:1")
	token.AuthMethod = "apikey"
	if got := renewSession(token); got != token || got.IsValid() {
		t.Fatalf("an API-key session must not be renewed, got %+v", got)
	}
}
