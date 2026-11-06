package registry

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

// runUtilityMeterCommand executes a utility-meter action against a TLS test server and returns the request it
// received (nil when the command failed before sending), plus the command error.
func runUtilityMeterCommand(t *testing.T, cliArgs ...string) (*http.Request, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var received *http.Request
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		received = request.Clone(request.Context())
		received.Body = io.NopCloser(bytes.NewReader(body))
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"guid":"11111111-1111-4111-8111-111111111111","isActive":false}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "utility-meter-archive-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	var meters *Domain
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "utility-meter" {
			meters = domain
		}
	}
	if meters == nil {
		t.Fatal("utility-meter domain not registered")
	}
	apiClient := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
	format := "json"
	command := buildDomainCommand(meters, func() (*client.APIClient, error) { return apiClient, nil },
		logging.New(logging.LevelError), &format, &ExportConfig{})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(cliArgs)
	return received, command.Execute()
}

func TestUtilityMeterArchiveAndRestorePostToTheGuidRoutes(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for _, action := range []string{"archive", "restore"} {
		request, err := runUtilityMeterCommand(t, action, guid)
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if request.Method != http.MethodPost || request.URL.Path != "/api/utility-meter/by-guid/"+guid+"/"+action {
			t.Fatalf("unexpected %s route: %s %s", action, request.Method, request.URL)
		}
		if request.URL.RawQuery != "" {
			t.Fatalf("%s leaked arguments into the query: %s", action, request.URL)
		}
	}
}

func TestUtilityMeterArchiveRefusesBadGuidsBeforeSending(t *testing.T) {
	cases := [][]string{
		{"archive", "00000000-0000-0000-0000-000000000000"},
		{"archive", "not-a-guid"},
		{"restore", "11111111-1111-4111-8111-111111111111", "extra"},
		{"restore"},
	}
	for _, args := range cases {
		request, err := runUtilityMeterCommand(t, args...)
		if err == nil || request != nil {
			t.Fatalf("%v must be refused before any request; err=%v request=%v", args, err, request)
		}
	}
}
