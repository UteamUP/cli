package registry

import "testing"

func TestMeterReadingMyViewLogReadsOnlyTheCallersOwnLog(t *testing.T) {
	action := findDomainAction(t, "meter-reading", "my-view-log")
	if action.ToolName != "UteamupMeterreadingMyViewLog" || action.HTTPMethod != "GET" {
		t.Errorf("my-view-log is miswired: tool=%s method=%s", action.ToolName, action.HTTPMethod)
	}
	if len(action.Args) != 0 {
		t.Errorf("my-view-log must take no positional args (it is always the caller's own log), got %d", len(action.Args))
	}
	for _, name := range []string{"user-guid", "subject-user-guid", "user-id", "tenant-guid"} {
		if actionFlagExists(action, name) {
			t.Errorf("my-view-log must not let the caller name another person: found --%s", name)
		}
	}

	take := actionFlagByName(t, action, "take")
	if take.Type != "int" || take.Default != 100 || take.Required {
		t.Errorf("take must be an optional int defaulting to 100: %+v", take)
	}

	domain := findDomain("meter-reading")
	path, _ := buildRESTPath(domain, *action, map[string]any{"take": 25})
	if path != "/api/meter-readings/my/view-log" {
		t.Errorf("my-view-log path = %q, want /api/meter-readings/my/view-log", path)
	}
}
