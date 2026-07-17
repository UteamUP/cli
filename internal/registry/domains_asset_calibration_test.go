package registry

import "testing"

func TestAssetCalibrationOverdueActionMirrorsAssistantSafeMCPRead(t *testing.T) {
	domain := findDomain("asset-calibration")
	if domain == nil {
		t.Fatal("asset-calibration domain is not registered")
	}

	var action *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "overdue" {
			action = &domain.Actions[index]
			break
		}
	}
	if action == nil {
		t.Fatal("asset-calibration overdue action is not registered")
	}
	if action.ToolName != "UteamupAssetcalibrationGetOverdue" ||
		action.HTTPMethod != "GET" || action.RESTPath != "overdue" {
		t.Fatalf(
			"asset-calibration overdue action = tool %q, method %q, path %q",
			action.ToolName,
			action.HTTPMethod,
			action.RESTPath,
		)
	}
	if len(action.Args) != 0 || len(action.Flags) != 0 {
		t.Fatalf("asset-calibration overdue action unexpectedly accepts identifiers: %+v", action)
	}
}

func TestAssetCalibrationDueSoonActionUsesOnlyBoundedPlanningDays(t *testing.T) {
	domain := findDomain("asset-calibration")
	if domain == nil {
		t.Fatal("asset-calibration domain is not registered")
	}

	var action *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "due-soon" {
			action = &domain.Actions[index]
			break
		}
	}
	if action == nil {
		t.Fatal("asset-calibration due-soon action is not registered")
	}
	if action.ToolName != "UteamupAssetcalibrationGetDueSoon" ||
		action.HTTPMethod != "GET" || action.RESTPath != "due-soon" {
		t.Fatalf("unexpected due-soon action: %+v", action)
	}
	if len(action.Args) != 0 || len(action.Flags) != 1 || action.Flags[0].Name != "days" {
		t.Fatalf("due-soon action must accept only the days window: %+v", action)
	}
}
