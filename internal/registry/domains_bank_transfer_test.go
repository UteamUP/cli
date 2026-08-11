package registry

import (
	"strings"
	"testing"
)

func bankTransferAction(t *testing.T, actionName string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("bank-transfer")
	if domain == nil {
		t.Fatal("bank-transfer domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == actionName {
			return domain, action
		}
	}
	t.Fatalf("bank-transfer action %q is not registered", actionName)
	return nil, Action{}
}

func TestBankTransferRoutesAreGuidOnly(t *testing.T) {
	t.Parallel()
	tests := []struct {
		actionName string
		argName    string
		path       string
	}{
		{"get-invoice", "invoiceGuid", "/api/internalbilling/invoices/invoice-guid"},
		{"mark-paid", "invoiceGuid", "/api/internalbilling/admin/invoices/invoice-guid/mark-paid"},
		{"activate", "subscriptionGuid", "/api/internalbilling/admin/subscriptions/subscription-guid/activate"},
		{"refund", "paymentGuid", "/api/internalbilling/admin/payments/payment-guid/refund"},
	}
	for _, test := range tests {
		t.Run(test.actionName, func(t *testing.T) {
			domain, action := bankTransferAction(t, test.actionName)
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

func TestBankTransferReadSurfacesTargetAdminRoutes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		actionName string
		path       string
	}{
		{"list-invoices", "/api/internalbilling/admin/invoices/pending"},
		{"list-overdue", "/api/internalbilling/admin/invoices/overdue"},
		{"list-paid", "/api/internalbilling/admin/invoices/paid"},
		{"list-subscriptions", "/api/internalbilling/admin/subscriptions"},
		{"dashboard", "/api/internalbilling/admin/dashboard"},
	}
	for _, test := range tests {
		t.Run(test.actionName, func(t *testing.T) {
			domain, action := bankTransferAction(t, test.actionName)
			if action.HTTPMethod != "GET" {
				t.Fatalf("read surface %s must be GET, got %q", test.actionName, action.HTTPMethod)
			}
			path, _ := buildRESTPath(domain, action, map[string]any{})
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
		})
	}
}

func TestBankTransferNoActionCarriesIntIdentifier(t *testing.T) {
	t.Parallel()
	domain := findDomain("bank-transfer")
	if domain == nil {
		t.Fatal("bank-transfer domain is not registered")
	}
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Type == "int" {
				t.Fatalf("action %s declares int arg %s — public identifiers must be GUIDs", action.Name, arg.Name)
			}
		}
	}
}

func TestBankTransferReconciliationBodiesMirrorBackendModels(t *testing.T) {
	t.Parallel()
	_, markPaid := bankTransferAction(t, "mark-paid")
	wantMarkPaid := map[string]bool{"amount": false, "paymentDate": false, "bankReference": false, "adminNotes": false}
	for _, flag := range markPaid.Flags {
		if _, ok := wantMarkPaid[flag.BodyName]; ok {
			wantMarkPaid[flag.BodyName] = true
		}
		if flag.BodyName == "paymentDate" && !flag.Required {
			t.Fatal("paymentDate is required by MarkInvoicePaidRequestModel — the flag must be Required")
		}
	}
	for body, seen := range wantMarkPaid {
		if !seen {
			t.Fatalf("mark-paid is missing body field %q from MarkInvoicePaidRequestModel", body)
		}
	}

	_, refund := bankTransferAction(t, "refund")
	var reasonRequired, amountOptional bool
	for _, flag := range refund.Flags {
		if flag.BodyName == "reason" && flag.Required {
			reasonRequired = true
		}
		if flag.BodyName == "amount" && !flag.Required {
			amountOptional = true
		}
	}
	if !reasonRequired {
		t.Fatal("refund must require --reason (RefundPaymentRequestModel.Reason is [Required])")
	}
	if !amountOptional {
		t.Fatal("refund --amount must stay optional (null = full remaining refundable amount)")
	}
}
