package registry

import (
	"strings"
	"testing"
)

// billing-kling-remediation F5.3 — coverage for the billing/plan domains that shipped
// without a _test.go. Route assertions pin the URL contract for every domain, including
// admin-billing-gateway: its routing defect (no APIPath/RESTPath → phantom
// /api/adminbillinggateway that 404s) is now fixed against the real GlobalAdminController
// surface, so the paths are pinned here rather than left uncovered.
//
// Verifying a route against a running backend: probe it UNAUTHENTICATED and expect 401
// (the endpoint exists and challenges) versus 404 (no such endpoint). Do NOT probe with a
// valid token and read 403 as proof the route exists — TenantIdentificationMiddleware
// returns 403 before routing resolves, so a nonexistent path answers 403 identically.
//   curl -sk -o /dev/null -w '%{http_code}' -X POST -d '{}' \
//     https://localhost:5002/api/globaladmin/tenants/<guid>/billing-method    # 401 = exists

func billingPlanAction(t *testing.T, domainName, actionName string) (*Domain, Action) {
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

func TestBillingPlanDomainRoutes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		domainName string
		actionName string
		args       map[string]any
		path       string
	}{
		{"plan-analytics", "summary", map[string]any{}, "/api/plananalytics/summary"},
		{"plan-analytics", "insights", map[string]any{}, "/api/plananalytics/insights"},
		{"plan-approval", "pending", map[string]any{}, "/api/planapproval/pending"},
		{"plan-approval", "approve", map[string]any{"requestGuid": "req-guid"}, "/api/planapproval/by-guid/req-guid/approve"},
		{"plan-approval", "reject", map[string]any{"requestGuid": "req-guid"}, "/api/planapproval/by-guid/req-guid/reject"},
		{"plan-impact", "preview", map[string]any{"planGuid": "plan-guid"}, "/api/planimpact/by-plan/plan-guid/preview"},
		{"plan-limit", "list", map[string]any{"planGuid": "plan-guid"}, "/api/planlimit/by-plan/plan-guid"},
		{"plan-limit", "upsert", map[string]any{"planGuid": "plan-guid"}, "/api/planlimit/by-plan/plan-guid"},
		{"plan-migration", "migrate", map[string]any{}, "/api/planmigration"},
		{"subscription-lifecycle", "suspend", map[string]any{"guid": "sub-guid"}, "/api/internalbilling/admin/subscriptions/sub-guid/suspend"},
		{"subscription-lifecycle", "cancel", map[string]any{"guid": "sub-guid"}, "/api/internalbilling/admin/subscriptions/sub-guid/cancel"},
		{"subscription-lifecycle", "reactivate", map[string]any{"guid": "sub-guid"}, "/api/internalbilling/admin/subscriptions/sub-guid/reactivate"},
		{"subscription-lifecycle", "schedule-cancel", map[string]any{"guid": "sub-guid"}, "/api/internalbilling/admin/subscriptions/sub-guid/schedule-cancel"},
		{"subscription-lifecycle", "activate-without-payment", map[string]any{"guid": "sub-guid"}, "/api/internalbilling/admin/subscriptions/sub-guid/activate-without-payment"},
		{"subscription-lifecycle", "clear-scheduled-cancel", map[string]any{"guid": "sub-guid"}, "/api/internalbilling/admin/subscriptions/sub-guid/clear-scheduled-cancel"},
	}
	for _, test := range tests {
		t.Run(test.domainName+"-"+test.actionName, func(t *testing.T) {
			domain, action := billingPlanAction(t, test.domainName, test.actionName)
			path, _ := buildRESTPath(domain, action, test.args)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
		})
	}
}

func TestBillingPlanDomainsDeclareNoIntIdentityArg(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"admin-billing-gateway", "plan-analytics", "plan-approval", "plan-impact", "plan-limit", "plan-migration", "subscription-lifecycle"} {
		domain := findDomain(name)
		if domain == nil {
			t.Fatalf("%s domain is not registered", name)
		}
		for _, action := range domain.Actions {
			for _, arg := range action.Args {
				if arg.Type == "int" {
					t.Fatalf("%s %s declares int identity arg %s — public identifiers must be GUIDs", name, action.Name, arg.Name)
				}
			}
		}
	}
}

func TestBillingGatewaySwitcherRoutesMatchGlobalAdminController(t *testing.T) {
	t.Parallel()
	// Paths mirror GlobalAdminController: POST tenants/{tenantGuid:guid}/billing-method,
	// GET .../history, GET .../{auditGuid:guid}, POST .../{auditGuid:guid}/cancel.
	tests := []struct {
		actionName string
		method     string
		args       map[string]any
		path       string
		consumed   int
	}{
		{"change", "POST", map[string]any{"tenantGuid": "tenant-guid"},
			"/api/globaladmin/tenants/tenant-guid/billing-method", 1},
		{"history", "GET", map[string]any{"tenantGuid": "tenant-guid"},
			"/api/globaladmin/tenants/tenant-guid/billing-method/history", 1},
		{"get", "GET", map[string]any{"tenantGuid": "tenant-guid", "auditGuid": "audit-guid"},
			"/api/globaladmin/tenants/tenant-guid/billing-method/audit-guid", 2},
		{"cancel", "POST", map[string]any{"tenantGuid": "tenant-guid", "auditGuid": "audit-guid"},
			"/api/globaladmin/tenants/tenant-guid/billing-method/audit-guid/cancel", 2},
	}
	for _, test := range tests {
		t.Run(test.actionName, func(t *testing.T) {
			domain, action := billingPlanAction(t, "admin-billing-gateway", test.actionName)
			if action.HTTPMethod != test.method {
				t.Fatalf("method = %q, want %q — the HTTPMethod map has no entry for %q, so an "+
					"unset override silently falls back to GET", action.HTTPMethod, test.method, test.actionName)
			}
			path, consumed := buildRESTPath(domain, action, test.args)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			if len(consumed) != test.consumed {
				t.Fatalf("consumed %d path args, want %d — unconsumed identifiers leak into the body",
					len(consumed), test.consumed)
			}
		})
	}
}

