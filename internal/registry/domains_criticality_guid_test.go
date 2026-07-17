package registry

import "testing"

func TestCriticalityDomainIsUniqueAndGuidFirst(t *testing.T) {
	var matchingDomains []*Domain
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "criticality" || domain.Name == "asset-criticality" {
			matchingDomains = append(matchingDomains, domain)
		}
	}
	if len(matchingDomains) != 1 {
		t.Fatalf("criticality domains = %d, want exactly one", len(matchingDomains))
	}

	domain := matchingDomains[0]
	for _, action := range domain.Actions {
		for _, flag := range action.Flags {
			switch flag.Name {
			case "asset-id", "location-id", "asset-type-id", "id":
				t.Fatalf("%s exposes forbidden integer identity flag %q", action.Name, flag.Name)
			}
		}
	}

	assertGuidFlag(t, domain, "assess", "asset-guid", true)
	assertGuidFlag(t, domain, "get", "asset-guid", true)
	assertGuidFlag(t, domain, "history", "asset-guid", true)
	assertGuidFlag(t, domain, "matrix", "location-guid", false)
	assertGuidFlag(t, domain, "matrix", "asset-type-guid", false)
}
