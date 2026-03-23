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
// Maps to ProfileModel in C# backend.
type LoginResponse struct {
	AccessToken     string `json:"accessToken"`
	RefreshToken    string `json:"refreshToken,omitempty"`
	TokenExpiry     string `json:"tokenExpiry,omitempty"`
	DefaultTenantID int    `json:"defaultTenantId,omitempty"`
	HasTenants      bool   `json:"hasTenants"`
	TenantCount     int    `json:"tenantCount"`
}

// TenantInfo represents a tenant from the my-tenants endpoint.
// Maps to TenantResponseModel in C# backend.
type TenantInfo struct {
	ID       int    `json:"id"`
	Guid     string `json:"guid"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
	PlanID   int    `json:"planId"`
	PlanName string `json:"planName"`
}

// HasPlan returns true if the tenant has an active subscription plan.
func (t *TenantInfo) HasPlan() bool {
	return t.PlanID > 0 && t.PlanName != ""
}

// FetchTenantInfo calls the my-tenants endpoint and returns tenant info for the
// given tenant GUID. If tenantGuid is empty, returns the default/first tenant.
// Requires a valid access token and the backend base URL.
func FetchTenantInfo(accessToken, baseURL, tenantGuid string) (*TenantInfo, error) {
	req, err := http.NewRequest("GET", strings.TrimRight(baseURL, "/")+"/api/tenant/my-tenants", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // dev support
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching tenants: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("my-tenants returned %d: %s", resp.StatusCode, string(body))
	}

	var tenants []TenantInfo
	if err := json.Unmarshal(body, &tenants); err != nil {
		return nil, fmt.Errorf("parsing tenants: %w", err)
	}

	if len(tenants) == 0 {
		return nil, fmt.Errorf("no tenants found for this user")
	}

	// If a specific tenant GUID is requested, find it.
	if tenantGuid != "" {
		for _, t := range tenants {
			if strings.EqualFold(t.Guid, tenantGuid) {
				return &t, nil
			}
		}
		return nil, fmt.Errorf("tenant with GUID %q not found — you may not have access to this tenant", tenantGuid)
	}

	// Default: return first tenant.
	return &tenants[0], nil
}

// FetchAllTenants calls the my-tenants endpoint and returns all tenants
// the user has access to. Used for interactive tenant selection.
func FetchAllTenants(accessToken, baseURL string) ([]TenantInfo, error) {
	req, err := http.NewRequest("GET", strings.TrimRight(baseURL, "/")+"/api/tenant/my-tenants", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // dev support
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching tenants: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("my-tenants returned %d: %s", resp.StatusCode, string(body))
	}

	var tenants []TenantInfo
	if err := json.Unmarshal(body, &tenants); err != nil {
		return nil, fmt.Errorf("parsing tenants: %w", err)
	}

	return tenants, nil
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

	// Parse expiry from response or default to 7 days
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if loginResp.TokenExpiry != "" {
		if parsed, err := time.Parse(time.RFC3339, loginResp.TokenExpiry); err == nil {
			expiresAt = parsed
		}
	}

	token := &TokenData{
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		ExpiresAt:    expiresAt,
		AuthMethod:   "login",
		Email:        email,
	}

	// Fetch tenants to get tenant ID and GUID
	tenants, err := a.fetchMyTenants(token.AccessToken)
	if err != nil {
		a.logger.Warn("could not fetch tenants: %v", err)
	} else if len(tenants) > 0 {
		// Use the first tenant (or match DefaultTenantId)
		selected := tenants[0]
		for _, t := range tenants {
			if t.ID == loginResp.DefaultTenantID {
				selected = t
				break
			}
		}
		token.TenantID = selected.ID
		token.TenantGuid = selected.Guid
		token.TenantName = selected.Name
		a.logger.Info("selected tenant: %s (ID: %d)", selected.Name, selected.ID)
	}

	a.logger.Info("login successful for %s", email)
	return token, nil
}

// fetchMyTenants calls GET /api/tenant/my-tenants with the access token.
func (a *AuthClient) fetchMyTenants(accessToken string) ([]TenantInfo, error) {
	req, err := http.NewRequest("GET", a.baseURL+"/api/tenant/my-tenants", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("my-tenants returned %d: %s", resp.StatusCode, string(body))
	}

	var tenants []TenantInfo
	if err := json.Unmarshal(body, &tenants); err != nil {
		return nil, fmt.Errorf("parsing tenants: %w", err)
	}

	return tenants, nil
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
