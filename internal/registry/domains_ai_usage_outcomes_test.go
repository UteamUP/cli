package registry

import "testing"

func TestAIUsageOutcomesMirrorsGuidFirstReadTool(t *testing.T) {
	domain := findRegisteredDomain(t, "ai-usage")
	var outcomes *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "outcomes" {
			outcomes = &domain.Actions[index]
			break
		}
	}
	if outcomes == nil {
		t.Fatal("outcomes action is not registered")
	}
	if outcomes.ToolName != "UteamupAIOutcomeMeasurements" ||
		outcomes.HTTPMethod != "GET" || outcomes.RESTBasePath != "/api/aianalytics" ||
		outcomes.RESTPath != "outcomes" {
		t.Fatalf("unexpected outcome contract: %#v", outcomes)
	}
	for _, flag := range outcomes.Flags {
		if flag.Name == "tenant-id" || flag.Name == "tenant-guid" || flag.Name == "id" {
			t.Fatalf("outcome list must derive tenant identity from authentication: %#v", flag)
		}
	}
}
