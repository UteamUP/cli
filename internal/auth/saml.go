package auth

import (
	"bufio"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/uteamup/cli/internal/security"
)

const samlCallbackPath = "/auth/saml/callback"

type samlStartResponse struct {
	AuthenticationURL string    `json:"authenticationUrl"`
	ExpiresAt         time.Time `json:"expiresAt"`
}

type samlExchangeRequest struct {
	Code              string `json:"code,omitempty"`
	ContinuationToken string `json:"continuationToken,omitempty"`
	CodeVerifier      string `json:"codeVerifier"`
	ClientID          string `json:"clientId"`
	VerificationCode  string `json:"verificationCode,omitempty"`
	MfaCode           string `json:"mfaCode,omitempty"`
}

type samlExchangeResponse struct {
	Status            string         `json:"status"`
	Profile           *LoginResponse `json:"profile"`
	TenantGUID        string         `json:"tenantGuid"`
	ContinuationToken string         `json:"continuationToken"`
	Email             string         `json:"email"`
	ExpiresAt         time.Time      `json:"expiresAt"`
}

type samlInteraction struct {
	launch func(context.Context, string) error
	prompt func(string, string) (string, error)
}

// LoginWithSAML opens the company identity provider in the system browser.
func (a *Client) LoginWithSAML(ctx context.Context, company string) (*TokenData, error) {
	reader := bufio.NewReader(os.Stdin)
	interaction := samlInteraction{
		launch: launchSamlBrowser,
		prompt: func(status, email string) (string, error) {
			if status == "link_required" {
				fmt.Fprintf(os.Stderr, "Verify ownership of %s using the code sent by UteamUP.\n", security.SafeText(email))
				return PromptSecret(reader, "Email verification code: ")
			}
			fmt.Fprintln(os.Stderr, "This account requires a second factor.")
			return PromptSecret(reader, "Authenticator or recovery code: ")
		},
	}
	return a.loginWithSaml(ctx, company, interaction)
}

func (a *Client) loginWithSaml(ctx context.Context, company string, interaction samlInteraction) (*TokenData, error) {
	company = strings.TrimSpace(company)
	if company == "" {
		return nil, fmt.Errorf("SAML login requires --company with your company code")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("creating SAML proof: %w", err)
	}
	state, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("creating SAML state: %w", err)
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("opening SAML callback listener: %w", err)
	}
	defer listener.Close()
	callback := make(chan samlCallback, 1)
	server := &http.Server{
		Handler:           samlCallbackHandler(state, listener.Addr().String(), callback),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       5 * time.Second,
		MaxHeaderBytes:    8192,
	}
	defer server.Close()
	go func() { _ = server.Serve(listener) }()
	var start samlStartResponse
	err = a.samlPost(ctx, "/api/auth/saml/start", map[string]string{
		"companyCode": company, "clientId": "cli",
		"redirectUri": "http://" + listener.Addr().String() + samlCallbackPath,
		"state":       state, "codeChallenge": CodeChallenge(verifier), "codeChallengeMethod": "S256",
	}, &start)
	if err != nil {
		return nil, err
	}
	authorize, err := url.Parse(start.AuthenticationURL)
	if err != nil || authorize.Scheme != "https" || authorize.Hostname() == "" || authorize.User != nil || authorize.Fragment != "" || !start.ExpiresAt.After(time.Now()) {
		return nil, fmt.Errorf("server returned an invalid or expired SAML sign-in link")
	}
	if err := interaction.launch(ctx, authorize.String()); err != nil {
		return nil, fmt.Errorf("opening SAML sign-in: %w", err)
	}
	select {
	case result := <-callback:
		_ = server.Close()
		if result.err != nil {
			return nil, result.err
		}
		return a.completeSamlLogin(ctx, result.code, verifier, interaction.prompt)
	case <-ctx.Done():
		return nil, fmt.Errorf("SAML sign-in cancelled or timed out; retry with --saml --company %s", security.SafeText(company))
	}
}

type samlCallback struct {
	code string
	err  error
}

