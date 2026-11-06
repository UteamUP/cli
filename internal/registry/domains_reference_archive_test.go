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

// referenceArchiveDomains are the reference domains whose IsActive flag is their archive, with the REST base
// path each archive and restore route hangs off.
var referenceArchiveDomains = map[string]string{
	"skill":                  "/api/skill",
	"certificate":            "/api/certificate",
	"currency":               "/api/currency",
	"shift":                  "/api/shift",
	"workorder-template":     "/api/workordertemplate",
	"asset-maintenance-plan": "/api/v1/maintenanceplans",
}

// runReferenceCommand executes one action of a domain against a TLS test server and returns the request it
// received (nil when the command failed before sending), plus the command error.
func runReferenceCommand(t *testing.T, domainName string, cliArgs ...string) (*http.Request, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var received *http.Request
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		received = request.Clone(request.Context())
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"type":"skill","guid":"11111111-1111-4111-8111-111111111111","isActive":false}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "reference-archive-token",
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

func TestReferenceArchiveAndRestorePostToTheGuidRoutes(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for domainName, basePath := range referenceArchiveDomains {
		for _, action := range []string{"archive", "restore"} {
			request, err := runReferenceCommand(t, domainName, action, guid)
			if err != nil {
				t.Fatalf("%s %s: %v", domainName, action, err)
			}
			want := basePath + "/by-guid/" + guid + "/" + action
			if request.Method != http.MethodPost || request.URL.Path != want {
				t.Fatalf("%s %s = %s %s, want POST %s", domainName, action, request.Method, request.URL.Path, want)
			}
			if request.URL.RawQuery != "" {
				t.Fatalf("%s %s leaked arguments into the query: %s", domainName, action, request.URL)
			}
		}
	}
}

func TestReferenceArchiveDeclaresOneRequiredNonEmptyGuid(t *testing.T) {
	tools := map[string]string{"archive": "ArchiveReferenceRecord", "restore": "RestoreReferenceRecord"}
	for domainName := range referenceArchiveDomains {
		for actionName, toolName := range tools {
			action := findDomainAction(t, domainName, actionName)
			if action.ToolName != toolName || action.HTTPMethod != "POST" || !action.RejectExtraArgs {
				t.Fatalf("%s %s = tool %q method %q rejectExtra %v", domainName, actionName,
					action.ToolName, action.HTTPMethod, action.RejectExtraArgs)
			}
			if len(action.Args) != 1 || !action.Args[0].Required || action.Args[0].Type != "non-empty-uuid" {
				t.Fatalf("%s %s args = %+v, want one required non-empty-uuid", domainName, actionName, action.Args)
			}
		}
	}
}

func TestReferenceArchiveRefusesBadGuidsBeforeSending(t *testing.T) {
	cases := [][]string{
		{"archive", "00000000-0000-0000-0000-000000000000"},
		{"archive", "not-a-guid"},
		{"restore", "00000000-0000-0000-0000-000000000000"},
		{"archive", "11111111-1111-4111-8111-111111111111", "extra"},
	}
	for domainName := range referenceArchiveDomains {
		for _, args := range cases {
			request, err := runReferenceCommand(t, domainName, args...)
			if err == nil || request != nil {
				t.Fatalf("%s %v must be refused before any request; err=%v request=%v", domainName, args, err, request)
			}
		}
	}
}

func TestReferenceListsHideInactiveUnlessAsked(t *testing.T) {
	for domainName := range referenceArchiveDomains {
		list := findDomainAction(t, domainName, "list")
		flag := findFlag(list, "is-active")
		if flag == nil || flag.QueryName != "isActive" || flag.Type != "bool" || flag.Default != nil || flag.Required {
			t.Fatalf("%s list --is-active = %+v, want optional bool sent as isActive with no default", domainName, flag)
		}
	}

	request, err := runReferenceCommand(t, "currency", "list", "--is-active=false")
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != http.MethodGet || request.URL.Query().Get("isActive") != "false" {
		t.Fatalf("currency list --is-active=false = %s %s, want isActive=false", request.Method, request.URL)
	}

	request, err = runReferenceCommand(t, "currency", "list")
	if err != nil {
		t.Fatal(err)
	}
	if request.URL.Query().Has("isActive") {
		t.Fatalf("currency list must leave isActive out so the server hides inactive records: %s", request.URL)
	}
}
