package registry

import "testing"

func TestAIUsageDataReadinessMirrorsGuidFreeReadTool(t *testing.T) {
	domain := findRegisteredDomain(t, "ai-usage")
	var readiness *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "data-readiness" {
			readiness = &domain.Actions[index]
			break
		}
	}
	if readiness == nil {
		t.Fatal("data-readiness action is not registered")
	}
	if readiness.ToolName != "UteamupAIDataReadiness" ||
		readiness.HTTPMethod != "GET" ||
		readiness.RESTBasePath != "/api/aianalytics" ||
		readiness.RESTPath != "data-readiness" {
		t.Fatalf("unexpected readiness contract: %#v", readiness)
	}
	if len(readiness.Args) != 0 || len(readiness.Flags) != 0 {
		t.Fatalf("readiness must derive tenant identity from authentication: %#v", readiness)
	}

	path, consumed := buildRESTPath(domain, *readiness, map[string]any{})
	if path != "/api/aianalytics/data-readiness" || len(consumed) != 0 {
		t.Fatalf("unexpected resolved route %q with consumed args %v", path, consumed)
	}
}
