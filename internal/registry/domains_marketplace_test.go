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
		"browse":              "UteamupMarketplaceBrowse",
		"listing-get":         "UteamupMarketplaceListingGet",
		"listing-report":      "UteamupMarketplaceListingReport",
		"messages-list":       "UteamupMarketplaceMessagesList",
		"message-send":        "UteamupMarketplaceMessageSend",
		"message-thread":      "UteamupMarketplaceMessageThreadGet",
		"requirements":        "UteamupMarketplaceRequirementsList",
		"my-offers":           "UteamupMarketplaceMyOffersList",
		"transactions":        "UteamupMarketplaceTransactionsList",
		"settings":            "UteamupMarketplaceSettingsGet",
		"saved-searches":      "UteamupMarketplaceSavedSearchesList",
		"save-search":         "UteamupMarketplaceSaveSearch",
		"delete-saved-search": "UteamupMarketplaceDeleteSavedSearch",
		"seller-scorecard":    "UteamupMarketplaceSellerScorecard",
		"facets":              "UteamupMarketplaceFacets",
		"buyer-reputation":    "UteamupMarketplaceBuyerReputation",
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

func TestMarketplaceListingReportFlags(t *testing.T) {
	d := findMarketplaceDomain()
	if d == nil {
		t.Fatal("expected marketplace domain to be registered")
	}
	var report *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "listing-report" {
			report = &d.Actions[i]
		}
	}
	if report == nil {
		t.Fatal("missing marketplace action \"listing-report\"")
	}
	required := map[string]bool{}
	for _, f := range report.Flags {
		if f.Required {
			required[f.Name] = true
		}
	}
	for _, want := range []string{"guid", "reason"} {
		if !required[want] {
			t.Errorf("listing-report must require the %q flag", want)
		}
	}
}

func TestMarketplaceSavedSearchFlags(t *testing.T) {
	d := findMarketplaceDomain()
	if d == nil {
		t.Fatal("expected marketplace domain to be registered")
	}
	byName := map[string]Action{}
	for _, a := range d.Actions {
		byName[a.Name] = a
	}

	save, ok := byName["save-search"]
	if !ok {
		t.Fatal("missing marketplace action \"save-search\"")
	}
	var nameRequired bool
	var notifyDefault any
	var notifyType string
	for _, f := range save.Flags {
		if f.Name == "name" && f.Required {
			nameRequired = true
		}
		if f.Name == "notify-on-new-match" {
			notifyDefault = f.Default
			notifyType = f.Type
		}
	}
	if !nameRequired {
		t.Error("save-search must require the \"name\" flag")
	}
	if notifyType != "bool" {
		t.Errorf("save-search \"notify-on-new-match\" flag type is %q, want \"bool\"", notifyType)
	}
	if v, ok := notifyDefault.(bool); !ok || !v {
		t.Errorf("save-search \"notify-on-new-match\" default is %v (%T), want true (bool)", notifyDefault, notifyDefault)
	}

	del, ok := byName["delete-saved-search"]
	if !ok {
		t.Fatal("missing marketplace action \"delete-saved-search\"")
	}
	var guidRequired bool
	for _, f := range del.Flags {
		if f.Name == "guid" && f.Required {
			guidRequired = true
		}
	}
	if !guidRequired {
		t.Error("delete-saved-search must require the \"guid\" flag")
	}
}

func TestMarketplaceSellerScorecardFlags(t *testing.T) {
	d := findMarketplaceDomain()
	if d == nil {
		t.Fatal("expected marketplace domain to be registered")
	}
	var scorecard *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "seller-scorecard" {
			scorecard = &d.Actions[i]
		}
	}
	if scorecard == nil {
		t.Fatal("missing marketplace action \"seller-scorecard\"")
	}
	var sellerGuid *FlagDef
	for i := range scorecard.Flags {
		if scorecard.Flags[i].Name == "seller-guid" {
			sellerGuid = &scorecard.Flags[i]
		}
	}
	if sellerGuid == nil {
		t.Fatal("seller-scorecard must define the \"seller-guid\" flag")
	}
	if !sellerGuid.Required {
		t.Error("seller-scorecard must require the \"seller-guid\" flag")
	}
	if sellerGuid.Type != "string" {
		t.Errorf("seller-scorecard \"seller-guid\" flag type is %q, want \"string\"", sellerGuid.Type)
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
