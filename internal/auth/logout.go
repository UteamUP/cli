package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/uteamup/cli/internal/security"
)

// RevokeSession sends only the captured human credential to its issuing origin.
// API keys have a separate lifecycle; clearing their local JWT does not revoke a key.
func RevokeSession(ctx context.Context, token *TokenData, insecure bool) error {
	if token == nil || (token.AuthMethod != "login" && token.AuthMethod != "saml") {
		return nil
	}
	origin, err := security.Origin(token.APIOrigin)
	if err != nil || origin != token.APIOrigin || token.AccessToken == "" {
		return fmt.Errorf("session origin is unavailable; remote logout could not be verified")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	credential := *token
	// An expired access token must not strand the still-valid refresh credential.
	if !time.Now().Before(credential.ExpiresAt) && credential.RefreshToken != "" {
		if err := NewClient(origin, insecure, nil).refreshSession(ctx, &credential); err != nil {
			return fmt.Errorf("server session renewal was unavailable; remote logout could not be verified")
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, origin+"/api/auth/logout", nil)
	if err != nil {
		return fmt.Errorf("creating logout request")
	}
	req.Header.Set("Authorization", "Bearer "+credential.AccessToken)
	resp, err := tenantHTTPClient(origin, insecure).Do(req)
	if err != nil {
		return fmt.Errorf("server logout was unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server logout returned status %d", resp.StatusCode)
	}
	return nil
}
