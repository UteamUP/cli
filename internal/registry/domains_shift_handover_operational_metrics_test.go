package registry

import "testing"

func TestShiftHandoverOperationalMetricsMirrorsMcpAndRest(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	action := findShiftHandoverAction(t, domain, "operational-metrics")

	if action.ToolName != "UteamupShiftHandoverGetStats" {
		t.Errorf(
			"ToolName = %q, want UteamupShiftHandoverGetStats",
			action.ToolName,
		)
	}
	if action.HTTPMethod != "GET" || action.RESTPath != "stats" {
		t.Errorf(
			"route = %s %s, want GET stats",
			action.HTTPMethod,
			action.RESTPath,
		)
	}
	if len(action.Args) != 0 {
		t.Errorf("args = %+v, want no positional identities", action.Args)
	}

	flags := map[string]FlagDef{}
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	for cliName, queryName := range map[string]string{
		"from-date": "fromDate",
		"to-date":   "toDate",
	} {
		flag, ok := flags[cliName]
		if !ok {
			t.Fatalf("missing optional --%s reporting-window flag", cliName)
		}
		if flag.Required || flag.Default != nil || flag.Type != "string" {
			t.Errorf("flag --%s must be an optional string without a default: %+v", cliName, flag)
		}
		if flag.BodyName != queryName {
			t.Errorf("flag --%s maps to %q, want %q", cliName, flag.BodyName, queryName)
		}
	}
}

func TestShiftHandoverOperationalMetricsPathHasNoIdentity(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	action := findShiftHandoverAction(t, domain, "operational-metrics")

	path, consumed := buildRESTPath(domain, *action, map[string]any{
		"fromDate": "2026-06-13T12:00:00Z",
		"toDate":   "2026-07-13T12:00:00Z",
	})

	if path != "/api/shifthandover/stats" {
		t.Errorf("path = %q, want /api/shifthandover/stats", path)
	}
	if len(consumed) != 0 {
		t.Errorf("consumed = %v, want no reporting-window query fields consumed", consumed)
	}
}