func samlCallbackHandler(state, host string, callback chan<- samlCallback) http.Handler {
	var completed atomic.Bool
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil || r.Method != http.MethodGet || r.Host != host || r.URL.Path != samlCallbackPath || len(r.URL.RawQuery) > 8192 || !validSamlCallback(query, state) {
			http.Error(w, "Invalid sign-in callback. Return to UteamUP CLI.", http.StatusBadRequest)
			return
		}
		if !completed.CompareAndSwap(false, true) {
			http.Error(w, "This sign-in callback was already received.", http.StatusGone)
			return
		}
		result := samlCallback{code: query.Get("code")}
		if query.Get("error") != "" {
			result.err = fmt.Errorf("SAML sign-in was cancelled or refused; retry using your company code")
		}
		callback <- result
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Return to UteamUP CLI to finish sign-in. You can close this window."))
	})
}

func validSamlCallback(query url.Values, state string) bool {
	if len(query) != 2 || len(query["state"]) != 1 || subtle.ConstantTimeCompare([]byte(query.Get("state")), []byte(state)) != 1 {
		return false
	}
	return (len(query["code"]) == 1 && query.Get("code") != "" && len(query["error"]) == 0) ||
		(len(query["error"]) == 1 && query.Get("error") != "" && len(query["code"]) == 0)
}

func (a *Client) completeSamlLogin(ctx context.Context, code, verifier string, prompt func(string, string) (string, error)) (*TokenData, error) {
	request := samlExchangeRequest{Code: code, CodeVerifier: verifier, ClientID: "cli"}
	var challenge *samlExchangeResponse
	for attempts := 0; attempts < 6; attempts++ {
		var response samlExchangeResponse
		if err := a.samlPost(ctx, "/api/auth/saml/exchange", request, &response); err != nil {
			if challenge == nil || ctx.Err() != nil {
				return nil, err
			}
			fmt.Fprintln(os.Stderr, "Verification failed. Check the code and try again.")
			response = *challenge
		}
		if response.Status == "authenticated" {
			return a.acceptSamlSession(&response)
		}
		if (response.Status != "link_required" && response.Status != "mfa_required") || response.ContinuationToken == "" || !response.ExpiresAt.After(time.Now()) {
			return nil, fmt.Errorf("SAML sign-in could not complete; start again using your company code")
		}
		challenge = &response
		proof, err := prompt(response.Status, response.Email)
		if err != nil || strings.TrimSpace(proof) == "" {
			return nil, fmt.Errorf("SAML verification cancelled")
		}
		request = samlExchangeRequest{ContinuationToken: response.ContinuationToken, CodeVerifier: verifier, ClientID: "cli"}
		if response.Status == "link_required" {
			request.VerificationCode = strings.TrimSpace(proof)
		} else {
			request.MfaCode = strings.TrimSpace(proof)
		}
	}
	return nil, fmt.Errorf("SAML verification attempt limit reached; start again using your company code")
}

func (a *Client) acceptSamlSession(response *samlExchangeResponse) (*TokenData, error) {
	if response.Profile == nil || response.Profile.AccessToken == "" || response.TenantGUID == "" {
		return nil, fmt.Errorf("SAML sign-in returned an incomplete session")
	}
	expiry, err := time.Parse(time.RFC3339, response.Profile.TokenExpiry)
	if err != nil || !expiry.After(time.Now()) {
		return nil, fmt.Errorf("SAML sign-in returned an expired session")
	}
	tenant, err := FetchTenantInfo(response.Profile.AccessToken, a.baseURL, response.TenantGUID, a.insecure)
	if err != nil {
		return nil, fmt.Errorf("confirming SAML company membership failed; retry sign-in")
	}
	token := a.sessionFromLoginResponse(response.Profile, response.Profile.Email)
	token.AuthMethod = "saml"
	token.APIOrigin = a.baseURL
	token.TenantID, token.TenantGUID, token.TenantName = tenant.ID, tenant.GUID, tenant.Name
	return token, nil
}

func (a *Client) samlPost(ctx context.Context, path string, data, result any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	response, err := a.httpClient().Do(request)
	if err != nil {
		return fmt.Errorf("SAML service unavailable; check your connection and retry")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("SAML request refused (HTTP %d); check your company code or verification code and retry", response.StatusCode)
	}
	encoded, err := security.ReadLimit(response.Body, 256*1024)
	if err != nil || json.Unmarshal(encoded, result) != nil {
		return fmt.Errorf("SAML service returned an invalid response")
	}
	return nil
}

func launchSamlBrowser(ctx context.Context, rawURL string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.CommandContext(ctx, "open", rawURL)
	case "windows":
		command = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		command = exec.CommandContext(ctx, "xdg-open", rawURL)
	}
	if err := command.Run(); err != nil {
		return fmt.Errorf("system browser could not open; check the default browser and retry")
	}
	return nil
}
