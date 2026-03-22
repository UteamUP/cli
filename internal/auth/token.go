package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/uteamup/cli/internal/config"
)

const tokenFileName = "token.json"

// TokenData holds cached authentication tokens and tenant context.
type TokenData struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	ExpiresAt    time.Time `json:"expiresAt"`
	AuthMethod   string    `json:"authMethod"` // "login" or "apikey"
	Email        string    `json:"email,omitempty"`
	Profile      string    `json:"profile"`
	TenantID     int       `json:"tenantId,omitempty"`
	TenantGuid   string    `json:"tenantGuid,omitempty"`
	TenantName   string    `json:"tenantName,omitempty"`
}

// IsValid returns true if the token exists and is not expired (with 5-minute margin).
func (t *TokenData) IsValid() bool {
	if t.AccessToken == "" {
		return false
	}
	return time.Now().Before(t.ExpiresAt.Add(-5 * time.Minute))
}

// tokenPath returns ~/.uteamup/token.json.
func tokenPath() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, tokenFileName), nil
}

// LoadToken reads the cached token from disk.
func LoadToken() (*TokenData, error) {
	path, err := tokenPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading token: %w", err)
	}

	var token TokenData
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	return &token, nil
}

// SaveToken writes the token to disk with 0600 permissions.
func SaveToken(token *TokenData) error {
	dir, err := config.ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	path := filepath.Join(dir, tokenFileName)
	return os.WriteFile(path, data, 0600)
}

// ClearToken deletes the cached token file.
func ClearToken() error {
	path, err := tokenPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
