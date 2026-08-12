package registry

import "testing"

func TestBillingCommerceDomainMirrorsProviderNeutralReadTools(t *testing.T) {
	domain := findDomainByName(t, "billing")

	if domain.APIPath != "/api/billing" {
		t.Fatalf("APIPath = %q, want /api/billing", domain.APIPath)
	}
	want := map[string]struct {
		tool string
		path string
	}{
		"context":    {"UteamupBillingContextGet", "context"},
		"profile":    {"UteamupBillingProfileGet", "profile"},
		"catalog":    {"UteamupBillingCatalogList", "catalog"},
		"order":      {"UteamupBillingOrderGet", "orders/{orderGuid}"},
		"agreements": {"UteamupBillingAgreementsList", "agreements"},
		"invoices":   {"UteamupBillingInvoicesList", "invoices"},
		"invoice":    {"UteamupBillingInvoiceGet", "invoices/{invoiceGuid}"},
	}
	if len(domain.Actions) != len(want) {
		t.Fatalf("actions = %d, want %d", len(domain.Actions), len(want))
	}
	for _, action := range domain.Actions {
		expected, ok := want[action.Name]
		if !ok {
			t.Fatalf("unexpected action %q", action.Name)
		}
		if action.ToolName != expected.tool || action.RESTPath != expected.path {
			t.Fatalf("%s = (%q, %q), want (%q, %q)",
				action.Name, action.ToolName, action.RESTPath, expected.tool, expected.path)
		}
		if action.HTTPMethod != "GET" {
			t.Fatalf("%s method = %q, want GET", action.Name, action.HTTPMethod)
		}
	}
}

func TestBillingCommerceDomainUsesOnlyPublicGuidArguments(t *testing.T) {
	domain := findDomainByName(t, "billing")

	for _, action := range domain.Actions {
		for _, argument := range action.Args {
			if argument.Type != "uuid" {
				t.Fatalf("%s argument %s type = %q, want uuid",
					action.Name, argument.Name, argument.Type)
			}
			if argument.Name != "orderGuid" && argument.Name != "invoiceGuid" {
				t.Fatalf("%s exposes unsupported identity argument %q", action.Name, argument.Name)
			}
		}
	}
}
