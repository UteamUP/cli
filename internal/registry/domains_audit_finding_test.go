package registry

import (
	"reflect"
	"testing"
)

func TestAuditFindingDomainRegistersAllRoutes(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "audit-finding", "/api/quality/audit-findings", 11)
	base := "/api/quality/audit-findings/" + qmsAuditValidGUID
	qmsAuditAssertRoutes(t, domain, map[string]qmsAuditRouteExpectation{
		"search": {
			method: "GET",
			path:   "/api/quality/audit-findings",
			tool:   "UteamupQualityAuditFindingSearch",
		},
		"get": {
			method: "GET",
			path:   base,
			tool:   "UteamupQualityAuditFindingGet",
		},
		"create": {
			method: "POST",
			path:   "/api/quality/audit-findings",
			tool:   "UteamupQualityAuditFindingCreate",
		},
		"update": {
			method: "PUT",
			path:   base,
			tool:   "UteamupQualityAuditFindingUpdate",
		},
		"transition": {
			method: "POST",
			path:   base + "/transitions/quality-audit-finding.issue",
			tool:   "UteamupQualityAuditFindingTransition",
		},
		"evidence-add": {
			method: "POST",
			path:   base + "/evidence",
			tool:   "UteamupQualityAuditFindingEvidenceAdd",
		},
		"evidence-revoke": {
			method: "POST",
			path:   base + "/evidence/" + qmsAuditValidGUID + "/revoke",
			tool:   "UteamupQualityAuditFindingEvidenceRevoke",
		},
		"ncr-link-add": {
			method: "POST",
			path:   base + "/non-conformances/" + qmsAuditValidGUID,
			tool:   "UteamupQualityAuditFindingNonConformanceAdd",
		},
		"ncr-link-revoke": {
			method: "POST",
			path:   base + "/non-conformances/" + qmsAuditValidGUID + "/revoke",
			tool:   "UteamupQualityAuditFindingNonConformanceRevoke",
		},
		"capa-link-add": {
			method: "POST",
			path:   base + "/corrective-preventive-actions/" + qmsAuditValidGUID,
			tool:   "UteamupQualityAuditFindingCapaAdd",
		},
		"capa-link-revoke": {
			method: "POST",
			path:   base + "/corrective-preventive-actions/" + qmsAuditValidGUID + "/revoke",
			tool:   "UteamupQualityAuditFindingCapaRevoke",
		},
	})
}

func TestAuditFindingSearchMirrorsControllerQueryNames(t *testing.T) {
	t.Parallel()
	search := qmsAuditAction(
		t,
		qmsAuditDomain(t, "audit-finding", "/api/quality/audit-findings", 11),
		"search",
	)
	qmsAuditAssertSearchFlags(t, search, map[string]FlagDef{
		"site-location-guid":                {QueryName: "siteLocationGuid", Type: "non-empty-uuid"},
		"quality-audit-guid":                {QueryName: "qualityAuditGuid", Type: "non-empty-uuid"},
		"quality-audit-checklist-item-guid": {QueryName: "qualityAuditChecklistItemGuid", Type: "non-empty-uuid"},
		"owner-user-guid":                   {QueryName: "ownerUserGuid", Type: "non-empty-uuid"},
		"classification":                    {QueryName: "classification", Type: "string"},
		"status":                            {QueryName: "status", Type: "string"},
		"due-on-or-after-utc":               {QueryName: "dueOnOrAfterUtc", Type: "string"},
		"due-before-utc":                    {QueryName: "dueBeforeUtc", Type: "string"},
		"query":                             {QueryName: "query", Type: "string"},
		"page":                              {QueryName: "page", Type: "int", Default: 1},
		"page-size":                         {QueryName: "pageSize", Type: "int", Default: 25},
	})
}

func TestAuditFindingMutationsRequireExactSafetyFlags(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "audit-finding", "/api/quality/audit-findings", 11)
	qmsAuditAssertMutationFlags(t, domain, map[string]qmsAuditMutationExpectation{
		"create":           {},
		"update":           {concurrency: true},
		"transition":       {concurrency: true, confirmation: true},
		"evidence-add":     {concurrency: true},
		"evidence-revoke":  {concurrency: true, confirmation: true},
		"ncr-link-add":     {concurrency: true},
		"ncr-link-revoke":  {concurrency: true, confirmation: true},
		"capa-link-add":    {concurrency: true},
		"capa-link-revoke": {concurrency: true, confirmation: true},
	})
}

func TestAuditFindingTransitionActionKeysAreClosed(t *testing.T) {
	t.Parallel()
	want := []string{
		"quality-audit-finding.issue",
		"quality-audit-finding.request-response",
		"quality-audit-finding.cancel",
		"quality-audit-finding.submit-response",
		"quality-audit-finding.close-without-response",
		"quality-audit-finding.verify-close",
		"quality-audit-finding.return-response",
		"quality-audit-finding.reopen",
		"quality-audit-finding.resume-response",
	}
	transition := qmsAuditAction(
		t,
		qmsAuditDomain(t, "audit-finding", "/api/quality/audit-findings", 11),
		"transition",
	)
	if len(transition.Args) != 2 || !reflect.DeepEqual(transition.Args[1].AllowedValues, want) {
		t.Fatalf("audit-finding transition keys = %v, want %v", transition.Args, want)
	}
}
