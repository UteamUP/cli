package registry

import (
	"strings"
	"testing"
)

func supplierInvoiceAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("supplier-invoice")
	if domain == nil {
		t.Fatal("supplier-invoice domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("supplier-invoice action %q is not registered", name)
	return nil, Action{}
}

func TestSupplierInvoiceRoutesAndArgumentsAreGuidOnly(t *testing.T) {
	t.Parallel()
	tests := []struct {
		actionName string
		path       string
	}{
		{"get", "/api/stock/supplier-invoices/invoice-guid"},
		{"match-preview", "/api/stock/supplier-invoices/invoice-guid/match-preview"},
		{"match-prepare", "/api/stock/supplier-invoices/invoice-guid/match-runs"},
	}
	for _, test := range tests {
		t.Run(test.actionName, func(t *testing.T) {
			domain, action := supplierInvoiceAction(t, test.actionName)
			path, consumed := buildRESTPath(domain, action, map[string]any{
				"invoiceGuid": "invoice-guid",
			})
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			if len(consumed) != 1 || len(action.Args) != 1 {
				t.Fatalf("route must consume exactly one GUID: args=%+v consumed=%v", action.Args, consumed)
			}
			argument := action.Args[0]
			if argument.Name != "invoiceGuid" || argument.Type != "uuid" ||
				strings.Contains(strings.ToLower(action.RESTPath), "{id}") {
				t.Fatalf("public invoice identity is not GUID-only: %+v", argument)
			}
		})
	}
}

func TestSupplierInvoiceActionsMirrorBackendToolsAndGovernance(t *testing.T) {
	t.Parallel()
	expected := map[string]string{
		"list":          "UteamupStockListSupplierInvoices",
		"get":           "UteamupStockGetSupplierInvoice",
		"create":        "UteamupStockCreateSupplierInvoice",
		"match-preview": "UteamupStockPreviewSupplierInvoiceMatch",
		"match-prepare": "UteamupStockPrepareSupplierInvoiceMatchRun",
	}
	for actionName, toolName := range expected {
		_, action := supplierInvoiceAction(t, actionName)
		if action.ToolName != toolName {
			t.Fatalf("%s tool = %q, want %q", actionName, action.ToolName, toolName)
		}
		for _, flag := range action.Flags {
			lower := strings.ToLower(flag.Name)
			if strings.Contains(lower, "tenant") || strings.Contains(lower, "user") {
				t.Fatalf("%s exposes caller-controlled tenant or user scope: %+v", actionName, flag)
			}
		}
	}

	_, list := supplierInvoiceAction(t, "list")
	for _, flag := range list.Flags {
		if flag.Name != "match-status" {
			continue
		}
		if !strings.Contains(flag.Description, "PendingReview") ||
			!strings.Contains(flag.Description, "Confirmed") {
			t.Fatalf("match-status must document the backend statuses PendingReview and Confirmed: %q", flag.Description)
		}
		if strings.Contains(flag.Description, "Unmatched") || strings.Contains(flag.Description, "NeedsReview") {
			t.Fatalf("match-status documents statuses the backend never stores: %q", flag.Description)
		}
	}

	_, preview := supplierInvoiceAction(t, "match-preview")
	if preview.HTTPMethod != "" {
		t.Fatalf("match preview must remain GET/read-only: %+v", preview)
	}
	_, prepare := supplierInvoiceAction(t, "match-prepare")
	if prepare.HTTPMethod != "POST" {
		t.Fatalf("match preparation must be explicit POST: %+v", prepare)
	}
	assertSupplierInvoiceFlag(t, prepare, "idempotency-guid", "idempotencyGuid", true, "uuid")
	assertSupplierInvoiceFlag(t, prepare, "conversation-guid", "conversationGuid", false, "uuid")
}

func assertSupplierInvoiceFlag(
	t *testing.T,
	action Action,
	name string,
	bodyName string,
	required bool,
	flagType string,
) {
	t.Helper()
	for _, flag := range action.Flags {
		if flag.Name == name {
			if flag.BodyName != bodyName || flag.Required != required || flag.Type != flagType {
				t.Fatalf("%s flag = %+v", name, flag)
			}
			return
		}
	}
	t.Fatalf("%s flag is missing", name)
}
