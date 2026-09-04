package registry

import "testing"

// The project-share domain mirrors the backend ProjectShareController
// (api/project/shares/*) and the three MCP share tools. Its routes and tool
// names are the backend contract; a typo here ships a command that always 404s.

func projectShareDomain(t *testing.T) *Domain {
	t.Helper()
	d := findDomain("project-share")
	if d == nil {
		t.Fatal("expected project-share domain to be registered")
	}
	return d
}

func TestProjectShareDomainRoutesUnderProjectShares(t *testing.T) {
	d := projectShareDomain(t)
	if d.APIPath != "/api/project/shares" {
		t.Errorf("APIPath = %q, want /api/project/shares", d.APIPath)
	}
	hasAlias := false
	for _, alias := range d.Aliases {
		if alias == "ps" {
			hasAlias = true
		}
	}
	if !hasAlias {
		t.Errorf("expected the ps alias, got %v", d.Aliases)
	}
}

func TestProjectShareActionsMatchBackendContract(t *testing.T) {
	d := projectShareDomain(t)
	cases := []struct {
		name, tool, method, path string
	}{
		{"list", "UteamupProjectShareList", "", "by-project/{projectGuid}"},
		{"create", "UteamupProjectShareCreate", "", ""},
		{"revoke", "UteamupProjectShareRevoke", "DELETE", "{shareGuid}"},
	}
	for _, c := range cases {
		a := findAction(d, c.name)
		if a == nil {
			t.Errorf("expected %s action on project-share", c.name)
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
		t.Errorf("project-share has %d actions, want %d (update is UI-only)", len(d.Actions), len(cases))
	}
}

// The scope grid is what makes a project share different from a workorder share.
// Losing --section would silently downgrade every CLI-created share to the default
// grid, which is a different (and broader) grant than the operator asked for.
func TestProjectShareCreateCarriesTheScopeFlags(t *testing.T) {
	d := projectShareDomain(t)
	create := findAction(d, "create")
	if create == nil {
		t.Fatal("expected create action on project-share")
	}

	want := map[string]struct {
		bodyName string
		typ      string
		required bool
	}{
		"project-guid":           {"projectGuid", "uuid", true},
		"email":                  {"", "string", false},
		"access-level":           {"accessLevel", "string", false},
		"workorder-access-level": {"workorderAccessLevel", "string", false},
		"section":                {"sections", "stringSlice", false},
		"expires-in-days":        {"expiresInDays", "int", false},
		"note":                   {"", "string", false},
	}

	got := make(map[string]bool, len(create.Flags))
	for _, f := range create.Flags {
		got[f.Name] = true
		w, ok := want[f.Name]
		if !ok {
			t.Errorf("unexpected flag %q on project-share create", f.Name)
			continue
		}
		if w.bodyName != "" && f.BodyName != w.bodyName {
			t.Errorf("%s: BodyName = %q, want %q", f.Name, f.BodyName, w.bodyName)
		}
		if f.Type != w.typ {
			t.Errorf("%s: Type = %q, want %q", f.Name, f.Type, w.typ)
		}
		if f.Required != w.required {
			t.Errorf("%s: Required = %v, want %v", f.Name, f.Required, w.required)
		}
	}

	for name := range want {
		if !got[name] {
			t.Errorf("missing flag %q on project-share create", name)
		}
	}
}

// Read is the safe default everywhere else in the product; the CLI must not be the
// one surface that hands out more than the operator typed.
func TestProjectShareCreateDefaultsToTheSafeGrant(t *testing.T) {
	d := projectShareDomain(t)
	create := findAction(d, "create")
	if create == nil {
		t.Fatal("expected create action on project-share")
	}

	for _, f := range create.Flags {
		switch f.Name {
		case "access-level":
			if f.Default != "read" {
				t.Errorf("access-level default = %v, want read", f.Default)
			}
		case "workorder-access-level":
			if f.Default != "none" {
				t.Errorf("workorder-access-level default = %v, want none", f.Default)
			}
		}
	}
}
