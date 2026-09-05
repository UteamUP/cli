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
		"browse":                   "UteamupMarketplaceBrowse",
		"listing-get":              "UteamupMarketplaceListingGet",
		"listing-report":           "UteamupMarketplaceListingReport",
		"list-from-stock":          "UteamupMarketplaceListFromStock",
		"messages-list":            "UteamupMarketplaceMessagesList",
		"message-send":             "UteamupMarketplaceMessageSend",
		"message-thread":           "UteamupMarketplaceMessageThreadGet",
		"requirements":             "UteamupMarketplaceRequirementsList",
		"requirement-get":          "UteamupMarketplaceRequirementGet",
		"requirement-draft-update": "UteamupMarketplaceRequirementUpdateDraft",
		"my-offers":                "UteamupMarketplaceMyOffersList",
		"transactions":             "UteamupMarketplaceTransactionsList",
		"settings":                 "UteamupMarketplaceSettingsGet",
		"saved-searches":           "UteamupMarketplaceSavedSearchesList",
		"save-search":              "UteamupMarketplaceSaveSearch",
		"delete-saved-search":      "UteamupMarketplaceDeleteSavedSearch",
		"seller-scorecard":         "UteamupMarketplaceSellerScorecard",
		"facets":                   "UteamupMarketplaceFacets",
		"buyer-reputation":         "UteamupMarketplaceBuyerReputation",
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

func TestMarketplaceDraftUpdateRequiresReviewedVersionAndFullCoreFields(t *testing.T) {
	d := findMarketplaceDomain()
	if d == nil {
		t.Fatal("expected marketplace domain")
	}
	for _, action := range d.Actions {
		if action.Name != "requirement-draft-update" {
			continue
		}
		required := map[string]bool{}
		for _, flag := range action.Flags {
			required[flag.Name] = flag.Required
		}
		for _, name := range []string{"draft-version", "item-name", "item-type", "requested-quantity", "currency", "audience"} {
			if !required[name] {
				t.Errorf("draft update must require %q", name)
			}
		}
		if len(action.Args) != 1 || action.Args[0].Type != "uuid" || !action.Args[0].Required {
			t.Error("draft update must identify one requirement by GUID")
		}
		return
	}
	t.Fatal("missing requirement-draft-update action")
}

