package auth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/logging"
)

// OAuthTokenResponse represents the backend token endpoint response.
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// LoginResponse represents the backend login endpoint response.
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int    `json:"expiresIn,omitempty"`
}

// AuthClient handles authentication flows.
type AuthClient struct {
	baseURL  string
	insecure bool
	logger   *logging.Logger
}

// NewAuthClient creates an AuthClient.
func NewAuthClient(baseURL string, insecure bool, logger *logging.Logger) *AuthClient {
	return &AuthClient{baseURL: strings.TrimRight(baseURL, "/"), insecure: insecure, logger: logger}
}

func (a *AuthClient) httpClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if a.insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-requested for dev
	}
	return &http.Client{Transport: transport, Timeout: 30 * time.Second}
}

// LoginWithCredentials authenticates with email/password and returns a token.
func (a *AuthClient) LoginWithCredentials(email, password string) (*TokenData, error) {
	a.logger.Debug("attempting login for %s", email)

	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
	req, err := http.NewRequest("POST", a.baseURL+"/api/auth/login", strings.NewReader(body))
	if err != nil {
		return nil, clierrors.NewAuthError("creating login request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient().Do(req)
	if err != nil {
		return nil, clierrors.NewAuthError("login request failed", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, clierrors.NewAuthError(
			fmt.Sprintf("login failed with status %d: %s", resp.StatusCode, string(respBody)), nil,
		)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		return nil, clierrors.NewAuthError("parsing login response", err)
	}

	expiresIn := loginResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 86400 // default 24h
	}

	token := &TokenData{
		AccessToken:  loginResp.Token,
		RefreshToken: loginResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
		AuthMethod:   "login",
		Email:        email,
	}

	a.logger.Info("login successful for %s", email)
	return token, nil
}

// LoginWithAPIKey authenticates using OAuth 2.0 + PKCE with an API key.
func (a *AuthClient) LoginWithAPIKey(apiKey, secret string) (*TokenData, error) {
	a.logger.Debug("attempting API key auth")

	if len(apiKey) != 32 {
		return nil, clierrors.NewValidationError("apiKey", "must be exactly 32 characters")
	}
	if len(secret) < 64 {
		return nil, clierrors.NewValidationError("secret", "must be at least 64 characters")
	}

	// Step 1: Generate PKCE pair
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, clierrors.NewAuthError("generating PKCE verifier", err)
	}
	challenge := CodeChallenge(verifier)

	// Step 2: Authorization request
	authURL := fmt.Sprintf(
		"%s/oauth/authorize?response_type=code&client_id=%s&code_challenge=%s&code_challenge_method=S256&redirect_uri=urn:ietf:wg:oauth:2.0:oob",
		a.baseURL, url.QueryEscape(apiKey), url.QueryEscape(challenge),
	)

	client := a.httpClient()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // don't follow redirects
	}

	authResp, err := client.Get(authURL)
	if err != nil {
		return nil, clierrors.NewAuthError("authorization request failed", err)
	}
	defer authResp.Body.Close()

	// Extract authorization code from Location header or response body
	code := ""
	if loc := authResp.Header.Get("Location"); loc != "" {
		parsed, _ := url.Parse(loc)
		if parsed != nil {
			code = parsed.Query().Get("code")
		}
	}
	if code == "" {
		body, _ := io.ReadAll(authResp.Body)
		var authBody struct {
			Code string `json:"code"`
		}
		if json.Unmarshal(body, &authBody) == nil && authBody.Code != "" {
			code = authBody.Code
		}
	}
	if code == "" {
		return nil, clierrors.NewAuthError("no authorization code received", nil)
	}

	// Step 3: Token exchange
	tokenData := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {apiKey},
		"client_secret": {secret},
		"code_verifier": {verifier},
		"redirect_uri":  {"urn:ietf:wg:oauth:2.0:oob"},
	}

	tokenReq, err := http.NewRequest("POST", a.baseURL+"/oauth/token", strings.NewReader(tokenData.Encode()))
	if err != nil {
		return nil, clierrors.NewAuthError("creating token request", err)
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := a.httpClient().Do(tokenReq)
	if err != nil {
		return nil, clierrors.NewAuthError("token exchange failed", err)
	}
	defer tokenResp.Body.Close()

	tokenBody, _ := io.ReadAll(tokenResp.Body)
	if tokenResp.StatusCode != http.StatusOK {
		return nil, clierrors.NewAuthError(
			fmt.Sprintf("token exchange failed with status %d: %s", tokenResp.StatusCode, string(tokenBody)), nil,
		)
	}

	var oauthResp OAuthTokenResponse
	if err := json.Unmarshal(tokenBody, &oauthResp); err != nil {
		return nil, clierrors.NewAuthError("parsing token response", err)
	}

	expiresIn := oauthResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 604800 // default 7 days
	}

	result := &TokenData{
		AccessToken:  oauthResp.AccessToken,
		RefreshToken: oauthResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
		AuthMethod:   "apikey",
	}

	a.logger.Info("API key auth successful")
	return result, nil
}
