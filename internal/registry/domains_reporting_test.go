package registry

import "testing"

func TestApprovedReportAnalyticsReadWired(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "report-analytics" {
			domain = candidate
			break
		}
	}
	if domain == nil {
		t.Fatal("expected report-analytics domain")
	}
	if domain.APIPath != "/api/report" {
		t.Fatalf("APIPath = %q, want /api/report", domain.APIPath)
	}
	if len(domain.Actions) != 1 {
		t.Fatalf("actions = %d, want one bounded read", len(domain.Actions))
	}

	action := domain.Actions[0]
	if action.Name != "read" || action.ToolName != "UteamupReportAnalytics" {
		t.Errorf("action = %q/%q, want read/UteamupReportAnalytics", action.Name, action.ToolName)
	}
	if action.HTTPMethod != "GET" || action.RESTPath != "analytics" {
		t.Errorf("route = %s %s, want GET analytics", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 0 {
		t.Errorf("read must not expose positional identifiers, got %+v", action.Args)
	}

	flags := make(map[string]FlagDef, len(action.Flags))
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	for _, name := range []string{"start-date", "end-date"} {
		if flag, ok := flags[name]; !ok || !flag.Required || flag.Type != "string" {
			t.Errorf("%s = %+v, want required string", name, flag)
		}
	}
	if flag, ok := flags["group-by"]; !ok || flag.Default != "month" || flag.Type != "string" {
		t.Errorf("group-by = %+v, want optional string default month", flag)
	}
}

func TestCostOverviewWorkorderGetUsesGuidContract(t *testing.T) {
	action := findDomainAction(t, "cost-overview", "get")

	if action.ToolName != "UteamupCostByWorkorder" {
		t.Fatalf("ToolName = %q, want UteamupCostByWorkorder", action.ToolName)
	}
	if action.HTTPMethod != "GET" || action.RESTPath != "workorders/by-guid/{workorderGuid}" {
		t.Fatalf("route = %s %s, want GET workorders/by-guid/{workorderGuid}", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 1 {
		t.Fatalf("args = %+v, want one GUID arg", action.Args)
	}
	arg := action.Args[0]
	if arg.Name != "workorderGuid" || arg.Type != "uuid" || !arg.Required {
		t.Fatalf("arg = %+v, want required workorderGuid uuid", arg)
	}
}
