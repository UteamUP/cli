package registry

import "testing"

func TestOperationalRouteActionsUseGuidIdentifiers(t *testing.T) {
	domain := findDomain("route")
	if domain == nil {
		t.Fatal("route domain is not registered")
	}

	expectedTools := map[string]string{
		"list":          "UteamupOperationalRouteList",
		"get":           "UteamupOperationalRouteGet",
		"schedules":     "UteamupInspectionScheduleList",
		"overdue":       "UteamupInspectionScheduleGetOverdue",
		"executions":    "UteamupInspectionExecutionList",
		"execution":     "UteamupInspectionExecutionGet",
		"start":         "UteamupInspectionExecutionStart",
		"complete-stop": "UteamupInspectionStopComplete",
		"flag-issue":    "UteamupInspectionIssueflag",
		"complete":      "UteamupInspectionExecutionComplete",
		"abandon":       "UteamupInspectionExecutionAbandon",
		"analytics":     "UteamupInspectionAnalyticsOverview",
		"anomalies":     "UteamupInspectionAnomalyList",
		"asset-health":  "UteamupInspectionAssetHealthscore",
		"optimize":      "UteamupOperationalRouteOptimize",
	}

	if len(domain.Actions) != len(expectedTools) {
		t.Fatalf("route actions = %d, want %d: %+v", len(domain.Actions), len(expectedTools), domain.Actions)
	}

	for _, action := range domain.Actions {
		wantTool, ok := expectedTools[action.Name]
		if !ok {
			t.Fatalf("unexpected route action %q", action.Name)
		}
		if action.ToolName != wantTool {
			t.Fatalf("%s tool = %q, want %q", action.Name, action.ToolName, wantTool)
		}
		for _, arg := range action.Args {
			if arg.Type == "int" || arg.Name == "id" || arg.Name == "routeId" ||
				arg.Name == "executionId" || arg.Name == "assetId" {
				t.Fatalf("route action %s leaks integer identifiers: %+v", action.Name, arg)
			}
		}
		for _, flag := range action.Flags {
			if flag.BodyName == "routeId" || flag.Name == "route-id" {
				t.Fatalf("route action %s leaks routeId: %+v", action.Name, flag)
			}
		}
	}
}
