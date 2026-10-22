package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/logging"
)

func TestWebOriginPairsKnownBackendsAndHonoursTheOverride(t *testing.T) {
	for apiOrigin, want := range map[string]string{
		"https://api.uteamup.com":     "https://app.uteamup.com",
		"https://devback.uteamup.com": "https://dev.uteamup.com",
		"https://localhost:5002":      "https://localhost:3000",
		"https://127.0.0.1:5002":      "https://localhost:3000",
	} {
		got, err := WebOrigin(apiOrigin, "")
		if err != nil || got != want {
			t.Errorf("WebOrigin(%s) = %q, %v; want %q", apiOrigin, got, err, want)
		}
	}
	if got, err := WebOrigin("https://prufaback.uteamup.com", "https://prufa.uteamup.com/"); err != nil || got != "https://prufa.uteamup.com" {
		t.Errorf("override not used: %q, %v", got, err)
	}
	for _, unknown := range []string{"https://prufaback.uteamup.com", "https://api.uteamup.com.evil.example", "https://evil.example"} {
		if _, err := WebOrigin(unknown, ""); err == nil || !strings.Contains(err.Error(), "UTEAMUP_WEB_BASE_URL") {
			t.Errorf("WebOrigin(%s) guessed a web app instead of asking for the override", unknown)
		}
	}
	for _, unsafe := range []string{"http://app.uteamup.com", "https://app.uteamup.com/login?x=1", "https://user@app.uteamup.com"} {
		if _, err := WebOrigin("https://api.uteamup.com", unsafe); err == nil {
			t.Errorf("override %q accepted", unsafe)
		}
	}
}

func TestBrowserSignInURLOpensTheApprovalPageOrLoginWithItAsDeepLink(t *testing.T) {
	const callback = "http://127.0.0.1:53682/auth/cli/callback"
	direct, err := url.Parse(browserSignInURL("https://app.uteamup.com", callback, "state", "challenge", ""))
	if err != nil || direct.Path != "/cli/authorize" {
		t.Fatalf("unexpected approval URL %v", direct)
	}
	query := direct.Query()
	if query.Get("redirect_uri") != callback || query.Get("state") != "state" || query.Get("code_challenge") != "challenge" || query.Get("code_challenge_method") != "S256" {
		t.Fatalf("approval URL lost the request: %s", direct.RawQuery)
	}

	hinted, err := url.Parse(browserSignInURL("https://app.uteamup.com", callback, "state", "challenge", "google"))
	if err != nil || hinted.Path != "/login" || hinted.Query().Get("provider") != "google" {
		t.Fatalf("unexpected hinted URL %v", hinted)
	}
	if deepLink := hinted.Query().Get("redirect"); deepLink != direct.RequestURI() {
		t.Fatalf("login deep link %q does not carry the approval page %q", deepLink, direct.RequestURI())
	}
}

// approveInBrowser stands in for the web page: it reads the CLI request out of the URL the CLI
// opened and sends the browser to the loopback callback, as /cli/authorize does after Approve.
func approveInBrowser(t *testing.T, outcome url.Values, opened *string) func(context.Context, string) error {
	return func(_ context.Context, raw string) error {
		*opened = raw
		page, err := url.Parse(raw)
		if err != nil {
			return err
		}
		request := page.Query()
		if page.Path == "/login" {
			deepLink, err := url.Parse(request.Get("redirect"))
			if err != nil {
				return err
			}
			request = deepLink.Query()
		}
		callback, err := url.Parse(request.Get("redirect_uri"))
		if err != nil || callback.Hostname() != "127.0.0.1" || callback.Path != cliCallbackPath {
			t.Errorf("unsafe callback %q", request.Get("redirect_uri"))
			return err
		}
		outcome.Set("state", request.Get("state"))
		callback.RawQuery = outcome.Encode()
		response, err := http.Get(callback.String())
		if err == nil {
			_ = response.Body.Close()
		}
		return err
	}
}

