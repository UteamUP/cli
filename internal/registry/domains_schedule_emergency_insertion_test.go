package registry

import "testing"

func TestScheduleEmergencyInsertionPreviewMirrorsGuidOnlyBackendContract(t *testing.T) {
	t.Parallel()

	domain := findDomain("schedule-emergency-insertion")
	if domain == nil {
		t.Fatal("schedule-emergency-insertion domain is not registered")
	}
	if domain.APIPath != "/api/schedule/emergency-insertions" {
		t.Fatalf("API path = %q", domain.APIPath)
	}
	if len(domain.Actions) != 1 {
		t.Fatalf("actions = %d, want 1", len(domain.Actions))
	}
	action := domain.Actions[0]
	if action.Name != "preview" || action.HTTPMethod != "POST" || action.RESTPath != "preview" {
		t.Fatalf("unexpected preview action: %+v", action)
	}
	if action.ToolName != "UteamupScheduleEmergencyInsertionPreview" {
		t.Fatalf("tool name = %q", action.ToolName)
	}
	path, consumed := buildRESTPath(domain, action, map[string]any{})
	if path != "/api/schedule/emergency-insertions/preview" {
		t.Fatalf("path = %q", path)
	}
	if len(consumed) != 0 {
		t.Fatalf("consumed = %v", consumed)
	}

	want := map[string]struct {
		bodyName string
		flagType string
		required bool
	}{
		"workorder-guid":            {bodyName: "workorderGuid", flagType: "string", required: true},
		"desired-start-utc":         {bodyName: "desiredStartUtc", flagType: "string", required: true},
		"planning-window-end-utc":   {bodyName: "planningWindowEndUtc", flagType: "string", required: false},
		"team-guid":                 {bodyName: "teamGuid", flagType: "string", required: false},
		"technician-guids":          {bodyName: "technicianGuids", flagType: "stringSlice", required: false},
		"max-displaced-assignments": {bodyName: "maxDisplacedAssignments", flagType: "int", required: false},
	}
	if len(action.Flags) != len(want) {
		t.Fatalf("flags = %d, want %d", len(action.Flags), len(want))
	}
	for _, flag := range action.Flags {
		expected, ok := want[flag.Name]
		if !ok {
			t.Fatalf("unexpected flag: %+v", flag)
		}
		if flag.BodyName != expected.bodyName ||
			flag.Type != expected.flagType ||
			flag.Required != expected.required {
			t.Fatalf("flag %q = %+v, want %+v", flag.Name, flag, expected)
		}
		if flag.BodyName == "workorderId" || flag.BodyName == "teamId" || flag.BodyName == "technicianIds" {
			t.Fatalf("integer/internal identity leaked: %+v", flag)
		}
	}
}
