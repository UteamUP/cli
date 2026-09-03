package registry

import "testing"

// The workorder-share domain mirrors the backend WorkorderShareController
// (api/workorder/shares/*) and the three MCP share tools. Its routes and tool
// names are the backend contract; a typo here ships a command that always 404s.

func workorderShareDomain(t *testing.T) *Domain {
	t.Helper()
	d := findDomain("workorder-share")
	if d == nil {
		t.Fatal("expected workorder-share domain to be registered")
	}
	return d
}

func TestWorkorderShareDomainRoutesUnderWorkorderShares(t *testing.T) {
	d := workorderShareDomain(t)
	if d.APIPath != "/api/workorder/shares" {
		t.Errorf("APIPath = %q, want /api/workorder/shares", d.APIPath)
	}
	hasAlias := false
	for _, alias := range d.Aliases {
		if alias == "wos" {
			hasAlias = true
		}
	}
	if !hasAlias {
		t.Errorf("expected the wos alias, got %v", d.Aliases)
	}
}

func TestWorkorderShareActionsMatchBackendContract(t *testing.T) {
	d := workorderShareDomain(t)
	cases := []struct {
		name, tool, method, path string
	}{
		{"list", "UteamupWorkorderShareList", "", "by-workorder/{workorderGuid}"},
		{"create", "UteamupWorkorderShareCreate", "", ""},
		{"revoke", "UteamupWorkorderShareRevoke", "DELETE", "{shareGuid}"},
	}
	for _, c := range cases {
		a := findAction(d, c.name)
		if a == nil {
			t.Errorf("expected %s action on workorder-share", c.name)
			continue
		}
		if a.ToolName != c.tool {
			t.Errorf("%s: tool = %q, want %q", c.name, a.ToolName, c.tool)
		}
		if a.HTTPMethod != c.method || a.RESTPath != c.path {
			t.Errorf("%s: route = %s %q, want %s %q", c.name, a.HTTPMethod, a.RESTPath, c.method, c.path)
		}
	}
	if len(d.Actions) != len(cases) {
		t.Errorf("workorder-share has %d actions, want %d (update is UI-only)", len(d.Actions), len(cases))
	}
}

// GUIDs in, never integer ids: positional args are uuid-typed and the create
// body addresses the workorder by GUID with read-only as the default access.
func TestWorkorderShareIsGuidFirstAndReadOnlyByDefault(t *testing.T) {
	d := workorderShareDomain(t)
	for _, name := range []string{"list", "revoke"} {
		a := findAction(d, name)
		if a == nil || len(a.Args) != 1 || a.Args[0].Type != "uuid" || !a.Args[0].Required {
			t.Errorf("%s: expected exactly one required uuid positional arg, got %+v", name, a)
		}
	}
	create := findAction(d, "create")
	if create == nil {
		t.Fatal("expected create action")
	}
	flags := map[string]FlagDef{}
	for _, f := range create.Flags {
		flags[f.Name] = f
	}
	wo, ok := flags["workorder-guid"]
	if !ok || wo.Type != "uuid" || !wo.Required || wo.BodyName != "workorderGuid" {
		t.Errorf("create: workorder-guid flag = %+v, want required uuid bound to workorderGuid", wo)
	}
	access, ok := flags["access-level"]
	if !ok || access.Default != "readOnly" || access.BodyName != "accessLevel" {
		t.Errorf("create: access-level flag = %+v, want default readOnly bound to accessLevel", access)
	}
	for name, f := range flags {
		if f.Type == "int" && name != "expires-in-days" {
			t.Errorf("create: integer flag %q looks like an id at the boundary", name)
		}
	}
}
