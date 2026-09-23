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

	assertGUIDFlag(t, domain, "assess", "asset-guid", true)
	assertGUIDFlag(t, domain, "get", "asset-guid", true)
	assertGUIDFlag(t, domain, "history", "asset-guid", true)
	assertGUIDFlag(t, domain, "matrix", "location-guid", false)
	assertGUIDFlag(t, domain, "matrix", "asset-type-guid", false)
}

func TestCriticalityAssessAcceptsConsequenceBreakdown(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "criticality" {
			domain = candidate
		}
	}
	if domain == nil {
		t.Fatal("criticality domain not registered")
	}
	for _, action := range domain.Actions {
		if action.Name != "assess" {
			continue
		}
		for _, flag := range action.Flags {
			if flag.Name == "consequence-breakdown-json" {
				if flag.Type != "string" || flag.Required {
					t.Fatalf("consequence-breakdown-json = %+v, want an optional string", flag)
				}
				if got := toCamelCase(flag.Name); got != "consequenceBreakdownJson" {
					t.Fatalf("body field = %q, want consequenceBreakdownJson", got)
				}
				return
			}
		}
		t.Fatal("assess has no consequence-breakdown-json flag")
	}
	t.Fatal("criticality has no assess action")
}
