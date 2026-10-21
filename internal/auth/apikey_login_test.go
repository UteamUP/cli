package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/uteamup/cli/internal/logging"
)

// The CLI calls REST routes with the API-key token, so the token exchange must name the REST
// resource (RFC 8707). Without it, a backend that confines resource-less tokens to /mcp would
// answer every CLI command with 401.
func TestLoginWithAPIKeyRequestsTheRestResource(t *testing.T) {
	apiKey := strings.Repeat("k", 32)
	secret := strings.Repeat("s", 64)
	var tokenForm map[string]string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/oauth/authorize":
			if r.URL.Query().Get("client_id") != apiKey || r.URL.Query().Get("code_challenge_method") != "S256" {
				t.Errorf("authorize query = %v", r.URL.Query())
			}
			http.Redirect(w, r, "urn:ietf:wg:oauth:2.0:oob?code=auth-code", http.StatusFound)
		case r.Method == http.MethodPost && r.URL.Path == "/oauth/token":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse token form: %v", err)
			}
			tokenForm = map[string]string{}
			for key := range r.PostForm {
				tokenForm[key] = r.PostForm.Get(key)
			}
			fmt.Fprint(w, `{"access_token":"rest-token","token_type":"Bearer","expires_in":3600}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	token, err := NewClient(server.URL, true, logging.New(logging.LevelError)).LoginWithAPIKey(apiKey, secret)
	if err != nil {
		t.Fatalf("LoginWithAPIKey: %v", err)
	}

	if got, want := tokenForm["resource"], server.URL+"/api"; got != want {
		t.Errorf("resource = %q, want %q", got, want)
	}
	if tokenForm["code"] != "auth-code" || tokenForm["client_id"] != apiKey || tokenForm["code_verifier"] == "" {
		t.Errorf("token form = %v", tokenForm)
	}
	if token.AccessToken != "rest-token" || token.AuthMethod != "apikey" {
		t.Errorf("token = %+v", token)
	}
}
