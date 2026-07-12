package registry

import "testing"

func TestMarketplaceConversationDomainIsParticipantScopedAndGuidFirst(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "marketplace-conversation" {
			domain = candidate
			break
		}
	}
	if domain == nil {
		t.Fatal("expected marketplace-conversation domain")
	}
	if domain.APIPath != "/api/marketplace/conversations" {
		t.Fatalf("APIPath = %q", domain.APIPath)
	}

	want := map[string]struct {
		tool string
		path string
		args int
	}{
		"search": {"UteamupMarketplaceConversationMessagesSearch", "{conversationGuid}/messages/search", 1},
		"mute":   {"UteamupMarketplaceConversationPreferencesUpdate", "{conversationGuid}/preferences", 1},
		"pin":    {"UteamupMarketplaceConversationMessagePinUpdate", "{conversationGuid}/messages/{messageGuid}/pin", 2},
	}
	for _, action := range domain.Actions {
		expected, ok := want[action.Name]
		if !ok {
			continue
		}
		if action.ToolName != expected.tool || action.RESTPath != expected.path {
			t.Errorf("%s contract = %q %q", action.Name, action.ToolName, action.RESTPath)
		}
		if len(action.Args) != expected.args || action.Args[0].Name != "conversationGuid" {
			t.Errorf("%s must use public GUID arguments, got %+v", action.Name, action.Args)
		}
		for _, arg := range action.Args {
			if arg.Type != "string" {
				t.Errorf("%s argument %s must remain a GUID string", action.Name, arg.Name)
			}
			if arg.Name == "userId" || arg.Name == "tenantId" {
				t.Errorf("%s must not accept spoofable identity arguments", action.Name)
			}
		}
		delete(want, action.Name)
	}
	if len(want) != 0 {
		t.Fatalf("missing actions: %v", want)
	}
}
