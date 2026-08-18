package registry

import "testing"

func TestTeamMoodDomainIsAggregatesOnly(t *testing.T) {
	domain := findDomain("teammood")
	if domain == nil {
		t.Fatal("expected teammood domain")
	}
	if domain.APIPath != "/api/teammood" {
		t.Fatalf("API path = %q", domain.APIPath)
	}
	if len(domain.Actions) != 1 {
		t.Fatalf("actions = %d, want 1 (aggregates only)", len(domain.Actions))
	}

	action := domain.Actions[0]
	if action.Name != "aggregates" ||
		action.ToolName != "UteamupTeamMoodRead" ||
		action.HTTPMethod != "GET" ||
		action.RESTPath != "aggregates" {
		t.Fatalf("aggregates action = %+v", action)
	}

	path, _ := buildRESTPath(domain, action, map[string]any{})
	if path != "/api/teammood/aggregates" {
		t.Fatalf("path = %q", path)
	}

	for _, item := range domain.Actions {
		if item.Name == "checkin" || item.Name == "checkins" || item.RESTPath == "checkins" {
			t.Fatalf("check-in dump action is forbidden: %+v", item)
		}
		for _, flag := range item.Flags {
			if flag.Name == "id" || flag.Name == "user-id" || flag.Name == "email" {
				t.Fatalf("forbidden flag %q on %s", flag.Name, item.Name)
			}
		}
	}
}
