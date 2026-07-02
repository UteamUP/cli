package registry

import "testing"

func findMarketplaceDomain() *Domain {
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "marketplace" {
			return d
		}
	}
	return nil
}

func TestMarketplaceDomainRegistered(t *testing.T) {
	d := findMarketplaceDomain()
	if d == nil {
		t.Fatal("expected marketplace domain to be registered")
	}
	if d.Description == "" {
		t.Error("marketplace domain must have a Description")
	}
	if len(d.Aliases) == 0 {
		t.Error("marketplace domain should have aliases")
	}
}

func TestMarketplaceActionsWired(t *testing.T) {
	d := findMarketplaceDomain()
	if d == nil {
		t.Fatal("expected marketplace domain to be registered")
	}
	expected := map[string]string{
		"browse":       "UteamupMarketplaceBrowse",
		"listing-get":  "UteamupMarketplaceListingGet",
		"requirements": "UteamupMarketplaceRequirementsList",
		"my-offers":    "UteamupMarketplaceMyOffersList",
		"transactions": "UteamupMarketplaceTransactionsList",
		"settings":     "UteamupMarketplaceSettingsGet",
	}
	actions := map[string]Action{}
	for _, a := range d.Actions {
		actions[a.Name] = a
	}
	for name, tool := range expected {
		a, ok := actions[name]
		if !ok {
			t.Errorf("missing marketplace action %q", name)
			continue
		}
		if a.ToolName != tool {
			t.Errorf("action %q maps to %q, want %q", name, a.ToolName, tool)
		}
	}
}

// Float flag defaults must be float literals — an untyped int default panics the
// registry's type assertion at command-build time.
func TestMarketplaceFloatDefaultsAreFloats(t *testing.T) {
	d := findMarketplaceDomain()
	if d == nil {
		t.Fatal("expected marketplace domain to be registered")
	}
	for _, a := range d.Actions {
		for _, f := range a.Flags {
			if f.Type == "float" && f.Default != nil {
				if _, ok := f.Default.(float64); !ok {
					t.Errorf("action %q flag %q: float default is %T, want float64", a.Name, f.Name, f.Default)
				}
			}
		}
	}
}
