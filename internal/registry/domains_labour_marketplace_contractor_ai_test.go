package registry

import "testing"

func TestLabourApplicationAiMirrorsCostAndEditableDraftTools(t *testing.T) {
	domain := findDomain("labour-application-ai")
	if domain == nil {
		t.Fatal("expected labour-application-ai domain")
	}
	if domain.APIPath != "/api/labour-marketplace/ai/contractor-application-draft" {
		t.Fatalf("APIPath = %q", domain.APIPath)
	}
	cost := findAction(domain, "cost")
	draft := findAction(domain, "draft")
	if cost == nil || cost.ToolName != "UteamupLabourMarketplaceContractorApplicationDraftCost" {
		t.Fatalf("unexpected cost action: %+v", cost)
	}
	if draft == nil || draft.ToolName != "UteamupLabourMarketplaceContractorApplicationDraftCreate" {
		t.Fatalf("unexpected draft action: %+v", draft)
	}
	required := map[string]bool{"job-guid": false, "provider-party-guid": false}
	for _, flag := range draft.Flags {
		if _, ok := required[flag.Name]; ok {
			required[flag.Name] = flag.Required && flag.Type == "string"
		}
	}
	for name, valid := range required {
		if !valid {
			t.Errorf("--%s must be a required GUID string", name)
		}
	}
}
