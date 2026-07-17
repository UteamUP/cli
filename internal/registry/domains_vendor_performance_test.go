package registry

import "testing"

func TestVendorPerformanceDomainUsesGuidFirstScorecardRoutes(t *testing.T) {
	domain := findDomain("vendor-performance")
	if domain == nil {
		t.Fatal("expected vendor-performance domain")
	}

	actions := make(map[string]Action, len(domain.Actions))
	for _, action := range domain.Actions {
		actions[action.Name] = action
	}

	for _, name := range []string{"scorecard", "events", "trends", "recalculate"} {
		action, ok := actions[name]
		if !ok {
			t.Fatalf("missing %s action", name)
		}
		if len(action.Args) != 1 ||
			action.Args[0].Name != "vendorGuid" ||
			action.Args[0].Type != "uuid" ||
			!action.Args[0].Required {
			t.Fatalf("%s must require one vendorGuid UUID, got %+v", name, action.Args)
		}
	}

	if actions["scorecard"].ToolName != "UteamupVendorScorecardGet" {
		t.Fatalf("scorecard tool = %q", actions["scorecard"].ToolName)
	}
	if actions["config"].ToolName != "UteamupVendorScorecardConfigGet" ||
		len(actions["config"].Args) != 0 {
		t.Fatalf("config action must be tenant-scoped without identifiers: %+v", actions["config"])
	}

	for _, action := range domain.Actions {
		for _, argument := range action.Args {
			if argument.Name == "id" || argument.Name == "vendorId" || argument.Type == "int" {
				t.Fatalf("%s leaks integer identity argument %+v", action.Name, argument)
			}
		}
	}
}
