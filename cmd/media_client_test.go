package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/config"
)

func validMediaToken() *auth.TokenData {
	return &auth.TokenData{
		AccessToken: "token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "b966b8c7-04a4-45d4-aa51-519ecf2ef13a",
	}
}

func TestValidateMediaTenantRequiresMatchingGUID(t *testing.T) {
	profile := &config.Profile{TenantGUID: "b966b8c7-04a4-45d4-aa51-519ecf2ef13a"}
	if err := validateMediaTenant(profile, validMediaToken()); err != nil {
		t.Fatal(err)
	}

	token := validMediaToken()
	token.TenantGUID = "not-a-guid"
	if err := validateMediaTenant(profile, token); err == nil || strings.Contains(err.Error(), token.TenantGUID) {
		t.Fatalf("invalid tenant GUID should be rejected without echoing it: %v", err)
	}

	profile.TenantGUID = "d647e60a-e756-4eea-b72c-7bc801911517"
	if err := validateMediaTenant(profile, validMediaToken()); err == nil {
		t.Fatal("expected tenant mismatch to be rejected")
	}
}

func TestNewMediaAPIClientRejectsUnsafeLimits(t *testing.T) {
	profile := &config.Profile{BaseURL: "http://api.uteamup.com", MaxRetries: 0}
	if _, err := newMediaAPIClient(profile, time.Minute); err == nil {
		t.Fatal("expected non-HTTPS API origin to be rejected")
	}

	profile.BaseURL = "https://api.uteamup.com"
	if _, err := newMediaAPIClient(profile, time.Millisecond); err == nil {
		t.Fatal("expected sub-second media timeout to be rejected")
	}

	profile.MaxRetries = 6
	if _, err := newMediaAPIClient(profile, time.Minute); err == nil {
		t.Fatal("expected excessive retries to be rejected")
	}
}
