package registry

import "testing"

func labourAiMonitorDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "labour-ai-monitor" {
			return domain
		}
	}
	t.Fatal("expected labour-ai-monitor domain")
	return nil
}

func labourAiMonitorAction(t *testing.T, name string) *Action {
	t.Helper()
	domain := labourAiMonitorDomain(t)
	for index := range domain.Actions {
		action := &domain.Actions[index]
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("expected labour-ai-monitor action %q", name)
	return nil
}

func TestLabourAiMonitorDomainMirrorsGuidFirstTools(t *testing.T) {
	domain := labourAiMonitorDomain(t)
	if domain.APIPath != "/api/labour-marketplace/ai/job-monitors" {
		t.Fatalf("APIPath = %q", domain.APIPath)
	}
	want := map[string]string{
		"list":   "UteamupLabourMarketplaceAiJobMonitorsList",
		"cost":   "UteamupLabourMarketplaceAiJobMonitorCost",
		"create": "UteamupLabourMarketplaceAiJobMonitorCreate",
		"update": "UteamupLabourMarketplaceAiJobMonitorUpdate",
		"delete": "UteamupLabourMarketplaceAiJobMonitorDelete",
		"run":    "UteamupLabourMarketplaceAiJobMonitorRun",
	}
	for name, tool := range want {
		if got := labourAiMonitorAction(t, name).ToolName; got != tool {
			t.Errorf("%s ToolName = %q, want %q", name, got, tool)
		}
	}
}

func TestLabourAiMonitorMutationRoutesUseMonitorGuid(t *testing.T) {
	for _, name := range []string{"update", "delete", "run"} {
		action := labourAiMonitorAction(t, name)
		if len(action.Args) != 1 || action.Args[0].Name != "monitorGuid" || action.Args[0].Type != "uuid" {
			t.Errorf("%s must use one GUID argument, got %+v", name, action.Args)
		}
		if action.RESTPath == "" {
			t.Errorf("%s must declare a GUID REST path", name)
		}
	}
	if action := labourAiMonitorAction(t, "run"); action.HTTPMethod != "POST" || action.RESTPath != "{monitorGuid}/run" {
		t.Errorf("run route = %s %s", action.HTTPMethod, action.RESTPath)
	}
}

func TestLabourAiMonitorCreateRequiresOwnedProfileAndExplicitBudgets(t *testing.T) {
	action := labourAiMonitorAction(t, "create")
	required := map[string]bool{"provider-party-guid": false, "name": false}
	budgetFlags := map[string]bool{"max-credits-per-run": false, "monthly-credit-budget": false}
	for _, flag := range action.Flags {
		if _, ok := required[flag.Name]; ok {
			required[flag.Name] = flag.Required
		}
		if _, ok := budgetFlags[flag.Name]; ok {
			budgetFlags[flag.Name] = true
		}
	}
	for name, ok := range required {
		if !ok {
			t.Errorf("--%s must be required", name)
		}
	}
	for name, ok := range budgetFlags {
		if !ok {
			t.Errorf("--%s must be present", name)
		}
	}
}
