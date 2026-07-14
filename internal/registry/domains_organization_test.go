package registry

import (
	"testing"
)

func findCodeDomain(t *testing.T) *Domain {
	t.Helper()
	for _, dom := range DefaultRegistry.Domains() {
		if dom.Name == "code" {
			return dom
		}
	}
	t.Fatal("expected code domain to be registered")
	return nil
}

func TestCodeDomainTargetsPluralRoute(t *testing.T) {
	d := findCodeDomain(t)
	// CodesController routes at api/codes (plural) — the auto-derived
	// "/api/code" base never matched a backend route.
	if d.APIPath != "/api/codes" {
		t.Errorf("code domain APIPath = %q, want %q", d.APIPath, "/api/codes")
	}
	if len(d.Aliases) != 1 || d.Aliases[0] != "codes" {
		t.Errorf("code domain aliases = %+v, want [codes]", d.Aliases)
	}
}

func TestCodeResolveActionWired(t *testing.T) {
	d := findCodeDomain(t)
	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "resolve" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected `resolve` action on code domain")
	}

	if action.ToolName != "UteamupCodeResolve" {
		t.Errorf("resolve ToolName = %q, want %q", action.ToolName, "UteamupCodeResolve")
	}
	// Default GET — the resolver is a read (soft-miss 200, never a 404).
	if action.HTTPMethod != "" {
		t.Errorf("resolve HTTPMethod = %q, want \"\" (defaults to GET)", action.HTTPMethod)
	}
	if action.RESTPath != "resolve/{value}" {
		t.Errorf("resolve RESTPath = %q, want %q (GET api/codes/resolve/{value})", action.RESTPath, "resolve/{value}")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "value" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Fatalf("resolve expected single required string positional arg 'value', got %+v", action.Args)
	}
}

func TestTenantHolidayGuidRoutesResolve(t *testing.T) {
	d := findDomain("tenant-holiday")
	if d == nil {
		t.Fatal("expected tenant-holiday domain to be registered")
	}

	actions := map[string]Action{}
	for _, action := range d.Actions {
		actions[action.Name] = action
	}

	cases := []struct {
		name string
		args map[string]any
		want string
	}{
		{"year", map[string]any{"year": 2026}, "/api/tenantholiday/year/2026"},
		{"update", map[string]any{"holidayGuid": "holiday-1"}, "/api/tenantholiday/by-guid/holiday-1"},
		{"delete", map[string]any{"holidayGuid": "holiday-1"}, "/api/tenantholiday/by-guid/holiday-1"},
		{"import", map[string]any{"year": 2026}, "/api/tenantholiday/import/2026"},
	}

	for _, tc := range cases {
		action, ok := actions[tc.name]
		if !ok {
			t.Fatalf("missing tenant-holiday action %q", tc.name)
		}
		got, consumed := buildRESTPath(d, action, tc.args)
		if got != tc.want {
			t.Fatalf("%s path = %q, want %q", tc.name, got, tc.want)
		}
		if len(consumed) != 1 {
			t.Fatalf("%s consumed = %v, want one path arg", tc.name, consumed)
		}
	}
}
