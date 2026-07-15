package registry

import (
	"reflect"
	"testing"
)

func scheduleOptimizationAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("schedule-optimization-run")
	if domain == nil {
		t.Fatal("schedule-optimization-run domain is not registered")
	}

	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}

	t.Fatalf("schedule-optimization-run action %q is not registered", name)
	return nil, Action{}
}

func TestScheduleOptimizationRunRoutesUseGuidIdentity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		args     map[string]any
		path     string
		consumed []string
	}{
		{name: "create", args: map[string]any{}, path: "/api/schedule/optimization-runs"},
		{name: "get", args: map[string]any{"runGuid": "run-guid"}, path: "/api/schedule/optimization-runs/run-guid", consumed: []string{"runGuid"}},
		{name: "cancel", args: map[string]any{"runGuid": "run-guid"}, path: "/api/schedule/optimization-runs/run-guid/cancel", consumed: []string{"runGuid"}},
		{name: "apply", args: map[string]any{"runGuid": "run-guid"}, path: "/api/schedule/optimization-runs/run-guid/apply", consumed: []string{"runGuid"}},
		{name: "revert", args: map[string]any{"runGuid": "run-guid"}, path: "/api/schedule/optimization-runs/run-guid/revert", consumed: []string{"runGuid"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			domain, action := scheduleOptimizationAction(t, testCase.name)
			path, consumed := buildRESTPath(domain, action, testCase.args)
			if path != testCase.path {
				t.Fatalf("path = %q, want %q", path, testCase.path)
			}
			if !reflect.DeepEqual(consumed, testCase.consumed) {
				t.Fatalf("consumed = %v, want %v", consumed, testCase.consumed)
			}
			for _, arg := range action.Args {
				if arg.Type != "uuid" {
					t.Fatalf("public identity arg is not UUID typed: %+v", arg)
				}
			}
		})
	}
}

func TestScheduleOptimizationCreateFlagsMirrorBackendModel(t *testing.T) {
	t.Parallel()

	_, create := scheduleOptimizationAction(t, "create")
	want := map[string]struct {
		bodyName string
		flagType string
		required bool
	}{
		"idempotency-key":  {bodyName: "idempotencyKey", flagType: "string", required: true},
		"week-start":       {bodyName: "weekStart", flagType: "string", required: true},
		"workorder-guids":  {bodyName: "workorderGuids", flagType: "stringSlice", required: true},
		"technician-guids": {bodyName: "technicianGuids", flagType: "stringSlice", required: true},
		"team-guid":        {bodyName: "teamGuid", flagType: "string", required: false},
	}

	if len(create.Flags) != len(want) {
		t.Fatalf("create flags = %d, want %d", len(create.Flags), len(want))
	}
	for _, flag := range create.Flags {
		expected, ok := want[flag.Name]
		if !ok {
			t.Fatalf("unexpected create flag: %+v", flag)
		}
		if flag.BodyName != expected.bodyName ||
			flag.Type != expected.flagType ||
			flag.Required != expected.required {
			t.Fatalf("flag %q = %+v, want %+v", flag.Name, flag, expected)
		}
	}

	_, get := scheduleOptimizationAction(t, "get")
	_, cancel := scheduleOptimizationAction(t, "cancel")
	_, apply := scheduleOptimizationAction(t, "apply")
	_, revert := scheduleOptimizationAction(t, "revert")
	if get.ToolName != "UteamupScheduleOptimizationRunGet" ||
		cancel.ToolName != "UteamupScheduleOptimizationRunCancel" ||
		apply.ToolName != "UteamupScheduleOptimizationRunApply" ||
		revert.ToolName != "UteamupScheduleOptimizationRunRevert" {
		t.Fatal("CLI tool names must mirror backend MCP methods exactly")
	}
}

func TestScheduleOptimizationApplyAndRevertFlagsMirrorBackendModels(t *testing.T) {
	t.Parallel()

	_, apply := scheduleOptimizationAction(t, "apply")
	wantApply := map[string]struct {
		bodyName string
		flagType string
		required bool
	}{
		"idempotency-key":          {bodyName: "idempotencyKey", flagType: "string", required: true},
		"selected-workorder-guids": {bodyName: "selectedWorkorderGuids", flagType: "stringSlice", required: false},
	}

	if len(apply.Flags) != len(wantApply) {
		t.Fatalf("apply flags = %d, want %d", len(apply.Flags), len(wantApply))
	}
	for _, flag := range apply.Flags {
		expected, ok := wantApply[flag.Name]
		if !ok {
			t.Fatalf("unexpected apply flag: %+v", flag)
		}
		if flag.BodyName != expected.bodyName ||
			flag.Type != expected.flagType ||
			flag.Required != expected.required {
			t.Fatalf("apply flag %q = %+v, want %+v", flag.Name, flag, expected)
		}
	}

	_, revert := scheduleOptimizationAction(t, "revert")
	wantRevert := map[string]struct {
		bodyName string
		flagType string
		required bool
	}{
		"idempotency-key": {bodyName: "idempotencyKey", flagType: "string", required: true},
		"reason":          {bodyName: "reason", flagType: "string", required: false},
	}

	if len(revert.Flags) != len(wantRevert) {
		t.Fatalf("revert flags = %d, want %d", len(revert.Flags), len(wantRevert))
	}
	for _, flag := range revert.Flags {
		expected, ok := wantRevert[flag.Name]
		if !ok {
			t.Fatalf("unexpected revert flag: %+v", flag)
		}
		if flag.BodyName != expected.bodyName ||
			flag.Type != expected.flagType ||
			flag.Required != expected.required {
			t.Fatalf("revert flag %q = %+v, want %+v", flag.Name, flag, expected)
		}
	}
}
