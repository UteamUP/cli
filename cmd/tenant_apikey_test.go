package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/config"
)

func TestAPIKeyTenantCommandsDoNotQueryUserMemberships(t *testing.T) {
	taskCLIHome := t.TempDir()
	t.Setenv("HOME", taskCLIHome)
	t.Setenv("USERPROFILE", taskCLIHome)
	t.Setenv("UTEAMUP_API_BASE_URL", "")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("machine identity queried user endpoint: %s", r.URL.Path)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	cfg := config.DefaultConfig()
	profile := cfg.Profiles[cfg.ActiveProfile]
	profile.BaseURL = server.URL
	cfg.Profiles[cfg.ActiveProfile] = profile
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	const tenantGUID = "b966b8c7-04a4-45d4-aa51-519ecf2ef13a"
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin: server.URL, Profile: cfg.ActiveProfile, AuthMethod: "apikey",
		AccessToken: "synthetic-api-key-token", ExpiresAt: time.Now().Add(time.Hour), TenantGUID: tenantGUID,
	}); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	previousWriter := tenantShowCmd.OutOrStdout()
	tenantShowCmd.SetOut(&output)
	t.Cleanup(func() { tenantShowCmd.SetOut(previousWriter) })
	if err := tenantShowCmd.RunE(tenantShowCmd, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), tenantGUID) || !strings.Contains(output.String(), "bound") {
		t.Fatalf("show did not explain the key's tenant: %s", output.String())
	}
	if err := tenantSelectCmd.RunE(tenantSelectCmd, nil); err == nil || !strings.Contains(err.Error(), "user sign-in") {
		t.Fatalf("select did not explain the machine identity: %v", err)
	}
}
