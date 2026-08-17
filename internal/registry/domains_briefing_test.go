package registry

import "testing"

func TestBriefingDomainUsesGuidFirstRoutesAndNoSms(t *testing.T) {
	domain := findDomain("briefing")
	if domain == nil {
		t.Fatal("expected briefing domain")
	}
	if domain.APIPath != "/api/dailybriefing" {
		t.Fatalf("API path = %q", domain.APIPath)
	}
	if len(domain.Actions) != 3 {
		t.Fatalf("actions = %d, want 3", len(domain.Actions))
	}

	today := domain.Actions[0]
	if today.Name != "today" ||
		today.ToolName != "UteamupBriefingDailyRead" ||
		today.HTTPMethod != "GET" ||
		today.RESTPath != "today" {
		t.Fatalf("today action = %+v", today)
	}

	confirm := domain.Actions[2]
	if confirm.Name != "confirm" ||
		confirm.ToolName != "UteamupBriefingDecisionPropose" ||
		confirm.HTTPMethod != "POST" ||
		confirm.RESTPath != "{briefingGuid}/confirm" {
		t.Fatalf("confirm action = %+v", confirm)
	}

	path, consumed := buildRESTPath(domain, confirm, map[string]any{
		"briefingGuid": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	})
	if path != "/api/dailybriefing/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/confirm" {
		t.Fatalf("path = %q", path)
	}
	if len(consumed) != 1 {
		t.Fatalf("consumed = %v", consumed)
	}

	for _, action := range domain.Actions {
		for _, flag := range action.Flags {
			if flag.Name == "id" || flag.Name == "briefing-id" || flag.Name == "sms" {
				t.Fatalf("forbidden flag %q on %s", flag.Name, action.Name)
			}
		}
	}
}
