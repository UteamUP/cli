package registry

import (
	"testing"
)

// --- Part 2 plan-management domains: plan-limit, pricing-rule,
// tenant-feature-override, plan-impact, plan-migration, plan-approval,
// plan-analytics, subscription-lifecycle ---
//
// Each domain gets a registration assertion (explicit APIPath — every one of
// these controllers routes under its own /api/<controller> base, so
// buildRESTPath can only reach it via APIPath) plus a representative action's
// route wiring (method + RESTPath template + positional arg names, which must
// literally match the `{...}` placeholders or expandPathTemplate leaves the
// raw token in the URL).

func assertDomainAPIPath(t *testing.T, name, apiPath string) *Domain {
	t.Helper()
	d := findDomain(name)
	if d == nil {
		t.Fatalf("expected %s domain to be registered", name)
	}
	if d.APIPath != apiPath {
		t.Errorf("%s APIPath = %q, want %s", name, d.APIPath, apiPath)
	}
	return d
}

func assertSingleGuidArg(t *testing.T, action *Action, argName string) {
	t.Helper()
	if len(action.Args) != 1 {
		t.Fatalf("%s expected 1 positional arg, got %+v", action.Name, action.Args)
	}
	if action.Args[0].Name != argName || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Errorf("%s arg must be a required string %q, got %+v", action.Name, argName, action.Args[0])
	}
}

func TestPlanLimitDomainRegistered(t *testing.T) {
	assertDomainAPIPath(t, "plan-limit", "/api/planlimit")

	// upsert is the representative action: PUT by-plan/{planGuid} with a
	// dimension int flag + optional max-value (omitted = unlimited).
	a := findDomainAction(t, "plan-limit", "upsert")
	if a.HTTPMethod != "PUT" || a.RESTPath != "by-plan/{planGuid}" {
		t.Errorf("plan-limit upsert: want method=PUT path=by-plan/{planGuid}, got method=%q path=%s", a.HTTPMethod, a.RESTPath)
	}
	assertSingleGuidArg(t, a, "planGuid")
	dim := findFlag(a, "dimension")
	if dim == nil || !dim.Required || dim.Type != "int" {
		t.Errorf("plan-limit upsert --dimension must be a required int flag, got %+v", dim)
	}
	max := findFlag(a, "max-value")
	// max-value must have NO default — an unset optional flag with a default
	// would still be sent in the body, silently capping an unlimited dimension.
	if max == nil || max.Required || max.Type != "int" || max.Default != nil {
		t.Errorf("plan-limit upsert --max-value must be an optional int flag without a default, got %+v", max)
	}
}

func TestPricingRuleDomainRegistered(t *testing.T) {
	assertDomainAPIPath(t, "pricing-rule", "/api/pricingrule")

	// create is the representative action: POST with the closed condition
	// vocabulary as flags. discount-percent is a float flag (decimal DTO field).
	a := findDomainAction(t, "pricing-rule", "create")
	if a.HTTPMethod != "" || a.RESTPath != "" {
		t.Errorf("pricing-rule create: method/path derive from the action name (POST /api/pricingrule), got method=%q path=%s", a.HTTPMethod, a.RESTPath)
	}
	discount := findFlag(a, "discount-percent")
	if discount == nil || !discount.Required || discount.Type != "float" {
		t.Errorf("pricing-rule create --discount-percent must be a required float flag, got %+v", discount)
	}
	name := findFlag(a, "name")
	if name == nil || !name.Required || name.Type != "string" {
		t.Errorf("pricing-rule create --name must be a required string flag, got %+v", name)
	}

	// get/update/delete all address the rule via by-guid/{guid}.
	for _, actionName := range []string{"get", "update", "delete"} {
		byGuid := findDomainAction(t, "pricing-rule", actionName)
		if byGuid.RESTPath != "by-guid/{guid}" {
			t.Errorf("pricing-rule %s RESTPath = %q, want by-guid/{guid}", actionName, byGuid.RESTPath)
		}
		assertSingleGuidArg(t, byGuid, "guid")
	}
}