func TestMarketplacePublishRequiresReviewedDraftVersion(t *testing.T) {
	d := findMarketplaceDomain()
	if d == nil {
		t.Fatal("expected marketplace domain")
	}
	for _, action := range d.Actions {
		if action.Name == "requirement-publish" {
			for _, flag := range action.Flags {
				if flag.Name == "draft-version" && flag.Required && flag.Type == "string" {
					return
				}
			}
		}
	}
	t.Fatal("marketplace publication must require the reviewed draft version")
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
	var sellerGUID *FlagDef
	for i := range scorecard.Flags {
		if scorecard.Flags[i].Name == "seller-guid" {
			sellerGUID = &scorecard.Flags[i]
		}
	}
	if sellerGUID == nil {
		t.Fatal("seller-scorecard must define the \"seller-guid\" flag")
	}
	if !sellerGUID.Required {
		t.Error("seller-scorecard must require the \"seller-guid\" flag")
	}
	if sellerGUID.Type != "string" {
		t.Errorf("seller-scorecard \"seller-guid\" flag type is %q, want \"string\"", sellerGUID.Type)
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

// The ToolName assertions above are metadata — the CLI calls REST, so they stayed green
// while every action in this domain 404'd on GET /api/marketplace. These pin the route
// that actually gets requested, against the real controller paths.
func TestMarketplaceActionsResolveToRealRoutes(t *testing.T) {
	d := findMarketplaceDomain()
	if d == nil {
		t.Fatal("expected marketplace domain to be registered")
	}

	args := map[string]any{
		"guid":            "LISTING",
		"stockItemGuid":   "STOCK",
		"requirementGuid": "REQ",
		"offerGuid":       "OFFER",
		"sellerGuid":      "SELLER",
		"buyerGuid":       "BUYER",
	}
	want := map[string]string{
		"browse":                   "/api/marketplace/listings",
		"listing-get":              "/api/marketplace/listings/LISTING",
		"listing-report":           "/api/marketplace/listings/LISTING/report",
		"list-from-stock":          "/api/marketplace/listings/from-stock/STOCK",
		"message-send":             "/api/marketplace/messages",
		"message-thread":           "/api/marketplace/messages/LISTING/thread",
		"requirements":             "/api/marketplace/requirements/open",
		"requirement-draft-create": "/api/marketplace/requirements",
		"requirement-get":          "/api/marketplace/requirements/REQ",
		"requirement-draft-update": "/api/marketplace/requirements/REQ/draft",
		"requirement-publish":      "/api/marketplace/requirements/REQ/publish",
		"requirement-offer-quote":  "/api/marketplace/requirements/REQ/offers/OFFER/quote",
		"requirement-offer-accept": "/api/marketplace/requirements/REQ/offers/OFFER/accept",
		"my-offers":                "/api/marketplace/requirements/my-offers",
		"transactions":             "/api/marketplace/transactions",
		"settings":                 "/api/marketplace/settings",
		"saved-searches":           "/api/marketplace/saved-searches",
		"save-search":              "/api/marketplace/saved-searches",
		"delete-saved-search":      "/api/marketplace/saved-searches/LISTING",
		"seller-scorecard":         "/api/marketplace/sellers/SELLER/scorecard",
		"facets":                   "/api/marketplace/facets",
		"buyer-reputation":         "/api/marketplace/buyers/BUYER/reputation",
	}
	// No REST adapter serves these; they go over the tools/call transport instead.
	mcpOnly := map[string]bool{"messages-list": true, "requirement-offers-compare": true}

	for _, a := range d.Actions {
		if mcpOnly[a.Name] {
			if !a.MCPOnly {
				t.Errorf("action %q must be MCPOnly — no REST route serves it", a.Name)
			}
			continue
		}
		if a.MCPOnly {
			t.Errorf("action %q is MCPOnly but a REST route exists", a.Name)
			continue
		}
		expect, ok := want[a.Name]
		if !ok {
			t.Errorf("action %q has no expected route — add it here and to the controller map", a.Name)
			continue
		}
		got, _ := buildRESTPath(d, a, args)
		if got != expect {
			t.Errorf("action %q routes to %q, want %q", a.Name, got, expect)
		}
		// The original defect: every action collapsed onto the domain root.
		if got == "/api/marketplace" {
			t.Errorf("action %q fell back to the bare domain path — no controller serves it", a.Name)
		}
	}
}

func TestMarketplaceNonGETActionsDeclareTheirMethod(t *testing.T) {
	d := findMarketplaceDomain()
	if d == nil {
		t.Fatal("expected marketplace domain to be registered")
	}
	// Action names here match none of the HTTPMethod map keys, so anything that
	// writes must say so or it silently ships as a GET.
	want := map[string]string{
		"listing-report":           "POST",
		"list-from-stock":          "POST",
		"message-send":             "POST",
		"requirement-draft-create": "POST",
		"requirement-draft-update": "PUT",
		"requirement-publish":      "POST",
		"requirement-offer-accept": "POST",
		"save-search":              "POST",
		"delete-saved-search":      "DELETE",
	}
	for _, a := range d.Actions {
		if method, ok := want[a.Name]; ok && a.HTTPMethod != method {
			t.Errorf("action %q declares HTTPMethod %q, want %q", a.Name, a.HTTPMethod, method)
		}
		if _, ok := want[a.Name]; !ok && a.HTTPMethod != "" {
			t.Errorf("action %q declares HTTPMethod %q but should default to GET", a.Name, a.HTTPMethod)
		}
	}
}

func TestMarketplaceMyOffersSupportsBoundedFilteredReads(t *testing.T) {
	domain := findMarketplaceDomain()
	if domain == nil {
		t.Fatal("expected marketplace domain")
	}
	for _, action := range domain.Actions {
		if action.Name != "my-offers" {
			continue
		}
		flags := map[string]bool{}
		for _, flag := range action.Flags {
			flags[flag.Name] = true
		}
		for _, expected := range []string{"page", "page-size", "status", "search"} {
			if !flags[expected] {
				t.Errorf("missing supplier offer flag %s", expected)
			}
		}
		return
	}
	t.Fatal("missing my-offers action")
}
