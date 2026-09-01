package registry

import (
	"reflect"
	"testing"
)

func TestAuditDomainRegistersAllRoutes(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "audit", "/api/quality/audits", 12)
	base := "/api/quality/audits/" + qmsAuditValidGUID
	qmsAuditAssertRoutes(t, domain, map[string]qmsAuditRouteExpectation{
		"search": {
			method: "GET",
			path:   "/api/quality/audits",
			tool:   "UteamupQualityAuditSearch",
		},
		"get": {
			method: "GET",
			path:   base,
			tool:   "UteamupQualityAuditGet",
		},
		"create": {
			method: "POST",
			path:   "/api/quality/audits",
			tool:   "UteamupQualityAuditCreate",
		},
		"update": {
			method: "PUT",
			path:   base,
			tool:   "UteamupQualityAuditUpdate",
		},
		"transition": {
			method: "POST",
			path:   base + "/transitions/quality-audit.approve-plan",
			tool:   "UteamupQualityAuditTransition",
		},
		"assignment-verify-competence": {
			method: "POST",
			path:   base + "/assignments/" + qmsAuditValidGUID + "/verify-competence",
			tool:   "UteamupQualityAuditAssignmentCompetenceVerify",
		},
		"assignment-review-independence": {
			method: "POST",
			path:   base + "/assignments/" + qmsAuditValidGUID + "/review-independence",
			tool:   "UteamupQualityAuditAssignmentIndependenceReview",
		},
		"assignment-accept": {
			method: "POST",
			path:   base + "/assignments/" + qmsAuditValidGUID + "/accept",
			tool:   "UteamupQualityAuditAssignmentAccept",
		},
		"assignment-remove": {
			method: "POST",
			path:   base + "/assignments/" + qmsAuditValidGUID + "/remove",
			tool:   "UteamupQualityAuditAssignmentRemove",
		},
		"checklist-evaluate": {
			method: "POST",
			path:   base + "/checklist-items/" + qmsAuditValidGUID + "/evaluate",
			tool:   "UteamupQualityAuditChecklistEvaluate",
		},
		"evidence-add": {
			method: "POST",
			path:   base + "/evidence",
			tool:   "UteamupQualityAuditEvidenceAdd",
		},
		"evidence-revoke": {
			method: "POST",
			path:   base + "/evidence/" + qmsAuditValidGUID + "/revoke",
			tool:   "UteamupQualityAuditEvidenceRevoke",
		},
	})
}

func TestAuditSearchMirrorsControllerQueryNames(t *testing.T) {
	t.Parallel()
	search := qmsAuditAction(t, qmsAuditDomain(t, "audit", "/api/quality/audits", 12), "search")
	qmsAuditAssertSearchFlags(t, search, map[string]FlagDef{
		"site-location-guid":         {QueryName: "siteLocationGuid", Type: "non-empty-uuid"},
		"quality-audit-program-guid": {QueryName: "qualityAuditProgramGuid", Type: "non-empty-uuid"},
		"vendor-guid":                {QueryName: "vendorGuid", Type: "non-empty-uuid"},
		"lead-auditor-user-guid":     {QueryName: "leadAuditorUserGuid", Type: "non-empty-uuid"},
		"status":                     {QueryName: "status", Type: "string"},
		"type":                       {QueryName: "type", Type: "string"},
		"scheduled-on-or-after-utc":  {QueryName: "scheduledOnOrAfterUtc", Type: "string"},
		"scheduled-before-utc":       {QueryName: "scheduledBeforeUtc", Type: "string"},
		"query":                      {QueryName: "query", Type: "string"},
		"page":                       {QueryName: "page", Type: "int", Default: 1},
		"page-size":                  {QueryName: "pageSize", Type: "int", Default: 25},
	})
}

func TestAuditMutationsRequireExactSafetyFlags(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "audit", "/api/quality/audits", 12)
	qmsAuditAssertMutationFlags(t, domain, map[string]qmsAuditMutationExpectation{
		"create":                         {},
		"update":                         {concurrency: true},
		"transition":                     {concurrency: true, confirmation: true},
		"assignment-verify-competence":   {concurrency: true, confirmation: true},
		"assignment-review-independence": {concurrency: true, confirmation: true},
		"assignment-accept":              {concurrency: true, confirmation: true},
		"assignment-remove":              {concurrency: true, confirmation: true},
		"checklist-evaluate":             {concurrency: true, confirmation: true},
		"evidence-add":                   {concurrency: true},
		"evidence-revoke":                {concurrency: true, confirmation: true},
	})
}

func TestAuditTransitionActionKeysAreClosed(t *testing.T) {
	t.Parallel()
	want := []string{
		"quality-audit.approve-plan",
		"quality-audit.schedule",
		"quality-audit.accept-assignments",
		"quality-audit.cancel",
		"quality-audit.open",
		"quality-audit.close-fieldwork",
		"quality-audit.abort",
		"quality-audit.document-partial-evidence",
		"quality-audit.issue-report",
		"quality-audit.issue-report-no-follow-up",
		"quality-audit.verify-follow-up",
		"quality-audit.reopen",
		"quality-audit.resume-follow-up",
	}
	transition := qmsAuditAction(
		t,
		qmsAuditDomain(t, "audit", "/api/quality/audits", 12),
		"transition",
	)
	if len(transition.Args) != 2 || !reflect.DeepEqual(transition.Args[1].AllowedValues, want) {
		t.Fatalf("audit transition keys = %v, want %v", transition.Args, want)
	}
}
