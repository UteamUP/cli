package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogoutUsesCapturedOriginForHumanSessionsAndLeavesMachineKeysAlone(t *testing.T) {
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/api/auth/logout" || r.Header.Get("Authorization") != "Bearer captured" {
			t.Errorf("unexpected logout destination or credential")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	t.Setenv("UTEAMUP_API_BASE_URL", "https://changed-origin.example.invalid")
	for _, method := range []string{"login", "saml", "apikey"} {
		token := &TokenData{APIOrigin: server.URL, AccessToken: "captured", AuthMethod: method, ExpiresAt: time.Now().Add(time.Hour)}
		if err := RevokeSession(context.Background(), token, true); err != nil {
			t.Fatal(err)
		}
		if token.AccessToken != "captured" {
			t.Fatal("logout changed captured input")
		}
	}
	if calls != 2 {
		t.Fatalf("got %d calls", calls)
	}
}

func TestExpiredHumanLogoutRenewsBeforeRevokingWithoutLosingTheCapturedOrigin(t *testing.T) {
	renewed := "header." + base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d}`, time.Now().Add(time.Hour).Unix()))) + ".signature"
	steps := []string{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		steps = append(steps, r.URL.Path)
		switch r.URL.Path {
		case "/api/auth/refresh-token":
			fmt.Fprintf(w, `{"accessToken":%q,"refreshToken":"new-refresh"}`, renewed)
		case "/api/auth/logout":
			if r.Header.Get("Authorization") != "Bearer "+renewed {
				t.Error("logout did not use renewed credential")
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Error("unexpected request")
		}
	}))
	defer server.Close()
	token := &TokenData{APIOrigin: server.URL, AccessToken: "expired", RefreshToken: "old-refresh", AuthMethod: "saml", ExpiresAt: time.Now().Add(-time.Hour)}
	if err := RevokeSession(context.Background(), token, true); err != nil {
		t.Fatal(err)
	}
	if len(steps) != 2 || steps[0] != "/api/auth/refresh-token" || steps[1] != "/api/auth/logout" {
		t.Fatal(steps)
	}
	if token.AccessToken != "expired" || token.RefreshToken != "old-refresh" {
		t.Fatal("captured input changed")
	}
}

func TestLogoutRejectsUnsafeOriginsAndReportsBackendFailure(t *testing.T) {
	for _, origin := range []string{"http://remote.invalid", "https://user:secret@example.invalid", "https://example.invalid/path", ""} {
		token := &TokenData{APIOrigin: origin, AccessToken: "captured", AuthMethod: "login", ExpiresAt: time.Now().Add(time.Hour)}
		if RevokeSession(context.Background(), token, false) == nil {
			t.Fatal("unsafe origin accepted")
		}
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer server.Close()
	token := &TokenData{APIOrigin: server.URL, AccessToken: "captured", AuthMethod: "login", ExpiresAt: time.Now().Add(time.Hour)}
	if RevokeSession(context.Background(), token, true) == nil {
		t.Fatal("unverified remote logout reported success")
	}
}
