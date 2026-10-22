package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/uteamup/cli/internal/security"
)

const (
	cliCallbackPath     = "/auth/cli/callback"
	browserLoginTimeout = 5 * time.Minute
)

// BrowserProviders are the --provider values the web sign-in page starts at once.
var BrowserProviders = []string{"microsoft", "google", "apple"}

type cliHandoffRedeemResponse struct {
	Profile    *LoginResponse `json:"profile"`
	TenantGUID string         `json:"tenantGuid"`
}

// WebOrigin returns the UteamUP web app that signs users in for apiOrigin: override when it is
// set, else the web app paired with a known API host (production, dev, a local backend).
func WebOrigin(apiOrigin, override string) (string, error) {
	if override != "" {
		origin, err := security.Origin(override)
		if err != nil {
			return "", fmt.Errorf("UTEAMUP_WEB_BASE_URL must be a clean HTTPS origin")
		}
		return origin, nil
	}
	if parsed, err := url.Parse(apiOrigin); err == nil {
		host := parsed.Hostname()
		ip := net.ParseIP(host)
		switch {
		case host == "api.uteamup.com":
			return "https://app.uteamup.com", nil
		case host == "devback.uteamup.com":
			return "https://dev.uteamup.com", nil
		case host == "localhost" || (ip != nil && ip.IsLoopback()):
			return "https://localhost:3000", nil
		}
	}
	return "", fmt.Errorf("no known UteamUP web app for %s; set UTEAMUP_WEB_BASE_URL to the address you sign in at", security.SafeText(apiOrigin))
}

// browserSignInURL is the page the CLI opens: /cli/authorize with the loopback request, or
// /login carrying that page as its deep link when a provider should start at once.
func browserSignInURL(webOrigin, redirectURI, state, challenge, provider string) string {
	authorize := "/cli/authorize?" + url.Values{
		"redirect_uri":          {redirectURI},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}.Encode()
	if provider == "" {
		return webOrigin + authorize
	}
	return webOrigin + "/login?" + url.Values{"redirect": {authorize}, "provider": {provider}}.Encode()
}

// LoginWithBrowser signs in through the UteamUP web app, so every sign-in the web offers
// (Microsoft, Google, Apple, password, passkey, company SSO) works for the CLI. The user
// approves the CLI on the web; the code comes back to a 127.0.0.1 listener and is redeemed
// with this process's PKCE verifier. provider, when set, starts that sign-in at once.
func (a *Client) LoginWithBrowser(ctx context.Context, webOrigin, provider string, out io.Writer) (*TokenData, error) {
	return a.loginWithBrowser(ctx, webOrigin, provider, out, launchSystemBrowser)
}

func (a *Client) loginWithBrowser(ctx context.Context, webOrigin, provider string, out io.Writer, launch func(context.Context, string) error) (*TokenData, error) {
	if provider != "" && !slices.Contains(BrowserProviders, provider) {
		return nil, fmt.Errorf("--provider must be one of: %s", strings.Join(BrowserProviders, ", "))
	}
	ctx, cancel := context.WithTimeout(ctx, browserLoginTimeout)
	defer cancel()
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("creating sign-in proof: %w", err)
	}
	state, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("creating sign-in state: %w", err)
	}
	loopback, err := listenLoopback(cliCallbackPath, state, fmt.Errorf("sign-in was declined in the browser"))
	if err != nil {
		return nil, err
	}
	defer loopback.Close()

	signInURL := browserSignInURL(webOrigin, loopback.RedirectURI, state, CodeChallenge(verifier), provider)
	fmt.Fprintf(out, "Continue in your browser to sign in and approve the CLI. If no browser opens, open this address on this computer:\n  %s\n", signInURL)
	if err := launch(ctx, signInURL); err != nil {
		fmt.Fprintln(out, "The browser could not be opened; open the address above.")
	}

	code, err := loopback.wait(ctx)
	if ctx.Err() != nil {
		return nil, fmt.Errorf("browser sign-in cancelled or timed out; run ut login --browser again")
	}
	if err != nil {
		return nil, err
	}
	return a.redeemBrowserCode(ctx, code, verifier)
}

func (a *Client) redeemBrowserCode(ctx context.Context, code, verifier string) (*TokenData, error) {
	body, err := json.Marshal(map[string]string{"code": code, "codeVerifier": verifier})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/api/auth/cli-handoff/redeem", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Cache-Control", "no-store")
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	response, err := a.httpClient().Do(request)
	if err != nil {
		return nil, fmt.Errorf("sign-in service unavailable; check your connection and run ut login --browser again")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sign-in code was refused (HTTP %d); run ut login --browser again", response.StatusCode)
	}
	encoded, err := security.ReadLimit(response.Body, 256*1024)
	var redeemed cliHandoffRedeemResponse
	if err != nil || json.Unmarshal(encoded, &redeemed) != nil {
		return nil, fmt.Errorf("sign-in service returned an invalid response")
	}
	if redeemed.Profile == nil || redeemed.Profile.AccessToken == "" || redeemed.Profile.RefreshToken == "" {
		return nil, fmt.Errorf("browser sign-in returned an incomplete session")
	}

	token := a.sessionFromLoginResponse(redeemed.Profile, redeemed.Profile.Email)
	token.AuthMethod = "browser"
	if samlTenantGUID.MatchString(redeemed.TenantGUID) && redeemed.TenantGUID != "00000000-0000-0000-0000-000000000000" {
		if tenant, err := FetchTenantInfoContext(ctx, token.AccessToken, a.baseURL, redeemed.TenantGUID, a.insecure); err == nil {
			token.TenantID, token.TenantGUID, token.TenantName = tenant.ID, tenant.GUID, tenant.Name
			return token, nil
		}
	}
	// The web had no company selected, or it is not readable: fall back to the default one.
	a.selectTenantForSession(token, redeemed.Profile.DefaultTenantID)
	return token, nil
}
