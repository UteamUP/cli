package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/logging"
)

func TestSamlCallbackRejectsUnboundAndDuplicateParameters(t *testing.T) {
	for _, query := range []string{
		"code=one&state=wrong", "code=one&state=secret&state=secret",
		"code=one&code=two&state=secret", "code=one&error=cancel&state=secret",
		"code=one&state=secret&accessToken=sensitive", "state=secret", "error=&state=secret",
	} {
		t.Run(query, func(t *testing.T) {
			callback := make(chan samlCallback, 1)
			handler := samlCallbackHandler("secret", "127.0.0.1:54321", callback)
			request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:54321"+samlCallbackPath+"?"+query, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest || len(callback) != 0 {
				t.Fatalf("unbound callback accepted: %d", response.Code)
			}
		})
	}
}

func TestSamlCallbackRequiresExactMethodHostAndPathAndConsumesOnce(t *testing.T) {
	callback := make(chan samlCallback, 1)
	handler := samlCallbackHandler("secret", "127.0.0.1:54321", callback)
	for _, invalid := range []struct{ method, target string }{
		{http.MethodPost, "http://127.0.0.1:54321" + samlCallbackPath},
		{http.MethodGet, "http://evil.example:54321" + samlCallbackPath},
		{http.MethodGet, "http://127.0.0.1:54321/wrong"},
	} {
		request := httptest.NewRequest(invalid.method, invalid.target+"?code=one&state=secret", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || len(callback) != 0 {
			t.Fatal("incorrect callback boundary accepted")
		}
	}
	for _, status := range []int{http.StatusOK, http.StatusGone} {
		request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:54321"+samlCallbackPath+"?code=one&state=secret", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != status || response.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("callback status %d, expected %d", response.Code, status)
		}
	}
	if len(callback) != 1 {
		t.Fatal("callback was consumed more than once")
	}
}

func TestSamlLoginKeepsPkceAndTenantBoundThroughLinkAndMfa(t *testing.T) {
	var mutex sync.Mutex
	var start map[string]string
	var exchanges []samlExchangeRequest
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/auth/saml/start":
			_ = json.NewDecoder(r.Body).Decode(&start)
			if start["companyCode"] != "iteggs" || start["clientId"] != "cli" || start["codeChallengeMethod"] != "S256" {
				t.Error("incorrect SAML start contract")
			}
			_ = json.NewEncoder(w).Encode(samlStartResponse{AuthenticationURL: "https://login.microsoftonline.com/authorize", ExpiresAt: time.Now().Add(time.Minute)})
		case "/api/auth/saml/exchange":
			var request samlExchangeRequest
			_ = json.NewDecoder(r.Body).Decode(&request)
			exchanges = append(exchanges, request)
			if CodeChallenge(request.CodeVerifier) != start["codeChallenge"] || request.ClientID != "cli" {
				t.Error("PKCE or client binding was lost")
			}
			response := samlExchangeResponse{ExpiresAt: time.Now().Add(time.Minute), Email: "gisli@iteggs.com"}
			switch len(exchanges) {
			case 1:
				if request.Code != "opaque-code" || request.ContinuationToken != "" {
					t.Error("first exchange did not use callback code")
				}
				response.Status, response.ContinuationToken = "link_required", "link-proof"
			case 2:
				if request.Code != "" || request.ContinuationToken != "link-proof" || request.VerificationCode != "123456" {
					t.Error("link proof was not bound to continuation")
				}
				response.Status, response.ContinuationToken = "mfa_required", "mfa-proof"
			case 3:
				if request.Code != "" || request.ContinuationToken != "mfa-proof" || request.MfaCode != "654321" {
					t.Error("MFA proof was not bound to continuation")
				}
				response.Status, response.TenantGUID = "authenticated", "11111111-1111-1111-1111-111111111111"
				response.Profile = &LoginResponse{AccessToken: "session", RefreshToken: "refresh", Email: "gisli@iteggs.com", TokenExpiry: time.Now().Add(time.Hour).Format(time.RFC3339)}
			default:
				t.Error("unexpected exchange replay")
			}
			_ = json.NewEncoder(w).Encode(response)
		case "/api/tenant/my-tenants":
			if r.Header.Get("Authorization") != "Bearer session" {
				t.Error("tenant membership was not checked with issued session")
			}
			_ = json.NewEncoder(w).Encode([]TenantInfo{
				{ID: 9, GUID: "other", Name: "Other"},
				{ID: 12, GUID: "11111111-1111-1111-1111-111111111111", Name: "Iteggs"},
			})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	client := NewClient(server.URL, true, logging.Default())
	interaction := samlInteraction{
		launch: func(_ context.Context, raw string) error {
			if raw != "https://login.microsoftonline.com/authorize" {
				t.Error("unexpected browser target")
			}
			mutex.Lock()
			redirectURI, state := start["redirectUri"], start["state"]
			mutex.Unlock()
			redirect, err := url.Parse(redirectURI)
			if err != nil || redirect.Hostname() != "127.0.0.1" || redirect.Path != samlCallbackPath {
				return fmt.Errorf("unsafe callback URL")
			}
			redirect.RawQuery = url.Values{"code": {"opaque-code"}, "state": {state}}.Encode()
			response, err := http.Get(redirect.String())
			if err == nil {
				_ = response.Body.Close()
			}
			return err
		},
		prompt: func(status, email string) (string, error) {
			if email != "gisli@iteggs.com" {
				t.Error("ownership prompt omitted account")
			}
			if status == "link_required" {
				return "123456", nil
			}
			return "654321", nil
		},
	}
	token, err := client.loginWithSaml(context.Background(), " iteggs ", interaction)
	if err != nil {
		t.Fatal(err)
	}
	if token.AuthMethod != "saml" || token.TenantName != "Iteggs" || token.TenantID != 12 || token.Email != "gisli@iteggs.com" {
		t.Fatalf("SAML session accepted in wrong scope: %+v", token)
	}
	token.Profile = "dev"
	if err := token.ValidateBinding(server.URL, "dev"); err != nil {
		t.Fatal(err)
	}
	if token.ValidateBinding("https://api.uteamup.com", "dev") == nil || token.ValidateBinding(server.URL, "prod") == nil {
		t.Fatal("SAML session crossed profile or origin")
	}
}