func TestTenantFeatureOverrideDomainRegistered(t *testing.T) {
	assertDomainAPIPath(t, "tenant-feature-override", "/api/tenantfeatureoverride")

	// upsert is the representative action: PUT by-tenant/{tenantGuid} with the
	// module GUID + Grant/Revoke mode in the body.
	a := findDomainAction(t, "tenant-feature-override", "upsert")
	if a.HTTPMethod != "PUT" || a.RESTPath != "by-tenant/{tenantGuid}" {
		t.Errorf("tenant-feature-override upsert: want method=PUT path=by-tenant/{tenantGuid}, got method=%q path=%s", a.HTTPMethod, a.RESTPath)
	}
	assertSingleGuidArg(t, a, "tenantGuid")
	mode := findFlag(a, "mode")
	if mode == nil || !mode.Required || mode.Type != "int" {
		t.Errorf("tenant-feature-override upsert --mode must be a required int flag, got %+v", mode)
	}

	// delete carries BOTH guids in the path.
	del := findDomainAction(t, "tenant-feature-override", "delete")
	if del.RESTPath != "by-tenant/{tenantGuid}/{featureCatalogGuid}" {
		t.Errorf("tenant-feature-override delete RESTPath = %q, want by-tenant/{tenantGuid}/{featureCatalogGuid}", del.RESTPath)
	}
	if len(del.Args) != 2 || del.Args[0].Name != "tenantGuid" || del.Args[1].Name != "featureCatalogGuid" {
		t.Errorf("tenant-feature-override delete args must be [tenantGuid, featureCatalogGuid], got %+v", del.Args)
	}
}

func TestPlanImpactDomainRegistered(t *testing.T) {
	assertDomainAPIPath(t, "plan-impact", "/api/planimpact")

	a := findDomainAction(t, "plan-impact", "preview")
	if a.HTTPMethod != "POST" || a.RESTPath != "by-plan/{planGuid}/preview" {
		t.Errorf("plan-impact preview: want method=POST path=by-plan/{planGuid}/preview, got method=%q path=%s", a.HTTPMethod, a.RESTPath)
	}
	assertSingleGuidArg(t, a, "planGuid")
	// Both proposed prices are decimal DTO fields → required float flags.
	for _, flagName := range []string{"proposed-price-per-license-isk", "proposed-price-per-helpdesk-license-isk"} {
		f := findFlag(a, flagName)
		if f == nil || !f.Required || f.Type != "float" {
			t.Errorf("plan-impact preview --%s must be a required float flag, got %+v", flagName, f)
		}
	}
}

func TestPlanMigrationDomainRegistered(t *testing.T) {
	assertDomainAPIPath(t, "plan-migration", "/api/planmigration")

	a := findDomainAction(t, "plan-migration", "migrate")
	if a.HTTPMethod != "POST" || a.RESTPath != "" {
		t.Errorf("plan-migration migrate: want method=POST at the domain base path, got method=%q path=%s", a.HTTPMethod, a.RESTPath)
	}
	// dry-run MUST default to true — the safe default for a bulk mutation —
	// and be sent explicitly so the CLI never depends on the backend default.
	dryRun := findFlag(a, "dry-run")
	if dryRun == nil || dryRun.Type != "bool" || dryRun.Default != true {
		t.Errorf("plan-migration migrate --dry-run must be a bool flag defaulting to true, got %+v", dryRun)
	}
	for _, flagName := range []string{"from-plan-guid", "to-plan-guid"} {
		f := findFlag(a, flagName)
		if f == nil || !f.Required || f.Type != "string" {
			t.Errorf("plan-migration migrate --%s must be a required string flag, got %+v", flagName, f)
		}
	}
}

