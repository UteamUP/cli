package registry

import (
	"reflect"
	"testing"
)

const scarAPIPath = "/api/quality/supplier-corrective-action-requests"

func TestScarDomainRegistersAllRoutes(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "scar", scarAPIPath, 17)
	record := scarAPIPath + "/" + qmsAuditValidGUID
	qmsAuditAssertRoutes(t, domain, map[string]qmsAuditRouteExpectation{
		"search":                 {method: "GET", path: scarAPIPath, tool: "UteamupQualityScarSearch"},
		"get":                    {method: "GET", path: record, tool: "UteamupQualityScarGet"},
		"create":                 {method: "POST", path: scarAPIPath, tool: "UteamupQualityScarCreate"},
		"update":                 {method: "PUT", path: record, tool: "UteamupQualityScarUpdate"},
		"transition":             {method: "POST", path: record + "/transitions/scar.submit", tool: "UteamupQualityScarTransition"},
		"evidence-add":           {method: "POST", path: record + "/evidence", tool: "UteamupQualityScarEvidenceAdd"},
		"evidence-revoke":        {method: "POST", path: record + "/evidence/" + qmsAuditValidGUID + "/revoke", tool: "UteamupQualityScarEvidenceRevoke"},
		"grant-create":           {method: "POST", path: record + "/external-grants", tool: "UteamupQualityScarGrantCreate"},
		"grant-revoke":           {method: "POST", path: record + "/external-grants/" + qmsAuditValidGUID + "/revoke", tool: "UteamupQualityScarGrantRevoke"},
		"response-review":        {method: "POST", path: record + "/response-reviews", tool: "UteamupQualityScarResponseReview"},
		"effectiveness-verify":   {method: "POST", path: record + "/effectiveness-reviews", tool: "UteamupQualityScarEffectivenessVerify"},
		"non-conformance-add":    {method: "POST", path: record + "/non-conformances/" + qmsAuditValidGUID, tool: "UteamupQualityScarNonConformanceAdd"},
		"non-conformance-revoke": {method: "POST", path: record + "/non-conformances/" + qmsAuditValidGUID + "/revoke", tool: "UteamupQualityScarLinkRevoke"},
		"capa-add":               {method: "POST", path: record + "/corrective-preventive-actions/" + qmsAuditValidGUID, tool: "UteamupQualityScarCapaAdd"},
		"capa-revoke":            {method: "POST", path: record + "/corrective-preventive-actions/" + qmsAuditValidGUID + "/revoke", tool: "UteamupQualityScarLinkRevoke"},
		"communication-add":      {method: "POST", path: record + "/communications", tool: "UteamupQualityScarCommunicationAdd"},
		"communication-revoke":   {method: "POST", path: record + "/communications/" + qmsAuditValidGUID + "/revoke", tool: "UteamupQualityScarLinkRevoke"},
	})
}

func TestScarSearchMirrorsControllerQueryNames(t *testing.T) {
	t.Parallel()
	search := qmsAuditAction(t, qmsAuditDomain(t, "scar", scarAPIPath, 17), "search")
	qmsAuditAssertSearchFlags(t, search, map[string]FlagDef{
		"site-location-guid":  {QueryName: "siteLocationGuid", Type: "non-empty-uuid"},
		"vendor-guid":         {QueryName: "vendorGuid", Type: "non-empty-uuid"},
		"owner-user-guid":     {QueryName: "ownerUserGuid", Type: "non-empty-uuid"},
		"status":              {QueryName: "status", Type: "string"},
		"severity":            {QueryName: "severity", Type: "string"},
		"due-on-or-after-utc": {QueryName: "dueOnOrAfterUtc", Type: "string"},
		"due-before-utc":      {QueryName: "dueBeforeUtc", Type: "string"},
		"query":               {QueryName: "query", Type: "string"},
		"page":                {QueryName: "page", Type: "int", Default: 1},
		"page-size":           {QueryName: "pageSize", Type: "int", Default: 25},
	})
}

func TestScarMutationsRequireExactSafetyFlags(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "scar", scarAPIPath, 17)
	qmsAuditAssertMutationFlags(t, domain, map[string]qmsAuditMutationExpectation{
		"create":                 {},
		"update":                 {concurrency: true},
		"transition":             {concurrency: true, confirmation: true},
		"evidence-add":           {concurrency: true},
		"evidence-revoke":        {concurrency: true, confirmation: true},
		"grant-create":           {concurrency: true},
		"grant-revoke":           {concurrency: true, confirmation: true},
		"response-review":        {concurrency: true, confirmation: true},
		"effectiveness-verify":   {concurrency: true, confirmation: true},
		"non-conformance-add":    {concurrency: true},
		"non-conformance-revoke": {concurrency: true, confirmation: true},
		"capa-add":               {concurrency: true},
		"capa-revoke":            {concurrency: true, confirmation: true},
		"communication-add":      {concurrency: true},
		"communication-revoke":   {concurrency: true, confirmation: true},
	})
}

// Supplier-only keys never reach a tenant session: the CLI refuses them before any request.
func TestScarTransitionOffersOnlyInternalLifecycleKeys(t *testing.T) {
	t.Parallel()
	transition := qmsAuditAction(t, qmsAuditDomain(t, "scar", scarAPIPath, 17), "transition")
	var actionKey *ArgDef
	for i := range transition.Args {
		if transition.Args[i].Name == "actionKey" {
			actionKey = &transition.Args[i]
		}
	}
	if actionKey == nil || !actionKey.Required {
		t.Fatalf("transition must carry a required actionKey argument: %+v", transition.Args)
	}
	want := []string{
		"scar.submit", "scar.approve-issue", "scar.return", "scar.cancel", "scar.withdraw", "scar.escalate",
		"scar.begin-internal-review", "scar.complete-implementation", "scar.reopen", "scar.resume-supplier-correction",
	}
	if !reflect.DeepEqual(actionKey.AllowedValues, want) {
		t.Fatalf("internal action keys = %v, want %v", actionKey.AllowedValues, want)
	}
	for _, supplierOnly := range []string{"scar.supplier-acknowledge", "scar.supplier-decline", "scar.submit-containment", "scar.submit-response", "scar.resubmit-response", "scar.accept-response", "scar.approve-effective"} {
		for _, allowed := range actionKey.AllowedValues {
			if allowed == supplierOnly {
				t.Errorf("%s must not be offered on the internal transition route", supplierOnly)
			}
		}
	}
}
