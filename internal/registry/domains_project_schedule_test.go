package registry

import "testing"

func TestProjectScheduleUsesReviewedOwningRoutes(t *testing.T) {
	for _, expected := range []struct{ action, method, suffix, tool string }{
		{"get", "GET", "", "UteamupProjectScheduleGet"},
		{"preview", "POST", "/preview", "UteamupProjectSchedulePreview"},
		{"apply", "POST", "/apply", "UteamupProjectScheduleApply"},
		{"calendar", "PUT", "/calendar", "UteamupProjectScheduleCalendarSet"},
		{"history", "GET", "/history", "UteamupProjectScheduleHistory"},
		{"dependency-create", "POST", "/dependencies", "UteamupProjectScheduleDependencySave"},
		{"dependency-update", "PUT", "/dependencies/{dependencyGuid}", "UteamupProjectScheduleDependencySave"},
		{"dependency-remove", "POST", "/dependencies/{dependencyGuid}/remove", "UteamupProjectScheduleDependencyRemove"},
	} {
		action := findDomainAction(t, "project-schedule", expected.action)
		if action.HTTPMethod != expected.method || action.RESTPath != "{projectGuid}/timeline"+expected.suffix || action.ToolName != expected.tool {
			t.Fatalf("%s must preserve its owning schedule contract: %+v", expected.action, action)
		}
		for _, arg := range action.Args {
			if arg.Type != "non-empty-uuid" || !arg.Required {
				t.Fatalf("%s must use non-empty public GUIDs: %+v", expected.action, arg)
			}
		}
		if expected.method != "GET" {
			payload := findFlag(action, "from-json")
			if payload == nil || !payload.Required || len(action.Flags) != 1 {
				t.Fatalf("%s must preserve the full reviewed request, including explicit null dates and tokens", expected.action)
			}
		}
	}
	if findDomain("project-schedule").APIPath != "/api/projects" {
		t.Fatal("project schedule must use existing project routes")
	}
	history := findDomainAction(t, "project-schedule", "history")
	if findFlag(history, "page") == nil || findFlag(history, "page-size") == nil {
		t.Fatal("retained history must remain reachable beyond its first page")
	}
}