func TestPlanApprovalDomainRegistered(t *testing.T) {
	assertDomainAPIPath(t, "plan-approval", "/api/planapproval")

	// pending: method "" derives GET (not in the HTTPMethod map, no update- prefix).
	pending := findDomainAction(t, "plan-approval", "pending")
	if pending.HTTPMethod != "" || pending.RESTPath != "pending" || len(pending.Args) != 0 {
		t.Errorf("plan-approval pending: want derived-GET static path `pending` with no args, got method=%q path=%s args=%+v", pending.HTTPMethod, pending.RESTPath, pending.Args)
	}

	for actionName, restPath := range map[string]string{
		"approve": "by-guid/{requestGuid}/approve",
		"reject":  "by-guid/{requestGuid}/reject",
	} {
		a := findDomainAction(t, "plan-approval", actionName)
		if a.HTTPMethod != "POST" || a.RESTPath != restPath {
			t.Errorf("plan-approval %s: want method=POST path=%s, got method=%q path=%s", actionName, restPath, a.HTTPMethod, a.RESTPath)
		}
		assertSingleGuidArg(t, a, "requestGuid")
	}
}

func TestPlanAnalyticsDomainRegistered(t *testing.T) {
	assertDomainAPIPath(t, "plan-analytics", "/api/plananalytics")

	for actionName, wantProviderFlag := range map[string]bool{"summary": false, "insights": true} {
		a := findDomainAction(t, "plan-analytics", actionName)
		// Method "" derives GET; flags travel on the query string for GET calls.
		if a.HTTPMethod != "" || a.RESTPath != actionName {
			t.Errorf("plan-analytics %s: want derived-GET static path %q, got method=%q path=%s", actionName, actionName, a.HTTPMethod, a.RESTPath)
		}
		// from-date must have NO default so an unset flag is omitted and the
		// backend's 90-days-back default applies.
		fromDate := findFlag(a, "from-date")
		if fromDate == nil || fromDate.Required || fromDate.Default != nil {
			t.Errorf("plan-analytics %s --from-date must be an optional string flag without a default, got %+v", actionName, fromDate)
		}
		if got := findFlag(a, "provider") != nil; got != wantProviderFlag {
			t.Errorf("plan-analytics %s: provider flag present = %v, want %v", actionName, got, wantProviderFlag)
		}
	}
}

func TestSubscriptionLifecycleDomainRegistered(t *testing.T) {
	d := assertDomainAPIPath(t, "subscription-lifecycle", "/api/internalbilling")
	if len(d.Actions) != 5 {
		t.Errorf("subscription-lifecycle expected 5 lifecycle actions, got %d", len(d.Actions))
	}

	for actionName, restPath := range map[string]string{
		"suspend":                "admin/subscriptions/{guid}/suspend",
		"cancel":                 "admin/subscriptions/{guid}/cancel",
		"reactivate":             "admin/subscriptions/{guid}/reactivate",
		"schedule-cancel":        "admin/subscriptions/{guid}/schedule-cancel",
		"clear-scheduled-cancel": "admin/subscriptions/{guid}/clear-scheduled-cancel",
	} {
		a := findDomainAction(t, "subscription-lifecycle", actionName)
		// Every lifecycle verb needs an explicit POST — none of these names is
		// in the HTTPMethod map, and schedule-cancel/clear-scheduled-cancel
		// would otherwise derive GET.
		if a.HTTPMethod != "POST" || a.RESTPath != restPath {
			t.Errorf("subscription-lifecycle %s: want method=POST path=%s, got method=%q path=%s", actionName, restPath, a.HTTPMethod, a.RESTPath)
		}
		assertSingleGuidArg(t, a, "guid")
	}

	// cancel deliberately has NO --reason flag: the backend binds
	// `[FromBody] string? reason` (a raw JSON string), which the flag→object
	// body mapping cannot express — a reason flag would produce {"reason":...}
	// and fail model binding with a 400.
	cancel := findDomainAction(t, "subscription-lifecycle", "cancel")
	if len(cancel.Flags) != 0 {
		t.Errorf("subscription-lifecycle cancel must take no flags (raw-string body not expressible), got %d", len(cancel.Flags))
	}

	sched := findDomainAction(t, "subscription-lifecycle", "schedule-cancel")
	cancelAt := findFlag(sched, "cancel-at")
	if cancelAt == nil || !cancelAt.Required || cancelAt.Type != "string" {
		t.Errorf("subscription-lifecycle schedule-cancel --cancel-at must be a required string flag, got %+v", cancelAt)
	}
}
