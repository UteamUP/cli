package registry

import (
	"reflect"
	"testing"
)

func TestAuditProgramDomainRegistersAllRoutes(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "audit-program", "/api/quality/audit-programs", 7)
	qmsAuditAssertRoutes(t, domain, map[string]qmsAuditRouteExpectation{
		"search": {
			method: "GET",
			path:   "/api/quality/audit-programs",
			tool:   "UteamupQualityAuditProgramSearch",
		},
		"get": {
			method: "GET",
			path:   "/api/quality/audit-programs/" + qmsAuditValidGUID,
			tool:   "UteamupQualityAuditProgramGet",
		},
		"create": {
			method: "POST",
			path:   "/api/quality/audit-programs",
			tool:   "UteamupQualityAuditProgramCreate",
		},
		"update": {
			method: "PUT",
			path:   "/api/quality/audit-programs/" + qmsAuditValidGUID,
			tool:   "UteamupQualityAuditProgramUpdate",
		},
		"transition": {
			method: "POST",
			path: "/api/quality/audit-programs/" + qmsAuditValidGUID +
				"/transitions/quality-audit-program.submit",
			tool: "UteamupQualityAuditProgramTransition",
		},
		"evidence-add": {
			method: "POST",
			path:   "/api/quality/audit-programs/" + qmsAuditValidGUID + "/evidence",
			tool:   "UteamupQualityAuditProgramEvidenceAdd",
		},
		"evidence-revoke": {
			method: "POST",
			path: "/api/quality/audit-programs/" + qmsAuditValidGUID +
				"/evidence/" + qmsAuditValidGUID + "/revoke",
			tool: "UteamupQualityAuditProgramEvidenceRevoke",
		},
	})
}

func TestAuditProgramSearchMirrorsControllerQueryNames(t *testing.T) {
	t.Parallel()
	search := qmsAuditAction(
		t,
		qmsAuditDomain(t, "audit-program", "/api/quality/audit-programs", 7),
		"search",
	)
	qmsAuditAssertSearchFlags(t, search, map[string]FlagDef{
		"project-guid":                  {QueryName: "projectGuid", Type: "non-empty-uuid"},
		"owner-user-guid":               {QueryName: "ownerUserGuid", Type: "non-empty-uuid"},
		"status":                        {QueryName: "status", Type: "string"},
		"period-starts-on-or-after-utc": {QueryName: "periodStartsOnOrAfterUtc", Type: "string"},
		"period-ends-on-or-before-utc":  {QueryName: "periodEndsOnOrBeforeUtc", Type: "string"},
		"query":                         {QueryName: "query", Type: "string"},
		"page":                          {QueryName: "page", Type: "int", Default: 1},
		"page-size":                     {QueryName: "pageSize", Type: "int", Default: 25},
	})
}

func TestAuditProgramMutationsRequireExactSafetyFlags(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "audit-program", "/api/quality/audit-programs", 7)
	qmsAuditAssertMutationFlags(t, domain, map[string]qmsAuditMutationExpectation{
		"create":          {},
		"update":          {concurrency: true},
		"transition":      {concurrency: true, confirmation: true},
		"evidence-add":    {concurrency: true},
		"evidence-revoke": {concurrency: true, confirmation: true},
	})
}

func TestAuditProgramTransitionActionKeysAreClosed(t *testing.T) {
	t.Parallel()
	want := []string{
		"quality-audit-program.submit",
		"quality-audit-program.return",
		"quality-audit-program.approve",
		"quality-audit-program.activate",
		"quality-audit-program.cancel",
		"quality-audit-program.complete",
	}
	transition := qmsAuditAction(
		t,
		qmsAuditDomain(t, "audit-program", "/api/quality/audit-programs", 7),
		"transition",
	)
	if len(transition.Args) != 2 || !reflect.DeepEqual(transition.Args[1].AllowedValues, want) {
		t.Fatalf("audit-program transition keys = %v, want %v", transition.Args, want)
	}
}
