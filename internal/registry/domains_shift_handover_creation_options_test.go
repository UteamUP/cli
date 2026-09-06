package registry

import "testing"

func TestHandoverCreationOptionsExposeBoundedGuidLookup(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	action := findShiftHandoverAction(t, domain, "creation-options")
	if action.ToolName != "UteamupShiftHandoverGetCreateOptions" || action.HTTPMethod != "GET" || action.RESTPath != "creation-options" {
		t.Fatalf("unexpected creation options route: %+v", action)
	}
	if len(action.Args) != 0 || len(action.Flags) != 5 {
		t.Fatalf("expected bounded filter flags: %+v", action)
	}
	for _, flag := range action.Flags {
		if flag.Name == "selected-guid" && (flag.Type != "uuid" || flag.BodyName != "selectedGuid") {
			t.Fatalf("selection must remain a public GUID: %+v", flag)
		}
	}
	path, consumed := buildRESTPath(domain, *action, map[string]any{})
	if path != "/api/shifthandover/creation-options" || len(consumed) != 0 {
		t.Fatalf("unexpected lookup path: %s, consumed=%v", path, consumed)
	}
}
