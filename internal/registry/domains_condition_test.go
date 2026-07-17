package registry

import "testing"

func TestConditionDomainIsUniqueAndGuidFirst(t *testing.T) {
	var matchingDomains []*Domain
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "condition" || domain.Name == "asset-condition" {
			matchingDomains = append(matchingDomains, domain)
		}
	}
	if len(matchingDomains) != 1 {
		t.Fatalf("condition domains = %d, want exactly one", len(matchingDomains))
	}

	domain := matchingDomains[0]
	for _, action := range domain.Actions {
		for _, flag := range action.Flags {
			switch flag.Name {
			case "asset-id", "location-id", "id":
				t.Fatalf("%s exposes forbidden integer identity flag %q", action.Name, flag.Name)
			}
		}
	}

	assertGuidFlag(t, domain, "assess", "asset-guid", true)
	assertGuidFlag(t, domain, "get", "asset-guid", true)
	assertGuidFlag(t, domain, "history", "asset-guid", true)
	assertGuidFlag(t, domain, "heat-map", "location-guid", false)
}

func assertGuidFlag(
	t *testing.T,
	domain *Domain,
	actionName string,
	flagName string,
	required bool,
) {
	t.Helper()
	for _, action := range domain.Actions {
		if action.Name != actionName {
			continue
		}
		for _, flag := range action.Flags {
			if flag.Name == flagName {
				if flag.Type != "string" || flag.Required != required {
					t.Fatalf("%s %s flag = %+v", actionName, flagName, flag)
				}
				return
			}
		}
		t.Fatalf("%s is missing %s", actionName, flagName)
	}
	t.Fatalf("condition action %s not found", actionName)
}
