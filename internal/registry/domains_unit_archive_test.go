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

// runUnitArchiveCommand executes an action of the named domain against a TLS test server and returns the
// request it received (nil when the command failed before sending), plus the command error.
func runUnitArchiveCommand(t *testing.T, domainName string, cliArgs ...string) (*http.Request, error) {
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
		_, _ = response.Write([]byte(`{"type":"customer","guid":"11111111-1111-4111-8111-111111111111","archived":true}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "unit-archive-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	domain := findDomain(domainName)
	if domain == nil {
		t.Fatalf("%s domain not registered", domainName)
	}
	apiClient := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
	format := "json"
	command := buildDomainCommand(domain, func() (*client.APIClient, error) { return apiClient, nil },
		logging.New(logging.LevelError), &format, &ExportConfig{})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(cliArgs)
	return received, command.Execute()
}

func TestUnitArchiveAndRestorePostToTheGuidRoutes(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for _, domain := range []string{"customer", "vendor", "team", "location"} {
		for _, action := range []string{"archive", "restore"} {
			request, err := runUnitArchiveCommand(t, domain, action, guid)
			if err != nil {
				t.Fatalf("%s %s: %v", domain, action, err)
			}
			want := "/api/" + domain + "/by-guid/" + guid + "/" + action
			if request.Method != http.MethodPost || request.URL.Path != want {
				t.Fatalf("%s %s: got %s %s, want POST %s", domain, action, request.Method, request.URL.Path, want)
			}
			if request.URL.RawQuery != "" {
				t.Fatalf("%s %s leaked arguments into the query: %s", domain, action, request.URL)
			}
		}
	}
}

func TestUnitArchiveRefusesBadGuidsAndExtraArgsBeforeSending(t *testing.T) {
	cases := [][]string{
		{"customer", "archive", "00000000-0000-0000-0000-000000000000"},
		{"vendor", "restore", "not-a-guid"},
		{"team", "archive"},
		{"location", "restore", "11111111-1111-4111-8111-111111111111", "extra"},
	}
	for _, args := range cases {
		request, err := runUnitArchiveCommand(t, args[0], args[1:]...)
		if err == nil || request != nil {
			t.Fatalf("%v must be refused before any request; err=%v request=%v", args, err, request)
		}
	}
}

func TestUnitArchiveActionsNameTheirMcpToolsAndPermission(t *testing.T) {
	tools := map[string]string{"customer": "UteamupCustomer", "vendor": "UteamupVendor", "team": "UteamupTeam", "location": "UteamupLocation"}
	permissions := map[string]string{"customer": "Customer.Delete", "vendor": "Vendor.Delete", "team": "Team.Delete", "location": "Location.Delete"}
	for domainName, prefix := range tools {
		domain := findDomain(domainName)
		for _, action := range []string{"archive", "restore"} {
			found := actionByName(t, domain, action)
			if found.ToolName != prefix+map[string]string{"archive": "Archive", "restore": "Restore"}[action] {
				t.Errorf("%s %s tool = %q", domainName, action, found.ToolName)
			}
			if !bytes.Contains([]byte(found.Description), []byte(permissions[domainName])) {
				t.Errorf("%s %s description must name %s: %q", domainName, action, permissions[domainName], found.Description)
			}
			assertNoIntegerIdentityArgs(t, domainName, found)
		}
	}
}
