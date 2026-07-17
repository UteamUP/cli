package registry

import "testing"

func TestScheduleMyActionMirrorsGuidSafeMCPRead(t *testing.T) {
	domain := findDomain("schedule")
	if domain == nil {
		t.Fatal("schedule domain is not registered")
	}
	if domain.APIPath != "/api/schedule" {
		t.Fatalf("schedule APIPath = %q, want /api/schedule", domain.APIPath)
	}

	var action *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "my" {
			action = &domain.Actions[index]
			break
		}
	}
	if action == nil {
		t.Fatal("schedule my action is not registered")
	}
	if action.ToolName != "UteamupScheduleGetMySchedule" ||
		action.HTTPMethod != "GET" || action.RESTPath != "me" {
		t.Fatalf(
			"schedule my action = tool %q, method %q, path %q",
			action.ToolName,
			action.HTTPMethod,
			action.RESTPath,
		)
	}

	flags := make(map[string]FlagDef, len(action.Flags))
	for _, flag := range action.Flags {
		flags[flag.Name] = flag
	}
	for name, bodyName := range map[string]string{
		"start-date": "startDate",
		"end-date":   "endDate",
	} {
		flag, ok := flags[name]
		if !ok || !flag.Required || flag.Type != "string" || flag.BodyName != bodyName {
			t.Fatalf("schedule my flag %q is not a required %q string: %+v", name, bodyName, flag)
		}
	}
}
