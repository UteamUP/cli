package registry

import (
	"strings"
	"testing"
)

func operationsBrainAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("operations-brain")
	if domain == nil {
		t.Fatal("operations-brain domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("operations-brain action %q is not registered", name)
	return nil, Action{}
}

func TestOperationsBrainRoutesAndArgumentsAreGuidOnly(t *testing.T) {
	t.Parallel()
	tests := []struct {
		actionName string
		arguments  map[string]any
		path       string
	}{
		{
			"get-plan",
			map[string]any{"planGuid": "plan-guid"},
			"/api/upmateassistant/operations/plans/plan-guid",
		},
		{
			"select-scenario",
			map[string]any{
				"planGuid":     "plan-guid",
				"scenarioGuid": "scenario-guid",
			},
			"/api/upmateassistant/operations/plans/plan-guid/scenarios/scenario-guid/select",
		},
		{
			"prepare-run",
			map[string]any{"planGuid": "plan-guid"},
			"/api/upmateassistant/operations/plans/plan-guid/runs",
		},
	}

	for _, test := range tests {
		t.Run(test.actionName, func(t *testing.T) {
			domain, action := operationsBrainAction(t, test.actionName)
			path, consumed := buildRESTPath(domain, action, test.arguments)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			if len(consumed) != len(action.Args) {
				t.Fatalf("route did not consume every GUID arg: args=%+v consumed=%v", action.Args, consumed)
			}
			for _, argument := range action.Args {
				lowerName := strings.ToLower(argument.Name)
				hasIntegerIDName := strings.HasSuffix(lowerName, "id") &&
					!strings.HasSuffix(lowerName, "guid")
				if argument.Type != "uuid" || hasIntegerIDName {
					t.Fatalf("public plan identity is not GUID-only: %+v", argument)
				}
			}
		})
	}
}

func TestOperationsBrainActionsMirrorBackendToolsAndGovernance(t *testing.T) {
	t.Parallel()
	expected := map[string]string{
		"risks":           "UteamupOperationsGetRisks",
		"create-plan":     "UteamupOperationsCreatePlan",
		"get-plan":        "UteamupOperationsGetPlan",
		"select-scenario": "UteamupOperationsSelectScenario",
		"prepare-run":     "UteamupOperationsPrepareRun",
	}
	for actionName, toolName := range expected {
		_, action := operationsBrainAction(t, actionName)
		if action.ToolName != toolName {
			t.Fatalf("%s tool = %q, want %q", actionName, action.ToolName, toolName)
		}
		for _, flag := range action.Flags {
			lower := strings.ToLower(flag.Name)
			if strings.Contains(lower, "tenant") || strings.Contains(lower, "user") {
				t.Fatalf("%s exposes caller-controlled scope: %+v", actionName, flag)
			}
		}
	}

	_, create := operationsBrainAction(t, "create-plan")
	assertOperationsBrainFlag(t, create, "request-guid", "requestGuid", true, "uuid")
	assertOperationsBrainFlag(t, create, "domains", "domains", true, "stringSlice")

	for _, actionName := range []string{"select-scenario", "prepare-run"} {
		_, action := operationsBrainAction(t, actionName)
		if action.HTTPMethod != "POST" {
			t.Fatalf("%s must be an explicit governed POST: %+v", actionName, action)
		}
		assertOperationsBrainFlag(
			t,
			action,
			"request-guid",
			"requestGuid",
			true,
			"uuid",
		)
		assertOperationsBrainFlag(
			t,
			action,
			"expected-version",
			"expectedVersion",
			true,
			"int",
		)
	}
}

func assertOperationsBrainFlag(
	t *testing.T,
	action Action,
	name string,
	bodyName string,
	required bool,
	flagType string,
) {
	t.Helper()
	for _, flag := range action.Flags {
		if flag.Name == name {
			if flag.BodyName != bodyName ||
				flag.Required != required ||
				flag.Type != flagType {
				t.Fatalf("%s flag = %+v", name, flag)
			}
			return
		}
	}
	t.Fatalf("%s flag is missing", name)
}
