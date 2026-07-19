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
	if len(domain.Actions) != 6 {
		t.Fatalf("actions = %d, want 6", len(domain.Actions))
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

func TestReliabilityEvidenceAndRunActionsStayGuidFirst(t *testing.T) {
	domain := findDomain("reliability")
	if domain == nil {
		t.Fatal("expected reliability domain")
	}

	expected := map[string]struct {
		tool   string
		method string
		path   string
	}{
		"assessments":       {"UteamupReliabilityAssessmentsList", "GET", "assessments"},
		"assessment-create": {"UteamupReliabilityAssessmentCreate", "POST", "assessments"},
		"assessment-approve": {
			"UteamupReliabilityAssessmentApprove",
			"POST",
			"assessments/{versionGuid}/approve",
		},
		"prepare-run": {
			"UteamupReliabilityStrategyPrepareRun",
			"POST",
			"strategies/prepare-run",
		},
	}

	for name, want := range expected {
		var action *Action
		for index := range domain.Actions {
			if domain.Actions[index].Name == name {
				action = &domain.Actions[index]
				break
			}
		}
		if action == nil {
			t.Fatalf("expected action %q", name)
		}
		if action.ToolName != want.tool ||
			action.HTTPMethod != want.method ||
			action.RESTPath != want.path {
			t.Fatalf("action %q = %+v", name, action)
		}
		for _, arg := range action.Args {
			if arg.Type != "uuid" {
				t.Fatalf("action %q identity arg = %+v", name, arg)
			}
		}
		for _, flag := range action.Flags {
			if flag.Name == "asset-guid" ||
				flag.Name == "request-guid" ||
				flag.Name == "run-request-guid" ||
				flag.Name == "failure-code-guid" {
				if flag.Type != "uuid" {
					t.Fatalf("action %q GUID flag = %+v", name, flag)
				}
			}
			if flag.Name == "id" || flag.Name == "asset-id" {
				t.Fatalf("action %q exposes integer-style identity flag", name)
			}
		}
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
