package registry

import "testing"

func TestProjectReservationReadUsesExactGUIDAndExplicitWindow(t *testing.T) {
	domain := findDomain("workforce-capacity")
	action := findAction(domain, "project-reservations")
	if action == nil || action.HTTPMethod != "GET" || action.ToolName != "UteamupWorkforceProjectReservations" {
		t.Fatalf("missing current reservation read: %+v", action)
	}
	assertRequiredUUIDArg(t, action, "projectGuid")
	flags := flagsToMap(action.Flags)
	if flags["from-utc"].QueryName != "from" || flags["from-utc"].BodyName != "fromUtc" || !flags["from-utc"].Required ||
		flags["to-utc"].QueryName != "to" || flags["to-utc"].BodyName != "toUtc" || !flags["to-utc"].Required {
		t.Fatalf("REST window and MCP arguments must stay explicit: %+v", flags)
	}
	path, consumed := buildRESTPath(domain, *action, map[string]any{"projectGuid": documentGUID})
	if path != "/api/workforcecapacity/projects/"+documentGUID+"/reservations" || len(consumed) != 1 {
		t.Fatalf("unexpected project reservation path %s", path)
	}
}