func TestBrowserLoginRedeemsTheCodeWithItsVerifierAndKeepsTheWebCompany(t *testing.T) {
	var challenge string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/cli-handoff/redeem":
			var request map[string]string
			_ = json.NewDecoder(r.Body).Decode(&request)
			if r.Method != http.MethodPost || request["code"] != "opaque-code" || CodeChallenge(request["codeVerifier"]) != challenge {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(cliHandoffRedeemResponse{
				TenantGUID: "11111111-1111-1111-1111-111111111111",
				Profile: &LoginResponse{AccessToken: "session", RefreshToken: "refresh", Email: "gisli@uteamup.com",
					TokenExpiry: time.Now().Add(time.Hour).Format(time.RFC3339)},
			})
		case "/api/tenant/my-tenants":
			if r.Header.Get("Authorization") != "Bearer session" {
				t.Error("tenant was not read with the issued session")
			}
			_ = json.NewEncoder(w).Encode([]TenantInfo{
				{ID: 9, GUID: "22222222-2222-2222-2222-222222222222", Name: "Other"},
				{ID: 12, GUID: "11111111-1111-1111-1111-111111111111", Name: "Iteggs"},
			})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	var opened string
	launch := approveInBrowser(t, url.Values{"code": {"opaque-code"}}, &opened)
	recordChallenge := func(ctx context.Context, raw string) error {
		page, _ := url.Parse(raw)
		deepLink, _ := url.Parse(page.Query().Get("redirect"))
		challenge = deepLink.Query().Get("code_challenge")
		return launch(ctx, raw)
	}
	var out bytes.Buffer
	token, err := NewClient(server.URL, true, logging.Default()).loginWithBrowser(context.Background(), "https://app.uteamup.com", "microsoft", &out, recordChallenge)
	if err != nil {
		t.Fatal(err)
	}
	if token.AuthMethod != "browser" || !token.IsUserSession() || token.RefreshToken != "refresh" || token.Email != "gisli@uteamup.com" {
		t.Fatalf("unexpected session %+v", token)
	}
	if token.TenantID != 12 || token.TenantName != "Iteggs" {
		t.Fatalf("web company not kept: %+v", token)
	}
	if !strings.HasPrefix(opened, "https://app.uteamup.com/login?") || !strings.Contains(out.String(), opened) {
		t.Fatalf("sign-in address not opened and printed: %q / %q", opened, out.String())
	}
}

func TestBrowserLoginStopsWhenTheUserDeclines(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("declined sign-in reached the API: %s", r.URL.Path)
	}))
	defer server.Close()
	var opened string
	_, err := NewClient(server.URL, true, logging.Default()).loginWithBrowser(context.Background(), "https://app.uteamup.com", "",
		&bytes.Buffer{}, approveInBrowser(t, url.Values{"error": {"access_denied"}}, &opened))
	if err == nil || !strings.Contains(err.Error(), "declined") {
		t.Fatalf("declined sign-in accepted: %v", err)
	}
	if !strings.HasPrefix(opened, "https://app.uteamup.com/cli/authorize?") {
		t.Fatalf("unexpected page opened %q", opened)
	}
}

func TestBrowserLoginRefusesAnUnknownProviderBeforeOpeningAnything(t *testing.T) {
	_, err := NewClient("https://api.uteamup.com", false, logging.Default()).loginWithBrowser(context.Background(), "https://app.uteamup.com", "github",
		&bytes.Buffer{}, func(context.Context, string) error {
			t.Fatal("browser opened for an unknown provider")
			return nil
		})
	if err == nil || !strings.Contains(err.Error(), "--provider") {
		t.Fatalf("unknown provider accepted: %v", err)
	}
}

func TestBrowserLoginReportsARefusedCode(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	var opened string
	_, err := NewClient(server.URL, true, logging.Default()).loginWithBrowser(context.Background(), "https://app.uteamup.com", "",
		&bytes.Buffer{}, approveInBrowser(t, url.Values{"code": {"stale"}}, &opened))
	if err == nil || !strings.Contains(err.Error(), "HTTP 401") {
		t.Fatalf("refused code accepted: %v", err)
	}
}

func TestBrowserSessionIsRenewableAndRevocableButAPIKeyIsNot(t *testing.T) {
	for method, want := range map[string]bool{"login": true, "saml": true, "browser": true, "apikey": false, "": false} {
		if got := (&TokenData{AuthMethod: method}).IsUserSession(); got != want {
			t.Errorf("IsUserSession(%q) = %v, want %v", method, got, want)
		}
	}
}
