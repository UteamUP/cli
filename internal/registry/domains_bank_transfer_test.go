package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
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

func TestBankTransferDomainDoesNotExposeLegacyTenantOverview(t *testing.T) {
	t.Parallel()
	domain := findDomain("bank-transfer")
	if domain == nil {
		t.Fatal("bank-transfer domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == "status" || action.ToolName == "UteamupBankTransferBillingOverview" {
			t.Fatalf("legacy provider-specific billing overview is registered: %+v", action)
		}
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

func TestBankTransferListSubscriptionsPreservesGuidOnlyResponse(t *testing.T) {
	domain, action := bankTransferAction(t, "list-subscriptions")
	path, _ := buildRESTPath(domain, action, nil)
	const tenantGUID = "11111111-1111-4111-8111-111111111111"
	const subscriptionGUID = "22222222-2222-4222-8222-222222222222"
	const planGUID = "33333333-3333-4333-8333-333333333333"
	const pendingPlanGUID = "44444444-4444-4444-8444-444444444444"

	apiClient := newBankTransferContractClient(t, tenantGUID, func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodGet || request.URL.Path != path {
			t.Errorf("request = %s %s, want GET %s", request.Method, request.URL.Path, path)
		}
		if request.Header.Get("X-Tenant-Guid") != tenantGUID {
			t.Errorf("tenant GUID header = %q", request.Header.Get("X-Tenant-Guid"))
		}
		if request.Header.Get("X-Tenant-ID") != "" {
			t.Errorf("integer tenant identity leaked in header: %q", request.Header.Get("X-Tenant-ID"))
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`[{
			"guid":"` + subscriptionGUID + `",
			"tenantGuid":"` + tenantGUID + `",
			"planGuid":"` + planGUID + `",
			"pendingPlanGuid":"` + pendingPlanGUID + `",
			"status":"Active"
		}]`))
	})

	result, err := apiClient.CallREST(
		context.Background(),
		action.HTTPMethod,
		path,
		nil,
		nil,
		action.Name,
	)
	if err != nil {
		t.Fatalf("list-subscriptions transport failed: %v", err)
	}
	var subscriptions []map[string]json.RawMessage
	if err := json.Unmarshal(result, &subscriptions); err != nil || len(subscriptions) != 1 {
		t.Fatalf("decode subscription response: %v (%s)", err, result)
	}
	for _, field := range []string{"guid", "tenantGuid", "planGuid", "pendingPlanGuid"} {
		if _, ok := subscriptions[0][field]; !ok {
			t.Errorf("subscription response is missing %q: %s", field, result)
		}
	}
	for _, retired := range []string{"id", "tenantId", "planId", "pendingPlanId"} {
		if _, ok := subscriptions[0][retired]; ok {
			t.Errorf("subscription response leaked retired field %q: %s", retired, result)
		}
	}
}

func newBankTransferContractClient(
	t *testing.T,
	tenantGUID string,
	handler http.HandlerFunc,
) *client.APIClient {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := auth.SaveToken(&auth.TokenData{
		AccessToken: "bank-transfer-contract-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  tenantGUID,
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	return client.NewAPIClient(
		server.URL,
		time.Second,
		true,
		client.RetryOptions{MaxRetries: 0},
		logging.New(logging.LevelError),
	)
}
