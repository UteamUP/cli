package registry

import (
	"strings"
	"testing"
)

func findAssetCalibrationAction(t *testing.T, name string) (*Domain, *Action) {
	t.Helper()
	domain := findDomain("asset-calibration")
	if domain == nil {
		t.Fatal("asset-calibration domain is not registered")
	}

	for index := range domain.Actions {
		if domain.Actions[index].Name == name {
			return domain, &domain.Actions[index]
		}
	}

	t.Fatalf("asset-calibration %s action is not registered", name)
	return nil, nil
}

func TestAssetCalibrationOverdueActionMirrorsAssistantSafeMCPRead(t *testing.T) {
	_, action := findAssetCalibrationAction(t, "overdue")
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
	_, action := findAssetCalibrationAction(t, "due-soon")
	if action.ToolName != "UteamupAssetcalibrationGetDueSoon" ||
		action.HTTPMethod != "GET" || action.RESTPath != "due-soon" {
		t.Fatalf("unexpected due-soon action: %+v", action)
	}
	if len(action.Args) != 0 || len(action.Flags) != 1 || action.Flags[0].Name != "days" {
		t.Fatalf("due-soon action must accept only the days window: %+v", action)
	}
}

func TestAssetCalibrationCRUDUsesOnlyPublicGuidSelectors(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		method   string
		path     string
		argName  string
	}{
		{
			name:     "list",
			toolName: "UteamupAssetcalibrationGetByAsset",
			method:   "GET",
			path:     "asset/by-guid/{assetGuid}",
			argName:  "assetGuid",
		},
		{
			name:     "get",
			toolName: "UteamupAssetcalibrationGet",
			method:   "GET",
			path:     "by-guid/{calibrationGuid}",
			argName:  "calibrationGuid",
		},
		{
			name:     "update",
			toolName: "UteamupAssetcalibrationUpdate",
			method:   "PUT",
			path:     "by-guid/{calibrationGuid}",
			argName:  "calibrationGuid",
		},
		{
			name:     "delete",
			toolName: "UteamupAssetcalibrationDelete",
			method:   "DELETE",
			path:     "by-guid/{calibrationGuid}",
			argName:  "calibrationGuid",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			domain, action := findAssetCalibrationAction(t, test.name)
			if domain.APIPath != "/api/assetcalibration" {
				t.Fatalf("APIPath = %q, want /api/assetcalibration", domain.APIPath)
			}
			if action.ToolName != test.toolName || action.HTTPMethod != test.method || action.RESTPath != test.path {
				t.Fatalf("unexpected %s action: %+v", test.name, action)
			}
			if len(action.Args) != 1 {
				t.Fatalf("%s args = %+v, want one public GUID", test.name, action.Args)
			}
			arg := action.Args[0]
			if arg.Name != test.argName || arg.Type != "string" || !arg.Required {
				t.Fatalf("%s selector = %+v, want required string %s", test.name, arg, test.argName)
			}
			if arg.Name == "id" || arg.Name == "asset-id" || arg.Type == "int" {
				t.Fatalf("%s leaked an integer selector: %+v", test.name, arg)
			}

			guid := "11111111-2222-3333-4444-555555555555"
			path, consumed := buildRESTPath(domain, *action, map[string]any{test.argName: guid})
			wantPath := strings.ReplaceAll(
				"/api/assetcalibration/"+test.path,
				"{"+test.argName+"}",
				guid,
			)
			if path != wantPath {
				t.Fatalf("%s path = %q, want %q", test.name, path, wantPath)
			}
			if len(consumed) != 1 || consumed[0] != test.argName {
				t.Fatalf("%s consumed args = %v, want [%s]", test.name, consumed, test.argName)
			}
		})
	}
}

func TestAssetCalibrationCreateUsesGuidBodyContractWithoutIntegerSelectors(t *testing.T) {
	domain, action := findAssetCalibrationAction(t, "create")
	if domain.APIPath != "/api/assetcalibration" {
		t.Fatalf("APIPath = %q, want /api/assetcalibration", domain.APIPath)
	}
	if action.ToolName != "UteamupAssetcalibrationCreate" || action.HTTPMethod != "POST" {
		t.Fatalf("unexpected create action: %+v", action)
	}
	if len(action.Args) != 0 {
		t.Fatalf("create unexpectedly accepts a positional identifier: %+v", action.Args)
	}
	if len(action.Flags) != 1 || action.Flags[0].Name != "from-json" {
		t.Fatalf("create must accept only the backend request JSON: %+v", action.Flags)
	}
}
