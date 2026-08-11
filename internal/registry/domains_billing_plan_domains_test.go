package registry

import "testing"

// billing-kling-remediation F5.3 — coverage for the billing/plan domains that shipped
// without a _test.go. Route assertions pin the URL contract for every domain whose
// registration declares real paths; admin-billing-gateway gets registration/flag-contract
// coverage only, because its routing is a known defect (no APIPath/RESTPath/MCPOnly →
// phantom derived URL) tracked for its own fix — pinning the broken path would bless it.

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
