package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func findAPIKeyDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "apikey" {
			return d
		}
	}
	t.Fatal("expected apikey domain to be registered")
	return nil
}

func findAPIKeyAction(t *testing.T, name string) *Action {
	t.Helper()
	d := findAPIKeyDomain(t)
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected %q action on apikey domain", name)
	return nil
}

func TestApiKeyDomainRegistered(t *testing.T) {
	d := findAPIKeyDomain(t)
	if d.Description == "" {
		t.Error("apikey domain must have a Description")
	}
	// APIPath must be explicit — the registry strips hyphens when deriving a base
	// path from the domain Name, so "apikey" would wrongly become "/api/apikey".
	if d.APIPath != "/api/tenant-api-keys" {
		t.Errorf("apikey APIPath = %q, want %q", d.APIPath, "/api/tenant-api-keys")
	}
	// Aliases let users type `apikeys` / `api-key`.
	wantAliases := map[string]bool{"apikeys": true, "api-key": true}
	for _, a := range d.Aliases {
		delete(wantAliases, a)
	}
	if len(wantAliases) != 0 {
		t.Errorf("apikey domain missing aliases: %v (got %v)", wantAliases, d.Aliases)
	}
}

func TestApiKeyActionsWired(t *testing.T) {
	d := findAPIKeyDomain(t)
	expected := map[string]string{
		"create": "UteamupTenantApiKeyCreate",
		"list":   "UteamupTenantApiKeyList",
		"get":    "UteamupTenantApiKeyGet",
		"revoke": "UteamupTenantApiKeyRevoke",
	}
	got := map[string]string{}
	for _, a := range d.Actions {
		got[a.Name] = a.ToolName
	}
	if len(d.Actions) != len(expected) {
		t.Errorf("apikey domain action count = %d, want %d (%v)", len(d.Actions), len(expected), got)
	}
	for action, tool := range expected {
		if got[action] != tool {
			t.Errorf("expected apikey action %q to map to %q, got %q", action, tool, got[action])
		}
	}
}

// create → POST /api/tenant-api-keys (no path id). Verb falls out of the
// action-name default map; no RESTPath/HTTPMethod override needed.
func TestApiKeyCreateAction(t *testing.T) {
	action := findAPIKeyAction(t, "create")

	if action.RESTPath != "" {
		t.Errorf("create RESTPath = %q, want empty (POST to base path)", action.RESTPath)
	}
	if action.HTTPMethod != "" {
		t.Errorf("create HTTPMethod = %q, want empty (defaults to POST via action-name map)", action.HTTPMethod)
	}
	if HTTPMethod["create"] != "POST" {
		t.Errorf("create resolves to %q, want POST", HTTPMethod["create"])
	}
	if len(action.Args) != 0 {
		t.Errorf("create must take no positional args, got %+v", action.Args)
	}

	flagByName := func(name string) *FlagDef {
		for i := range action.Flags {
			if action.Flags[i].Name == name {
				return &action.Flags[i]
			}
		}
		return nil
	}

	// Flags map onto CreateTenantApiKeyModel, with explicit names where needed.
	for _, want := range []struct {
		name     string
		typ      string
		required bool
	}{
		{"name", "string", true},
		{"description", "string", false},
		{"role-id", "string", false},
		{"mcp-enabled", "bool", false},
		{"expires-at", "string", false},
		{"requests-per-minute", "int", false},
		{"requests-per-hour", "int", false},
		{"requests-per-day", "int", false},
		{"allowed-ip-addresses", "stringSlice", false},
	} {
		f := flagByName(want.name)
		if f == nil {
			t.Fatalf("create must expose a %q flag", want.name)
		}
		if f.Type != want.typ {
			t.Errorf("%s flag type = %q, want %q", want.name, f.Type, want.typ)
		}
		if f.Required != want.required {
			t.Errorf("%s flag Required = %v, want %v", want.name, f.Required, want.required)
		}
	}
}

