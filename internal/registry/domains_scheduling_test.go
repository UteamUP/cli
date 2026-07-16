package registry

import (
	"strings"
	"testing"
)

func scheduleAssignmentDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "schedule-assignment" {
			return domain
		}
	}
	t.Fatal("expected schedule-assignment domain")
	return nil
}

func scheduleAssignmentAction(t *testing.T, name string) *Action {
	t.Helper()
	for index := range scheduleAssignmentDomain(t).Actions {
		action := &scheduleAssignmentDomain(t).Actions[index]
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("expected schedule-assignment %q action", name)
	return nil
}

func TestScheduleAssignmentBookingActionsMirrorGuidMcpTools(t *testing.T) {
	expected := map[string]string{
		"week":   "UteamupScheduleAssignmentGetWeekByGuid",
		"move":   "UteamupScheduleAssignmentMoveByGuid",
		"status": "UteamupScheduleAssignmentUpdateStatusByGuid",
		"cancel": "UteamupScheduleAssignmentCancelByGuid",
	}

	for name, toolName := range expected {
		action := scheduleAssignmentAction(t, name)
		if action.ToolName != toolName {
			t.Errorf("%s ToolName = %q, want %q", name, action.ToolName, toolName)
		}
	}
}

func TestScheduleAssignmentBookingRoutesAreGuidFirst(t *testing.T) {
	domain := scheduleAssignmentDomain(t)
	if domain.APIPath != "/api/scheduleassignment" {
		t.Fatalf("APIPath = %q, want /api/scheduleassignment", domain.APIPath)
	}

	cases := []struct {
		name       string
		method     string
		path       string
		arguments  map[string]any
		consumedBy string
	}{
		{"week", "GET", "/api/scheduleassignment/week", map[string]any{}, ""},
		{"move", "PUT", "/api/scheduleassignment/assignment-guid/move", map[string]any{"assignmentGuid": "assignment-guid"}, "assignmentGuid"},
		{"status", "PUT", "/api/scheduleassignment/assignment-guid/status", map[string]any{"assignmentGuid": "assignment-guid"}, "assignmentGuid"},
		{"cancel", "DELETE", "/api/scheduleassignment/assignment-guid", map[string]any{"assignmentGuid": "assignment-guid"}, "assignmentGuid"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			action := scheduleAssignmentAction(t, test.name)
			path, consumed := buildRESTPath(domain, *action, test.arguments)
			if action.HTTPMethod != test.method || path != test.path {
				t.Errorf("route = %s %s, want %s %s", action.HTTPMethod, path, test.method, test.path)
			}
			if test.consumedBy != "" && (len(consumed) != 1 || consumed[0] != test.consumedBy) {
				t.Errorf("consumed args = %v, want [%s]", consumed, test.consumedBy)
			}
		})
	}
}

func TestScheduleAssignmentDomainPublishesNoIntegerIdentifierArguments(t *testing.T) {
	legacyActions := map[string]bool{"list": true, "get": true, "create": true, "update": true, "delete": true}
	for _, action := range scheduleAssignmentDomain(t).Actions {
		if legacyActions[action.Name] {
			t.Errorf("legacy integer CRUD action %q must not remain public", action.Name)
		}
		for _, argument := range action.Args {
			if strings.Contains(strings.ToLower(argument.Name), "id") && argument.Type == "int" {
				t.Errorf("%s exposes integer identifier argument %+v", action.Name, argument)
			}
		}
	}
}

func TestScheduleAssignmentCreateByGuidPostsBackendModelFieldNames(t *testing.T) {
	action := scheduleAssignmentAction(t, "create-by-guid")
	if action.HTTPMethod != "POST" || action.RESTPath != "" {
		t.Fatalf("create-by-guid route = %s %q, want POST base path", action.HTTPMethod, action.RESTPath)
	}

	wantBodyNames := map[string]string{
		"planned-start-utc": "scheduledStart",
		"planned-end-utc":   "scheduledEnd",
	}
	for _, flag := range action.Flags {
		if expected, ok := wantBodyNames[flag.Name]; ok {
			if flag.BodyName != expected {
				t.Errorf("--%s BodyName = %q, want %q", flag.Name, flag.BodyName, expected)
			}
			delete(wantBodyNames, flag.Name)
		}
	}
	if len(wantBodyNames) != 0 {
		t.Errorf("missing create-by-guid flags: %v", wantBodyNames)
	}
}
