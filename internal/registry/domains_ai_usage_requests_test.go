package registry

import "testing"

func TestAIUsageRequestsMirrorsTenantScopedHistory(t *testing.T) {
	domain := findRegisteredDomain(t, "ai-usage")
	var requests *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "requests" {
			requests = &domain.Actions[index]
			break
		}
	}
	if requests == nil {
		t.Fatal("requests action is not registered")
	}
	if requests.ToolName != "UteamupAIUsageRequests" || requests.HTTPMethod != "GET" ||
		requests.RESTBasePath != "/api/aiusage" || requests.RESTPath != "requests" {
		t.Fatalf("unexpected request-history contract: %#v", requests)
	}
	flags := map[string]FlagDef{}
	for _, flag := range requests.Flags {
		flags[flag.Name] = flag
		if flag.Name == "tenant-id" || flag.Name == "tenant-guid" || flag.Name == "user-id" {
			t.Fatalf("history must use server-owned tenant and public actor GUIDs: %#v", flag)
		}
	}
	for _, name := range []string{"from", "to", "page", "page-size", "task-key", "operation-key", "credential-source", "user-guid"} {
		if _, found := flags[name]; !found {
			t.Errorf("missing request-history filter %s", name)
		}
	}
	if flags["page"].Default != 1 || flags["page-size"].Default != 25 {
		t.Fatal("request-history pagination defaults drifted")
	}
}
