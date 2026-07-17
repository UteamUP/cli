package registry

import (
	"strings"
	"testing"
)

func serviceBillingAction(t *testing.T, domainName, actionName string) (*Domain, Action) {
	t.Helper()
	domain := findDomain(domainName)
	if domain == nil {
		t.Fatalf("%s domain is not registered", domainName)
	}
	for _, action := range domain.Actions {
		if action.Name == actionName {
			return domain, action
		}
	}
	t.Fatalf("%s action %q is not registered", domainName, actionName)
	return nil, Action{}
}

func TestServiceBillingRoutesAndArgumentsAreGuidOnly(t *testing.T) {
	t.Parallel()
	tests := []struct {
		domainName string
		actionName string
		argName    string
		path       string
	}{
		{"service-billing", "get", "runGuid", "/api/service-billing-runs/run-guid"},
		{"service-billing", "approve", "runGuid", "/api/service-billing-runs/run-guid/approve"},
		{"service-billing", "cancel", "runGuid", "/api/service-billing-runs/run-guid/cancel"},
		{"service-invoice", "get", "invoiceGuid", "/api/service-invoices/invoice-guid"},
		{"service-invoice", "issue", "invoiceGuid", "/api/service-invoices/invoice-guid/issue"},
		{"service-invoice", "send", "invoiceGuid", "/api/service-invoices/invoice-guid/send"},
		{"service-invoice", "paid", "invoiceGuid", "/api/service-invoices/invoice-guid/paid"},
		{"service-invoice", "void", "invoiceGuid", "/api/service-invoices/invoice-guid/void"},
		{"service-invoice", "credit", "invoiceGuid", "/api/service-invoices/invoice-guid/credit"},
		{"service-invoice", "accounting-exports", "invoiceGuid", "/api/service-invoices/invoice-guid/exports"},
		{"service-invoice", "accounting-export", "invoiceGuid", "/api/service-invoices/invoice-guid/exports"},
	}
	for _, test := range tests {
		t.Run(test.domainName+"-"+test.actionName, func(t *testing.T) {
			domain, action := serviceBillingAction(t, test.domainName, test.actionName)
			path, consumed := buildRESTPath(domain, action, map[string]any{
				test.argName: strings.ReplaceAll(test.argName, "Guid", "-guid"),
			})
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			if len(consumed) != 1 || strings.Contains(strings.ToLower(action.RESTPath), "{id}") {
				t.Fatalf("route is not GUID-only: %q consumed=%v", action.RESTPath, consumed)
			}
			if len(action.Args) != 1 || action.Args[0].Name != test.argName || action.Args[0].Type != "uuid" {
				t.Fatalf("public identity argument is not a UUID: %+v", action.Args)
			}
		})
	}
}

