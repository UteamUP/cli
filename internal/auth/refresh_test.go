package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func testJWT(t *testing.T, exp int64) string {
	t.Helper()
	enc := base64.RawURLEncoding
	payload, err := json.Marshal(map[string]any{"sub": "user-1", "exp": exp})
	if err != nil {
		t.Fatal(err)
	}
	return enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." +
		enc.EncodeToString(payload) + "." + enc.EncodeToString([]byte("sig"))
}

func expiredLoginSession() *TokenData {
	return &TokenData{
		AccessToken:  "old-access",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(-time.Hour),
		AuthMethod:   "login",
		Email:        "user@example.com",
		Profile:      "prod",
		TenantGUID:   "tenant-1",
	}
}

// The server rotates refresh tokens, so a renewal must store the new refresh token too: keeping
// the old one would make the next renewal fail and send the user back to "uteamup login".
func TestRefreshSessionStoresRotatedTokensAndExpiry(t *testing.T) {
	exp := time.Now().Add(2 * time.Hour).Unix()
	newAccess := testJWT(t, exp)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/auth/refresh-token" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken != "old-refresh" {
			t.Errorf("refresh body = %+v (%v), want the stored refresh token", body, err)
		}
		fmt.Fprintf(w, `{"accessToken":%q,"refreshToken":"new-refresh"}`, newAccess)
	}))
	defer server.Close()

	token := expiredLoginSession()
	if err := NewClient(server.URL, true, nil).RefreshSession(token); err != nil {
		t.Fatalf("RefreshSession: %v", err)
	}

	if token.AccessToken != newAccess || token.RefreshToken != "new-refresh" {
		t.Errorf("tokens not replaced: access=%q refresh=%q", token.AccessToken, token.RefreshToken)
	}
	if token.ExpiresAt.Unix() != exp {
		t.Errorf("expiry = %v, want the JWT exp %v", token.ExpiresAt.Unix(), exp)
	}
	if !token.IsValid() {
		t.Error("renewed session should be valid")
	}
	if token.Email != "user@example.com" || token.Profile != "prod" || token.TenantGUID != "tenant-1" {
		t.Errorf("session identity changed: %+v", token)
	}
}

// A refresh token the server no longer accepts must leave the session as it was, so the
// command reports "Not authenticated" instead of saving a broken session.
func TestRefreshSessionRejectedLeavesSessionUntouched(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Invalid refresh token.", http.StatusUnauthorized)
	}))
	defer server.Close()

	token := expiredLoginSession()
	before := *token
	if err := NewClient(server.URL, true, nil).RefreshSession(token); err == nil {
		t.Fatal("expected an error for a rejected refresh token")
	}
	if *token != before {
		t.Errorf("session changed after a failed renewal: %+v", token)
	}
}

// An API-key session has its own credential lifecycle; it must never be sent to the
// email/password refresh endpoint.
func TestRefreshSessionSkipsSessionsItCannotRenew(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits.Add(1)
	}))
	defer server.Close()
	client := NewClient(server.URL, true, nil)

	apiKey := expiredLoginSession()
	apiKey.AuthMethod = "apikey"
	noRefresh := expiredLoginSession()
	noRefresh.RefreshToken = ""

	for name, token := range map[string]*TokenData{"apikey": apiKey, "no refresh token": noRefresh, "nil": nil} {
		if err := client.RefreshSession(token); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	if hits.Load() != 0 {
		t.Errorf("refresh endpoint called %d times, want 0", hits.Load())
	}
}

func TestRefreshSessionRejectsAnAccessTokenWithoutExpiry(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"accessToken":"not-a-jwt","refreshToken":"new-refresh"}`)
	}))
	defer server.Close()

	token := expiredLoginSession()
	if err := NewClient(server.URL, true, nil).RefreshSession(token); err == nil {
		t.Fatal("expected an error for an access token without a readable expiry")
	}
	if token.AccessToken != "old-access" || token.RefreshToken != "old-refresh" {
		t.Errorf("session changed after a failed renewal: %+v", token)
	}
}

// Renewal rewrites the session on ordinary commands, so the save must not leave a temp file
// behind or lose the 0600 permission.
func TestSaveTokenReplacesTheSessionAtomically(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	first := expiredLoginSession()
	if err := SaveToken(first); err != nil {
		t.Fatal(err)
	}
	renewed := expiredLoginSession()
	renewed.AccessToken = "renewed-access"
	if err := SaveToken(renewed); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadToken()
	if err != nil || loaded == nil || loaded.AccessToken != "renewed-access" {
		t.Fatalf("LoadToken = %+v, %v; want the renewed session", loaded, err)
	}
	path, _ := tokenPath()
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temp file left behind: %v", err)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0600 {
		t.Errorf("token file mode = %v (%v), want 0600", info.Mode().Perm(), err)
	}
}

func TestAccessTokenExpiry(t *testing.T) {
	exp := time.Date(2026, 10, 5, 21, 8, 41, 0, time.UTC).Unix()
	got, err := accessTokenExpiry(testJWT(t, exp))
	if err != nil || got.Unix() != exp {
		t.Fatalf("accessTokenExpiry = %v, %v; want %v", got, err, exp)
	}
	for _, bad := range []string{"", "a.b", "a.!!!.c", "a." + base64.RawURLEncoding.EncodeToString([]byte(`{}`)) + ".c"} {
		if _, err := accessTokenExpiry(bad); err == nil {
			t.Errorf("accessTokenExpiry(%q): expected an error", bad)
		}
	}
}
