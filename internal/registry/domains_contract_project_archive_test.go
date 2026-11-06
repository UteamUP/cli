package registry

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

// runContractProjectCommand executes a contract or project action against a TLS test server and returns the
// request it received (nil when the command failed before sending), plus the command error.
func runContractProjectCommand(t *testing.T, domainName string, cliArgs ...string) (*http.Request, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var received *http.Request
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		received = request.Clone(request.Context())
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"guid":"11111111-1111-4111-8111-111111111111","status":5}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "contract-project-archive-token",
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

func TestContractAndProjectArchiveAndRestorePostToTheGuidRoutes(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	basePaths := map[string]string{"contract": "/api/contracts", "project": "/api/project"}
	for domainName, basePath := range basePaths {
		for _, action := range []string{"archive", "restore"} {
			request, err := runContractProjectCommand(t, domainName, action, guid)
			if err != nil {
				t.Fatalf("%s %s: %v", domainName, action, err)
			}
			want := basePath + "/by-guid/" + guid + "/" + action
			if request.Method != http.MethodPost || request.URL.Path != want {
				t.Fatalf("%s %s: unexpected route %s %s, want POST %s", domainName, action, request.Method, request.URL, want)
			}
			if request.URL.RawQuery != "" {
				t.Fatalf("%s %s leaked arguments into the query: %s", domainName, action, request.URL)
			}
		}
	}
}

func TestContractAndProjectArchiveRefuseBadGuidsBeforeSending(t *testing.T) {
	for _, domainName := range []string{"contract", "project"} {
		for _, args := range [][]string{
			{"archive", "00000000-0000-0000-0000-000000000000"},
			{"restore", "not-a-guid"},
			{"archive"},
		} {
			request, err := runContractProjectCommand(t, domainName, args...)
			if err == nil || request != nil {
				t.Fatalf("%s %v must be refused before any request; err=%v request=%v", domainName, args, err, request)
			}
		}
	}
}
