package registry

import "testing"

func TestLocationAccessScheduleDomainUsesGuidOnlyRoutesAndFlags(t *testing.T) {
	domain := findDomain("location-access-schedule")
	if domain == nil {
		t.Fatal("location-access-schedule domain is not registered")
	}
	if domain.APIPath != "/api/location" {
		t.Fatalf("APIPath = %q, want /api/location", domain.APIPath)
	}

	expected := map[string]struct {
		tool   string
		method string
		path   string
	}{
		"list":   {"UteamupLocationAccessScheduleList", "GET", "{locationGuid}/access-schedules"},
		"get":    {"UteamupLocationAccessScheduleGet", "GET", "access-schedules/{scheduleGuid}"},
		"create": {"UteamupLocationAccessScheduleCreate", "POST", "{locationGuid}/access-schedules"},
		"update": {"UteamupLocationAccessScheduleUpdate", "PUT", "access-schedules/{scheduleGuid}"},
		"delete": {"UteamupLocationAccessScheduleDelete", "DELETE", "access-schedules/{scheduleGuid}"},
	}
	for name, want := range expected {
		action := findAction(domain, name)
		if action == nil {
			t.Fatalf("action %q is not registered", name)
		}
		if action.ToolName != want.tool || action.HTTPMethod != want.method || action.RESTPath != want.path {
			t.Errorf(
				"%s = tool %q method %q path %q, want %q %q %q",
				name,
				action.ToolName,
				action.HTTPMethod,
				action.RESTPath,
				want.tool,
				want.method,
				want.path,
			)
		}
		for _, arg := range action.Args {
			if arg.Name == "locationId" || arg.Name == "scheduleId" || arg.Type == "int" {
				t.Errorf("%s exposes integer identifier argument %+v", name, arg)
			}
		}
	}
}

func TestLocationAccessScheduleMutationsRequireSafetyEvidence(t *testing.T) {
	domain := findDomain("location-access-schedule")
	for _, actionName := range []string{"create", "update", "delete"} {
		action := findAction(domain, actionName)
		assertRequiredFlag(t, action, "idempotency-key", "idempotencyKey")
	}
	for _, actionName := range []string{"update", "delete"} {
		action := findAction(domain, actionName)
		assertRequiredFlag(t, action, "expected-updated-at", "expectedUpdatedAtUtc")
	}
}

func assertRequiredFlag(
	t *testing.T,
	action *Action,
	name string,
	bodyName string,
) {
	t.Helper()
	if action == nil {
		t.Fatalf("action is nil while checking flag %q", name)
	}
	for _, flag := range action.Flags {
		if flag.Name == name {
			if !flag.Required || flag.BodyName != bodyName {
				t.Errorf(
					"%s flag %q = required %v body %q",
					action.Name,
					name,
					flag.Required,
					flag.BodyName,
				)
			}
			return
		}
	}
	t.Errorf("%s action is missing flag %q", action.Name, name)
}
