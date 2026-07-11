package registry

import "testing"

func TestLabourOfferAiDomainMirrorsAdvisoryGuidFirstTools(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "labour-ai-offers" {
			domain = candidate
			break
		}
	}
	if domain == nil {
		t.Fatal("expected labour-ai-offers domain")
	}
	if domain.APIPath != "/api/labour-marketplace/ai/offer-comparison" {
		t.Fatalf("APIPath = %q", domain.APIPath)
	}

	want := map[string]string{
		"cost":    "UteamupLabourMarketplaceOfferComparisonCost",
		"compare": "UteamupLabourMarketplaceOfferComparison",
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

	compare := domain.Actions[1]
	flags := make(map[string]FlagDef, len(compare.Flags))
	for _, flag := range compare.Flags {
		flags[flag.Name] = flag
	}
	if job := flags["job-guid"]; !job.Required || job.Type != "string" {
		t.Errorf("--job-guid must be a required GUID string flag, got %+v", job)
	}
	if revisions := flags["revision-guids"]; revisions.Required || revisions.Type != "stringSlice" {
		t.Errorf("--revision-guids must be an optional GUID string slice, got %+v", revisions)
	}
	if description := compare.Description; description == "" {
		t.Error("compare action must explain its advisory-only behavior")
	}
}