func TestSamlLoginCancellationClosesListener(t *testing.T) {
	var mutex sync.Mutex
	var redirect string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]string
		_ = json.NewDecoder(r.Body).Decode(&request)
		mutex.Lock()
		redirect = request["redirectUri"]
		mutex.Unlock()
		_ = json.NewEncoder(w).Encode(samlStartResponse{AuthenticationURL: "https://login.microsoftonline.com/authorize", ExpiresAt: time.Now().Add(time.Minute)})
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	_, err := NewClient(server.URL, true, logging.Default()).loginWithSaml(ctx, "iteggs", samlInteraction{
		launch: func(context.Context, string) error { cancel(); return nil },
		prompt: func(string, string) (string, error) { t.Fatal("cancelled login requested proof"); return "", nil },
	})
	if err == nil || !strings.Contains(err.Error(), "cancelled") {
		t.Fatal("cancelled login was accepted")
	}
	mutex.Lock()
	callbackURI := redirect
	mutex.Unlock()
	response, err := http.Get(callbackURI)
	if err == nil {
		_ = response.Body.Close()
		t.Fatal("cancelled listener remained open")
	}
}

func TestSamlSessionRefreshPreservesCompanyAndMethod(t *testing.T) {
	newAccess := testJWT(t, time.Now().Add(time.Hour).Unix())
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/refresh-token" {
			t.Errorf("unexpected renewal path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"accessToken": newAccess, "refreshToken": "rotated"})
	}))
	defer server.Close()
	token := expiredLoginSession()
	token.AuthMethod = "saml"
	token.APIOrigin = server.URL
	if err := NewClient(server.URL, true, logging.Default()).RefreshSession(token); err != nil {
		t.Fatal(err)
	}
	if token.AuthMethod != "saml" || token.TenantGUID != "tenant-1" || token.APIOrigin != server.URL || token.RefreshToken != "rotated" {
		t.Fatal("renewal changed the SAML company or method")
	}
}

func TestSamlSessionRejectsMissingTenantAndExpiredProfile(t *testing.T) {
	client := NewClient("https://devback.uteamup.com", false, logging.Default())
	for _, response := range []samlExchangeResponse{
		{Status: "authenticated"},
		{TenantGUID: "tenant", Profile: &LoginResponse{AccessToken: "session", TokenExpiry: time.Now().Add(-time.Hour).Format(time.RFC3339)}},
	} {
		if _, err := client.acceptSamlSession(&response); err == nil {
			t.Fatal("incomplete or expired session was accepted")
		}
	}
}
