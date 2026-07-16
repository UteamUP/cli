package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/config"
	"github.com/uteamup/cli/internal/logging"
)

const maxMediaTimeout = 15 * time.Minute

var tenantGUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func newMediaAPIClient(profile *config.Profile, timeout time.Duration) (*client.APIClient, error) {
	if profile == nil {
		return nil, fmt.Errorf("active profile is required for media analysis")
	}
	baseURL := strings.TrimSpace(profile.BaseURL)
	if err := client.ValidateBaseURL(baseURL); err != nil {
		return nil, err
	}
	if timeout < time.Second || timeout > maxMediaTimeout {
		return nil, fmt.Errorf("media timeout must be between 1 second and %s", maxMediaTimeout)
	}
	if profile.MaxRetries < 0 || profile.MaxRetries > 5 {
		return nil, fmt.Errorf("maxRetries must be between 0 and 5 for media uploads")
	}

	level := logging.ParseLevel(profile.LogLevel)
	if verbose {
		level = logging.LevelDebug
	}
	retries := client.RetryOptions{
		MaxRetries: profile.MaxRetries,
		BaseDelay:  time.Second,
		MaxDelay:   10 * time.Second,
	}
	return client.NewAPIClient(baseURL, timeout, insecure, retries, logging.New(level)), nil
}

func validateMediaTenant(profile *config.Profile, token *auth.TokenData) error {
	if profile == nil {
		return fmt.Errorf("active profile is required for media analysis")
	}
	if token == nil || !token.IsValid() {
		return fmt.Errorf("not authenticated; run \"uteamup login\" first")
	}
	tokenTenantGUID := strings.TrimSpace(token.TenantGuid)
	if !tenantGUIDPattern.MatchString(tokenTenantGUID) {
		return fmt.Errorf("the authenticated session has no tenant GUID; sign in to a tenant")
	}
	profileTenantGUID := strings.TrimSpace(profile.TenantGuid)
	if profileTenantGUID != "" && !tenantGUIDPattern.MatchString(profileTenantGUID) {
		return fmt.Errorf("the active profile contains an invalid tenant GUID")
	}
	if profileTenantGUID != "" && !strings.EqualFold(profileTenantGUID, tokenTenantGUID) {
		return fmt.Errorf("the active profile tenant does not match the authenticated tenant; sign in again")
	}
	return nil
}
