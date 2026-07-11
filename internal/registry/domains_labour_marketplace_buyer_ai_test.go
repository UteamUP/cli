package registry

import "testing"

func TestLabourBuyerAiDomainMirrorsEditableGuidFirstTools(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "labour-ai-buyer" {
			domain = candidate
			break
		}
	}
	if domain == nil {
		t.Fatal("expected labour-ai-buyer domain")
	}
	if domain.APIPath != "/api/labour-marketplace/ai/buyer-job-draft" {
		t.Fatalf("APIPath = %q", domain.APIPath)
	}

	want := map[string]string{
		"cost":   "UteamupLabourMarketplaceBuyerJobDraftCost",
		"create": "UteamupLabourMarketplaceBuyerJobDraftCreate",
	}
	for _, action := range domain.Actions {
		if tool, ok := want[action.Name]; ok {
			if action.ToolName != tool {
				t.Errorf("%s ToolName = %q, want %q", action.Name, action.ToolName, tool)
			}
			delete(want, action.Name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing actions: %v", want)
	}

	create := domain.Actions[1]
	requiredDescription := false
	guidFlags := map[string]bool{"project-guid": false, "workorder-guid": false}
	for _, flag := range create.Flags {
		if flag.Name == "description" {
			requiredDescription = flag.Required
		}
		if _, ok := guidFlags[flag.Name]; ok {
			guidFlags[flag.Name] = flag.Type == "string"
		}
	}
	if !requiredDescription {
		t.Error("--description must be required")
	}
	for name, valid := range guidFlags {
		if !valid {
			t.Errorf("--%s must remain a GUID string flag", name)
		}
	}
}
