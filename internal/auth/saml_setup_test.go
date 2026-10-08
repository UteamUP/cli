package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/logging"
)

func TestSamlSetupUsesOwnerStartAndAnonymousPkceExchange(t *testing.T) {
	for _, status := range []string{"setup_validated", "authenticated", "link_required"} {
		t.Run(status, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			owner := &TokenData{AccessToken: "OWNER", RefreshToken: "OWNER-REFRESH", ExpiresAt: time.Now().Add(time.Hour), AuthMethod: "login", Profile: "dev", APIOrigin: "https://devback.uteamup.com"}
			if err := SaveToken(owner); err != nil {
				t.Fatal(err)
			}
			cache, _ := tokenPath()
			before, err := os.ReadFile(cache)
			if err != nil {
				t.Fatal(err)
			}
			const tenant = "11111111-1111-4111-8111-111111111111"
			var mutex sync.Mutex
			var start map[string]string
			var exchange samlExchangeRequest
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mutex.Lock()
				defer mutex.Unlock()
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/api/tenant/saml/test":
					if r.URL.Query().Get("tenantGuid") != tenant || r.Header.Get("Authorization") != "Bearer OWNER" {
						t.Error("setup test lost owner or tenant binding")
					}
					_ = json.NewDecoder(r.Body).Decode(&start)
					if start["tenantGuid"] != "" || start["companyCode"] != "" {
						t.Error("tenant selection must stay in the query")
					}
					_ = json.NewEncoder(w).Encode(samlStartResponse{AuthenticationURL: "https://idp.example/login", ExpiresAt: time.Now().Add(time.Minute)})
				case "/api/auth/saml/exchange":
					if r.Header.Get("Authorization") != "" {
						t.Error("setup redemption must not carry the owner session")
					}
					_ = json.NewDecoder(r.Body).Decode(&exchange)
					if exchange.Code != "SETUP-CODE" || CodeChallenge(exchange.CodeVerifier) != start["codeChallenge"] || exchange.ClientID != "cli" {
						t.Error("setup redemption lost code/PKCE/client binding")
					}
					_ = json.NewEncoder(w).Encode(samlExchangeResponse{Status: status, TenantGUID: tenant})
				default:
					t.Errorf("setup must not issue or load an operational session: %s", r.URL.Path)
				}
			}))
			defer server.Close()
			client := NewClient(server.URL, true, logging.New(logging.LevelError))
			err = client.testSaml(context.Background(), tenant, "OWNER", func(ctx context.Context, _ string) error {
				mutex.Lock()
				redirect := start["redirectUri"]
				state := start["state"]
				mutex.Unlock()
				callback, _ := url.Parse(redirect)
				callback.RawQuery = url.Values{"state": {state}, "code": {"SETUP-CODE"}}.Encode()
				response, err := http.Get(callback.String())
				if err != nil {
					return err
				}
				defer response.Body.Close()
				return nil
			})
			if (err == nil) != (status == "setup_validated") {
				t.Fatalf("setup result %s had unexpected outcome: %v", status, err)
			}
			after, err := os.ReadFile(cache)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("setup test replaced the owner session cache")
			}
		})
	}
}

func TestSamlSetupRequiresOwnerSessionAndTenantGuid(t *testing.T) {
	client := NewClient("https://api.example", false, logging.New(logging.LevelError))
	for _, request := range []struct{ tenant, owner string }{
		{"7", "OWNER"}, {"11111111-1111-4111-8111-111111111111", ""},
	} {
		if err := client.testSaml(context.Background(), request.tenant, request.owner, func(context.Context, string) error { t.Fatal("invalid setup request opened browser"); return nil }); err == nil {
			t.Fatal("invalid setup context accepted")
		}
	}
}