func TestApiKeyCreateSendsMcpEnabledContract(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "default disabled", true: "explicitly enabled"}[enabled], func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("UTEAMUP_TENANT_GUID", "52e73fd7-cd7b-45c6-9872-b6be9415203c")
			var body map[string]any
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/api/tenant-api-keys" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "test-token", AuthMethod: "apikey", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			logger := logging.New(logging.LevelError)
			format := "json"
			factory := func() (*client.APIClient, error) {
				return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{}, logger), nil
			}
			command := buildActionCommand(findAPIKeyDomain(t), *findAPIKeyAction(t, "create"), factory, logger, &format, nil)
			args := []string{"--name", "Private pilot", "--role-id", "dc2327fc-091c-4e45-ab37-a26768432afe"}
			if enabled {
				args = append(args, "--mcp-enabled")
			}
			command.SetArgs(args)
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if body["isMcpEnabled"] != enabled {
				t.Errorf("isMcpEnabled = %v, want %v", body["isMcpEnabled"], enabled)
			}
			if _, legacy := body["mcpEnabled"]; legacy {
				t.Error("request sent the unsupported mcpEnabled field")
			}
			if body["roleId"] != "dc2327fc-091c-4e45-ab37-a26768432afe" {
				t.Errorf("role scope changed: %v", body["roleId"])
			}
		})
	}
}

// list → GET /api/tenant-api-keys (no path id, no RESTPath).
func TestApiKeyListAction(t *testing.T) {
	action := findAPIKeyAction(t, "list")

	if action.RESTPath != "" {
		t.Errorf("list RESTPath = %q, want empty (GET base path)", action.RESTPath)
	}
	if action.HTTPMethod != "" {
		t.Errorf("list HTTPMethod = %q, want empty (defaults to GET)", action.HTTPMethod)
	}
	if HTTPMethod["list"] != "GET" {
		t.Errorf("list resolves to %q, want GET", HTTPMethod["list"])
	}
	if len(action.Args) != 0 {
		t.Errorf("list must take no positional args, got %+v", action.Args)
	}
}

// get → GET /api/tenant-api-keys/by-guid/{guid}. Needs an explicit RESTPath so
// the GUID lands under by-guid/ (default get routing would emit {base}/{guid}).
func TestApiKeyGetAction(t *testing.T) {
	action := findAPIKeyAction(t, "get")

	if action.RESTPath != "by-guid/{guid}" {
		t.Errorf("get RESTPath = %q, want %q", action.RESTPath, "by-guid/{guid}")
	}
	if HTTPMethod["get"] != "GET" {
		t.Errorf("get resolves to %q, want GET", HTTPMethod["get"])
	}
	if len(action.Args) != 1 || action.Args[0].Name != "guid" {
		t.Fatalf("get expected single positional arg 'guid', got %+v", action.Args)
	}
	if action.Args[0].Type != "string" {
		t.Errorf("guid arg type = %q, want string (Guids are strings)", action.Args[0].Type)
	}
	if !action.Args[0].Required {
		t.Error("guid arg must be Required")
	}
}

// revoke → POST /api/tenant-api-keys/by-guid/{guid}/revoke. POST is not the
// name-default for an unknown verb, so HTTPMethod must be set explicitly.
func TestApiKeyRevokeAction(t *testing.T) {
	action := findAPIKeyAction(t, "revoke")

	if action.HTTPMethod != "POST" {
		t.Errorf("revoke HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "by-guid/{guid}/revoke" {
		t.Errorf("revoke RESTPath = %q, want %q", action.RESTPath, "by-guid/{guid}/revoke")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "guid" {
		t.Fatalf("revoke expected single positional arg 'guid', got %+v", action.Args)
	}
	if action.Args[0].Type != "string" {
		t.Errorf("guid arg type = %q, want string (Guids are strings)", action.Args[0].Type)
	}
	if !action.Args[0].Required {
		t.Error("guid arg must be Required")
	}
}
