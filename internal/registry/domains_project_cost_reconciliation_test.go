package registry

import "testing"

func TestManualProjectCostCreateKeepsCompleteRetryPayload(t *testing.T) {
	action := findDomainAction(t, "project-cost", "create")
	if action.ToolName != "UteamupProjectCostRecordCreate" || action.HTTPMethod != "POST" || action.RESTPath != "{projectGuid}/cost-records" {
		t.Fatalf("manual cost creation must use its existing backend route and tool: %+v", action)
	}
	if action.Args[0].Name != "projectGuid" || len(action.Flags) != 1 || action.Flags[0].Name != "json" {
		t.Fatal("manual cost creation must preserve project GUID and the complete payload with stable requestGuid")
	}
}

func TestProjectCostReconciliationRoutesPreserveEvidence(t *testing.T) {
	for _, scenario := range []struct{ action, method, path, tool string }{
		{"preview", "", "{projectGuid}/cost-records/reconciliation", "UteamupProjectCostReconciliation"},
		{"history", "", "{projectGuid}/cost-records/reconciliation/history", "UteamupProjectCostReconciliationHistory"},
		{"review", "POST", "{projectGuid}/cost-records/reconciliation", "UteamupProjectCostReconcile"},
	} {
		action := findDomainAction(t, "project-cost-reconciliation", scenario.action)
		if action.HTTPMethod != scenario.method || action.RESTPath != scenario.path || action.ToolName != scenario.tool {
			t.Errorf("%s route must preserve the REST/MCP reconciliation contract: %+v", scenario.action, action)
		}
		if action.Args[0].Name != "projectGuid" || !action.Args[0].Required {
			t.Errorf("%s requires the project GUID", scenario.action)
		}
	}
	history := findDomainAction(t, "project-cost-reconciliation", "history")
	if history.Args[1].QueryName != "sourceGuid" || !history.Args[1].Required {
		t.Fatal("history must scope the query to the exact stock source GUID")
	}
	review := findDomainAction(t, "project-cost-reconciliation", "review")
	if len(review.Flags) != 1 || review.Flags[0].Name != "json" {
		t.Fatal("review must accept the complete reviewed payload, including stable request identity and fingerprints")
	}
}

func TestProjectDeliverableCostPlanRoutePreservesFullPayload(t *testing.T) {
	action := findDomainAction(t, "project-deliverable-cost", "plan")
	if action.HTTPMethod != "PUT" || action.RESTPath != "{projectGuid}/outputitems/{deliverableGuid}/costs/plan" ||
		action.ToolName != "UteamupProjectDeliverableCostPlanUpdate" {
		t.Fatalf("cost-plan action has mismatched REST/MCP wiring: %+v", action)
	}
	if action.Args[1].Name != "deliverableGuid" || len(action.Flags) != 1 || action.Flags[0].Name != "json" {
		t.Fatal("cost-plan updates require exact deliverable identity and the complete concurrency-aware payload")
	}
}

func TestProjectDeliverableCostPlanReadIsDistinctFromSubtreeSummary(t *testing.T) {
	action := findDomainAction(t, "project-deliverable-cost", "get-plan")
	if action.HTTPMethod != "" || action.RESTPath != "{projectGuid}/outputitems/{deliverableGuid}/costs/plan" ||
		action.ToolName != "UteamupProjectDeliverableCostPlanGet" || action.Args[1].Name != "deliverableGuid" {
		t.Fatalf("direct cost-plan read must preserve its project, deliverable and read-only REST/MCP contract: %+v", action)
	}
}
