package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/security"
)

// RefreshSession renews a password or SAML session with its refresh token, so a signed-in
// user is not sent back to "uteamup login" each time the access token expires.
//
// The server rotates refresh tokens: the old one stops working once it is used, so both
// values are replaced. On any failure the token is left untouched.
func (a *Client) RefreshSession(token *TokenData) error {
	return a.refreshSession(context.Background(), token)
}

func (a *Client) refreshSession(ctx context.Context, token *TokenData) error {
	if token == nil || (token.AuthMethod != "login" && token.AuthMethod != "saml") || token.RefreshToken == "" {
		return clierrors.NewAuthError("this session cannot be renewed; sign in again", nil)
	}

	body, err := json.Marshal(map[string]string{"refreshToken": token.RefreshToken})
	if err != nil {
		return clierrors.NewAuthError("creating session renewal request", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/api/auth/refresh-token", bytes.NewReader(body))
	if err != nil {
		return clierrors.NewAuthError("creating session renewal request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient().Do(req)
	if err != nil {
		return clierrors.NewAuthError("session renewal request failed", err)
	}
	defer resp.Body.Close()

	respBody, _ := security.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return clierrors.NewAuthError(fmt.Sprintf("session renewal failed with status %d", resp.StatusCode), nil)
	}

	var renewed struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.Unmarshal(respBody, &renewed); err != nil || renewed.AccessToken == "" {
		return clierrors.NewAuthError("parsing session renewal response", err)
	}
	expiresAt, err := accessTokenExpiry(renewed.AccessToken)
	if err != nil {
		return clierrors.NewAuthError("reading the renewed session's expiry", err)
	}

	token.AccessToken = renewed.AccessToken
	if renewed.RefreshToken != "" {
		token.RefreshToken = renewed.RefreshToken
	}
	token.ExpiresAt = expiresAt
	return nil
}

// accessTokenExpiry reads the exp claim of a JWT. The signature is not checked: the CLI
// only needs to know when to renew, and the backend validates every request anyway.
func accessTokenExpiry(jwt string) (time.Time, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("access token is not a JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("decoding access token payload: %w", err)
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, fmt.Errorf("parsing access token payload: %w", err)
	}
	if claims.Exp == 0 {
		return time.Time{}, fmt.Errorf("access token has no expiry")
	}
	return time.Unix(claims.Exp, 0), nil
}