func TestBillingGatewaySwitcherNeverDerivesThePhantomPath(t *testing.T) {
	t.Parallel()
	// The regression this domain shipped with: no APIPath + no RESTPath means buildRESTPath
	// derives "/api/" + domainName with dashes stripped — a controller that does not exist.
	domain := findDomain("admin-billing-gateway")
	if domain == nil {
		t.Fatal("admin-billing-gateway domain is not registered")
	}
	if domain.APIPath != "/api/globaladmin" {
		t.Fatalf("APIPath = %q, want /api/globaladmin", domain.APIPath)
	}
	for _, action := range domain.Actions {
		if action.RESTPath == "" {
			t.Fatalf("action %s declares no RESTPath", action.Name)
		}
		path, _ := buildRESTPath(domain, action, map[string]any{
			"tenantGuid": "t", "auditGuid": "a",
		})
		if strings.Contains(path, "adminbillinggateway") {
			t.Fatalf("action %s still resolves to the phantom derived path: %s", action.Name, path)
		}
		if strings.Contains(path, "{") {
			t.Fatalf("action %s left an unexpanded placeholder: %s", action.Name, path)
		}
	}
}

func TestBillingGatewaySwitcherFieldMappingMatchesBackendBinding(t *testing.T) {
	t.Parallel()
	// The backend binds the idempotency key with [FromHeader(Name = "Idempotency-Key")], so it
	// must travel as a header — a BodyName mapping would put it in the JSON body where nothing
	// reads it. Everything else maps onto BillingMethodChangeRequest / …CancelRequest fields.
	_, change := billingPlanAction(t, "admin-billing-gateway", "change")
	body := map[string]string{}
	var idempotencyHeader string
	for _, flag := range change.Flags {
		if flag.HeaderName != "" {
			idempotencyHeader = flag.HeaderName
			if flag.BodyName != "" {
				t.Fatalf("--%s must not also declare BodyName; the backend reads it from the header only", flag.Name)
			}
			continue
		}
		body[flag.Name] = flag.BodyName
	}
	if idempotencyHeader != "Idempotency-Key" {
		t.Fatalf("idempotency key header = %q, want Idempotency-Key", idempotencyHeader)
	}
	for flagName, want := range map[string]string{
		"tenant": "tenantGuid", "to": "newBillingMethod",
		"reason": "reason", "kennitala": "kennitala", "effective": "effective",
	} {
		if body[flagName] != want {
			t.Fatalf("--%s maps to %q, want %q", flagName, body[flagName], want)
		}
	}

	_, history := billingPlanAction(t, "admin-billing-gateway", "history")
	pagination := map[string]string{}
	for _, flag := range history.Flags {
		pagination[flag.Name] = flag.BodyName
	}
	if pagination["page"] != "page" || pagination["page-size"] != "pageSize" {
		t.Fatalf("history pagination maps to %v, want page/pageSize", pagination)
	}
}

func TestBillingGatewaySwitcherUsesBackendEnumVocabulary(t *testing.T) {
	t.Parallel()
	// The API registers JsonStringEnumConverter(camelCase, allowIntegerValues: true), so the
	// enum NAMES bind directly and the CLI translates nothing. The prior header comment
	// promised stripe/ibt/kling aliases that no code implemented; this pins the honest
	// vocabulary so the docs cannot drift back.
	_, change := billingPlanAction(t, "admin-billing-gateway", "change")
	for _, flag := range change.Flags {
		switch flag.Name {
		case "to":
			for _, member := range []string{"stripe", "icelandicBankTransfer"} {
				if !strings.Contains(flag.Description, member) {
					t.Fatalf("--to must document the %q enum member, got %q", member, flag.Description)
				}
			}
			for _, ghost := range []string{"ibt", "kling"} {
				if strings.Contains(strings.ToLower(flag.Description), ghost) {
					t.Fatalf("--to documents the %q alias, which no code implements", ghost)
				}
			}
		case "effective":
			for _, member := range []string{"endOfCurrentCycle", "startImmediately"} {
				if !strings.Contains(flag.Description, member) {
					t.Fatalf("--effective must document the %q enum member, got %q", member, flag.Description)
				}
			}
			if flag.Default != "endOfCurrentCycle" {
				t.Fatalf("--effective default = %v, want endOfCurrentCycle", flag.Default)
			}
		}
	}
}

func TestBillingGatewaySwitcherFlagContract(t *testing.T) {
	t.Parallel()
	required := map[string][]string{
		"change":  {"tenant", "to", "reason"},
		"history": {"tenant"},
		"get":     {"tenant", "audit"},
		"cancel":  {"tenant", "audit", "reason"},
	}
	for actionName, flags := range required {
		_, action := billingPlanAction(t, "admin-billing-gateway", actionName)
		byName := map[string]FlagDef{}
		for _, flag := range action.Flags {
			byName[flag.Name] = flag
		}
		for _, name := range flags {
			flag, ok := byName[name]
			if !ok {
				t.Fatalf("admin-billing-gateway %s is missing flag --%s", actionName, name)
			}
			if !flag.Required {
				t.Fatalf("admin-billing-gateway %s --%s must be required", actionName, name)
			}
		}
	}
}