func TestServiceBillingActionsMirrorBackendToolsAndReviewedEvidence(t *testing.T) {
	t.Parallel()
	expectedTools := map[string]map[string]string{
		"service-billing": {
			"list": "UteamupServiceBillingRunList", "get": "UteamupServiceBillingRunGet",
			"create": "UteamupServiceBillingRunCreate", "approve": "UteamupServiceBillingRunApprove",
			"cancel": "UteamupServiceBillingRunCancel",
		},
		"service-invoice": {
			"list": "UteamupServiceInvoiceList", "get": "UteamupServiceInvoiceGet",
			"issue": "UteamupServiceInvoiceIssue", "send": "UteamupServiceInvoiceSend",
			"paid": "UteamupServiceInvoicePaid", "void": "UteamupServiceInvoiceVoid",
			"credit":                      "UteamupServiceInvoiceCredit",
			"accounting-exports":          "UteamupServiceAccountingExportList",
			"accounting-export":           "UteamupServiceAccountingExportCreate",
			"accounting-export-retry":     "UteamupServiceAccountingExportRetry",
			"accounting-export-reconcile": "UteamupServiceAccountingExportReconcile",
			"accounting-export-download":  "UteamupServiceAccountingExportDownload",
		},
	}
	for domainName, actions := range expectedTools {
		for actionName, toolName := range actions {
			_, action := serviceBillingAction(t, domainName, actionName)
			if action.ToolName != toolName {
				t.Fatalf("%s %s tool = %q, want %q", domainName, actionName, action.ToolName, toolName)
			}
			for _, flag := range action.Flags {
				lower := strings.ToLower(flag.Name)
				if strings.Contains(lower, "tenant") || strings.Contains(lower, "user") {
					t.Fatalf("%s exposes caller-controlled scope: %+v", actionName, flag)
				}
			}
		}
	}

	for _, actionName := range []string{"create", "approve", "cancel"} {
		_, action := serviceBillingAction(t, "service-billing", actionName)
		assertServiceBillingFlag(t, action, "idempotency-key", "idempotencyKey", true)
	}
	for _, actionName := range []string{"issue", "send", "paid", "void", "credit"} {
		_, action := serviceBillingAction(t, "service-invoice", actionName)
		assertServiceBillingFlag(t, action, "idempotency-key", "idempotencyKey", true)
		assertServiceBillingFlag(t, action, "expected-updated-at", "expectedUpdatedAt", true)
	}
	_, issue := serviceBillingAction(t, "service-invoice", "issue")
	assertServiceBillingFlag(t, issue, "invoice-number", "invoiceNumber", true)
	assertServiceBillingFlag(t, issue, "issued-at", "issuedAt", false)
	for _, actionName := range []string{"void", "credit"} {
		_, action := serviceBillingAction(t, "service-invoice", actionName)
		assertServiceBillingFlag(t, action, "reason", "reason", true)
	}

	_, accountingExport := serviceBillingAction(t, "service-invoice", "accounting-export")
	assertServiceBillingFlag(
		t,
		accountingExport,
		"idempotency-key",
		"idempotencyKey",
		true)
	assertServiceBillingFlag(
		t,
		accountingExport,
		"expected-invoice-updated-at",
		"expectedInvoiceUpdatedAt",
		true)
	assertServiceBillingFlag(t, accountingExport, "connector", "connectorKey", false)

	_, accountingRetry := serviceBillingAction(
		t,
		"service-invoice",
		"accounting-export-retry")
	assertServiceBillingFlag(
		t,
		accountingRetry,
		"idempotency-key",
		"idempotencyKey",
		true)

	_, accountingReconcile := serviceBillingAction(
		t,
		"service-invoice",
		"accounting-export-reconcile")
	assertServiceBillingFlag(
		t,
		accountingReconcile,
		"external-reference",
		"externalReference",
		true)
	assertServiceBillingFlag(t, accountingReconcile, "status", "status", true)
	assertServiceBillingFlag(t, accountingReconcile, "reason", "reason", false)
}

func TestServiceAccountingExportRoutesUseOnlyGuidArguments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		actionName string
		path       string
	}{
		{
			"accounting-export-retry",
			"/api/service-invoices/invoice-guid/exports/export-guid/retry",
		},
		{
			"accounting-export-reconcile",
			"/api/service-invoices/invoice-guid/exports/export-guid/reconcile",
		},
		{
			"accounting-export-download",
			"/api/service-invoices/invoice-guid/exports/export-guid/content",
		},
	}

	for _, test := range tests {
		t.Run(test.actionName, func(t *testing.T) {
			_, action := serviceBillingAction(t, "service-invoice", test.actionName)
			path, consumed := buildRESTPath(
				findDomain("service-invoice"),
				action,
				map[string]any{
					"invoiceGuid": "invoice-guid",
					"exportGuid":  "export-guid",
				})
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			if len(consumed) != 2 || len(action.Args) != 2 {
				t.Fatalf(
					"accounting route must consume two GUIDs: args=%+v consumed=%v",
					action.Args,
					consumed)
			}
			for _, argument := range action.Args {
				if argument.Type != "uuid" || !strings.HasSuffix(argument.Name, "Guid") {
					t.Fatalf("public accounting identity is not a GUID: %+v", argument)
				}
			}
		})
	}
}

func assertServiceBillingFlag(t *testing.T, action Action, name, bodyName string, required bool) {
	t.Helper()
	for _, flag := range action.Flags {
		if flag.Name == name {
			if flag.BodyName != bodyName || flag.Required != required {
				t.Fatalf("%s flag = %+v, want body=%q required=%t", name, flag, bodyName, required)
			}
			return
		}
	}
	t.Fatalf("%s flag is missing from %s", name, action.Name)
}
