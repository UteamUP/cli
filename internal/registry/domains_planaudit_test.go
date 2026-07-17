package registry

import (
	"testing"
)

// --- plan-audit: read-only access to the plan-change audit trail ---

func TestPlanAuditDomainRegistered(t *testing.T) {
	d := findDomain("plan-audit")
	if d == nil {
		t.Fatal("expected plan-audit domain to be registered")
	}
	// PlanAuditController routes under /api/planaudit — a distinct controller
	// from the `plan` domain, so buildRESTPath can only reach it via an
	// explicit APIPath (RESTPath is always appended to the domain base path).
	if d.APIPath != "/api/planaudit" {
		t.Errorf("plan-audit APIPath = %q, want /api/planaudit", d.APIPath)
	}
	expected := map[string]bool{"plan-audits": true, "planaudit": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func assertPlanGUIDArg(t *testing.T, action *Action) {
	t.Helper()
	// The positional arg name must literally be `planGuid` or expandPathTemplate
	// leaves the `{planGuid}` token unresolved in the URL.
	if len(action.Args) != 1 {
		t.Fatalf("%s expected 1 positional arg, got %+v", action.Name, action.Args)
	}
	if action.Args[0].Name != "planGuid" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Errorf("%s arg must be a required string 'planGuid', got %+v", action.Name, action.Args[0])
	}
	if len(action.Flags) != 0 {
		t.Errorf("%s should take no flags, got %d", action.Name, len(action.Flags))
	}
}

func TestPlanAuditActionRouteTemplates(t *testing.T) {
	// Method "" means the runtime derives GET — neither `history` nor `export`
	// is in the HTTPMethod map and neither has an `update-` prefix.
	cases := []struct {
		action   string
		tool     string
		restPath string
	}{
		{"history", "UteamupPlanAuditHistory", "by-plan/{planGuid}"},
		{"export", "UteamupPlanAuditExport", "by-plan/{planGuid}/export"},
	}
	for _, c := range cases {
		a := findDomainAction(t, "plan-audit", c.action)
		if a.ToolName != c.tool || a.HTTPMethod != "" || a.RESTPath != c.restPath {
			t.Errorf("plan-audit %s: want tool=%s method=%q path=%s, got tool=%s method=%q path=%s",
				c.action, c.tool, "", c.restPath, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
		assertPlanGUIDArg(t, a)
	}
}
