package registry

import "testing"

func TestAIUsageKnowledgeTutorialTrustMirrorsGuidFreeReadTool(t *testing.T) {
	domain := findRegisteredDomain(t, "ai-usage")
	var trust *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "knowledge-tutorial-trust" {
			trust = &domain.Actions[index]
			break
		}
	}
	if trust == nil {
		t.Fatal("knowledge-tutorial-trust action is not registered")
	}
	if trust.ToolName != "UteamupAIKnowledgeTutorialTrust" ||
		trust.HTTPMethod != "GET" ||
		trust.RESTBasePath != "/api/aianalytics" ||
		trust.RESTPath != "knowledge-tutorial-trust" {
		t.Fatalf("unexpected Knowledge/tutorial trust contract: %#v", trust)
	}
	if len(trust.Args) != 0 || len(trust.Flags) != 0 {
		t.Fatalf("trust read must derive tenant identity from authentication: %#v", trust)
	}

	path, consumed := buildRESTPath(domain, *trust, map[string]any{})
	if path != "/api/aianalytics/knowledge-tutorial-trust" || len(consumed) != 0 {
		t.Fatalf("unexpected resolved route %q with consumed args %v", path, consumed)
	}
}
