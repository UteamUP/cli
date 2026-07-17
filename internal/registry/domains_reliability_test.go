package registry

import "testing"

func TestReliabilityRiskUsesGuidFirstEvidenceRoute(t *testing.T) {
	domain := findDomain("reliability")
	if domain == nil {
		t.Fatal("expected reliability domain")
	}
	if domain.APIPath != "/api/analytics/reliability" {
		t.Fatalf("API path = %q", domain.APIPath)
	}
	if len(domain.Actions) != 2 {
		t.Fatalf("actions = %d, want 2", len(domain.Actions))
	}

	action := domain.Actions[0]
	if action.Name != "risk" ||
		action.ToolName != "UteamupReliabilityRiskGet" ||
		action.HTTPMethod != "GET" ||
		action.RESTPath != "risks" {
		t.Fatalf("risk action = %+v", action)
	}
	path, consumed := buildRESTPath(domain, action, map[string]any{})
	if path != "/api/analytics/reliability/risks" {
		t.Fatalf("path = %q", path)
	}
	if len(consumed) != 0 {
		t.Fatalf("unexpected consumed args: %v", consumed)
	}

	flags := make(map[string]FlagDef, len(action.Flags))
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	if flags["asset-guid"].Type != "string" {
		t.Fatalf("asset-guid flag = %+v", flags["asset-guid"])
	}
	for _, forbidden := range []string{"id", "asset-id"} {
		if _, exists := flags[forbidden]; exists {
			t.Fatalf("integer-style identity flag %q must not be exposed", forbidden)
		}
	}
	if flags["limit"].Default != 20 || flags["limit"].Type != "int" {
		t.Fatalf("limit flag = %+v", flags["limit"])
	}
}

func TestReliabilityStrategyUsesReviewOnlyGuidFirstProposalRoute(t *testing.T) {
	domain := findDomain("reliability")
	if domain == nil {
		t.Fatal("expected reliability domain")
	}

	var action *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "strategy" {
			action = &domain.Actions[index]
			break
		}
	}
	if action == nil {
		t.Fatal("expected strategy action")
	}
	if action.ToolName != "UteamupReliabilityStrategyPropose" ||
		action.HTTPMethod != "POST" ||
		action.RESTPath != "strategies/propose" {
		t.Fatalf("strategy action = %+v", action)
	}

	flags := make(map[string]FlagDef, len(action.Flags))
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	if !flags["asset-guid"].Required || flags["asset-guid"].BodyName != "assetGuid" {
		t.Fatalf("asset-guid flag = %+v", flags["asset-guid"])
	}
	if flags["objective"].Default != "availability" ||
		flags["objective"].BodyName != "objectiveKey" {
		t.Fatalf("objective flag = %+v", flags["objective"])
	}
	for _, forbidden := range []string{"id", "asset-id"} {
		if _, exists := flags[forbidden]; exists {
			t.Fatalf("integer-style identity flag %q must not be exposed", forbidden)
		}
	}
}
