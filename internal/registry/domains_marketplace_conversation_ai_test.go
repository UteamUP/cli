package registry

import "testing"

func TestMarketplaceConversationAiDomainIsParticipantScopedAndGuidFirst(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "marketplace-conversation-ai" {
			domain = candidate
			break
		}
	}
	if domain == nil {
		t.Fatal("expected marketplace-conversation-ai domain")
	}
	if domain.APIPath != "/api/marketplace/conversations" {
		t.Fatalf("APIPath = %q", domain.APIPath)
	}

	want := map[string]string{
		"cost":      "UteamupMarketplaceConversationAiSummaryCost",
		"summarize": "UteamupMarketplaceConversationAiSummary",
	}
	for _, action := range domain.Actions {
		if tool, ok := want[action.Name]; ok {
			if action.ToolName != tool {
				t.Errorf("%s ToolName = %q, want %q", action.Name, action.ToolName, tool)
			}
			if len(action.Args) != 1 || action.Args[0].Name != "conversationGuid" || action.Args[0].Type != "string" {
				t.Errorf("%s must use one public conversation GUID argument, got %+v", action.Name, action.Args)
			}
			delete(want, action.Name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing actions: %v", want)
	}
}
