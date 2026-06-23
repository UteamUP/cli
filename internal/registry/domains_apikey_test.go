package registry

import "testing"

func findApiKeyDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "apikey" {
			return d
		}
	}
	t.Fatal("expected apikey domain to be registered")
	return nil
}

func findApiKeyAction(t *testing.T, name string) *Action {
	t.Helper()
	d := findApiKeyDomain(t)
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected %q action on apikey domain", name)
	return nil
}

func TestApiKeyDomainRegistered(t *testing.T) {
	d := findApiKeyDomain(t)
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
	d := findApiKeyDomain(t)
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
	action := findApiKeyAction(t, "create")

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

	// Flags map onto CreateTenantApiKeyModel via camelCase auto-conversion.
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

// list → GET /api/tenant-api-keys (no path id, no RESTPath).
func TestApiKeyListAction(t *testing.T) {
	action := findApiKeyAction(t, "list")

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
	action := findApiKeyAction(t, "get")

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
	action := findApiKeyAction(t, "revoke")

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
