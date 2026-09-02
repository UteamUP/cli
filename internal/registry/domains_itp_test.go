package registry

import (
	"reflect"
	"testing"
)

const itpAPIPath = "/api/quality/inspection-test-plans"

func TestItpDomainRegistersAllRoutes(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "itp", itpAPIPath, 8)
	qmsAuditAssertRoutes(t, domain, map[string]qmsAuditRouteExpectation{
		"search":              {method: "GET", path: itpAPIPath, tool: "UteamupQualityItpSearch"},
		"get":                 {method: "GET", path: itpAPIPath + "/" + qmsAuditValidGUID, tool: "UteamupQualityItpGet"},
		"create":              {method: "POST", path: itpAPIPath, tool: "UteamupQualityItpCreate"},
		"update":              {method: "PUT", path: itpAPIPath + "/" + qmsAuditValidGUID, tool: "UteamupQualityItpUpdate"},
		"revision-create":     {method: "POST", path: itpAPIPath + "/" + qmsAuditValidGUID + "/revisions", tool: "UteamupQualityItpRevisionCreate"},
		"revision-transition": {method: "POST", path: itpAPIPath + "/" + qmsAuditValidGUID + "/revisions/" + qmsAuditValidGUID + "/transitions/itp-revision.submit", tool: "UteamupQualityItpRevisionTransition"},
		"evidence-add":        {method: "POST", path: itpAPIPath + "/" + qmsAuditValidGUID + "/evidence", tool: "UteamupQualityItpEvidenceAdd"},
		"evidence-revoke":     {method: "POST", path: itpAPIPath + "/" + qmsAuditValidGUID + "/evidence/" + qmsAuditValidGUID + "/revoke", tool: "UteamupQualityItpEvidenceRevoke"},
	})
}

func TestItpSearchMirrorsControllerQueryNames(t *testing.T) {
	t.Parallel()
	search := qmsAuditAction(t, qmsAuditDomain(t, "itp", itpAPIPath, 8), "search")
	qmsAuditAssertSearchFlags(t, search, map[string]FlagDef{
		"project-guid":    {QueryName: "projectGuid", Type: "non-empty-uuid"},
		"contract-guid":   {QueryName: "contractGuid", Type: "non-empty-uuid"},
		"customer-guid":   {QueryName: "customerGuid", Type: "non-empty-uuid"},
		"vendor-guid":     {QueryName: "vendorGuid", Type: "non-empty-uuid"},
		"part-guid":       {QueryName: "partGuid", Type: "non-empty-uuid"},
		"owner-user-guid": {QueryName: "ownerUserGuid", Type: "non-empty-uuid"},
		"status":          {QueryName: "status", Type: "string"},
		"query":           {QueryName: "query", Type: "string"},
		"page":            {QueryName: "page", Type: "int", Default: 1},
		"page-size":       {QueryName: "pageSize", Type: "int", Default: 25},
	})
}

func TestItpMutationsRequireExactSafetyFlags(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "itp", itpAPIPath, 8)
	qmsAuditAssertMutationFlags(t, domain, map[string]qmsAuditMutationExpectation{
		"create":              {},
		"update":              {concurrency: true},
		"revision-create":     {concurrency: true},
		"revision-transition": {concurrency: true, confirmation: true},
		"evidence-add":        {concurrency: true},
		"evidence-revoke":     {concurrency: true, confirmation: true},
	})
}

func TestItpRevisionTransitionMirrorsTheClosedLifecycle(t *testing.T) {
	t.Parallel()
	transition := qmsAuditAction(t, qmsAuditDomain(t, "itp", itpAPIPath, 8), "revision-transition")
	var actionKey *ArgDef
	for i := range transition.Args {
		if transition.Args[i].Name == "actionKey" {
			actionKey = &transition.Args[i]
		}
	}
	if actionKey == nil || !actionKey.Required {
		t.Fatalf("revision-transition must carry a required actionKey argument: %+v", transition.Args)
	}
	want := []string{"itp-revision.submit", "itp-revision.approve", "itp-revision.return", "itp-revision.release", "itp-revision.approve-successor", "itp-revision.withdraw"}
	if !reflect.DeepEqual(actionKey.AllowedValues, want) {
		t.Fatalf("revision action keys = %v, want %v", actionKey.AllowedValues, want)
	}
}
