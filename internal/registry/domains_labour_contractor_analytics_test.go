package registry

import "testing"

func TestLabourContractorAnalyticsDomain(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "labour-contractor-analytics" {
			domain = candidate
			break
		}
	}
	if domain == nil {
		t.Fatal("labour-contractor-analytics domain not registered")
	}
	if domain.APIPath != "/api/labour-marketplace" {
		t.Fatalf("unexpected API path %q", domain.APIPath)
	}
	var action *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "me" {
			action = &domain.Actions[index]
			break
		}
	}
	if action == nil {
		t.Fatal("me action not registered")
	}
	if action.ToolName != "UteamupLabourContractorAnalyticsMe" {
		t.Fatalf("unexpected tool name %q", action.ToolName)
	}
	if action.RESTPath != "analytics/me" || action.HTTPMethod != "GET" {
		t.Fatalf("unexpected REST contract %s %s", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 0 || len(action.Flags) != 0 {
		t.Fatal("current-user analytics must not accept provider selectors")
	}
}
