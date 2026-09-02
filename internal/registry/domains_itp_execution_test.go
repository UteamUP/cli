package registry

import (
	"reflect"
	"testing"
)

const itpExecutionAPIPath = "/api/quality/inspection-test-plan-executions"

func TestItpExecutionDomainRegistersAllRoutes(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "itp-execution", itpExecutionAPIPath, 8)
	qmsAuditAssertRoutes(t, domain, map[string]qmsAuditRouteExpectation{
		"search":          {method: "GET", path: itpExecutionAPIPath, tool: "UteamupQualityItpExecutionSearch"},
		"get":             {method: "GET", path: itpExecutionAPIPath + "/" + qmsAuditValidGUID, tool: "UteamupQualityItpExecutionGet"},
		"create":          {method: "POST", path: itpExecutionAPIPath, tool: "UteamupQualityItpExecutionCreate"},
		"transition":      {method: "POST", path: itpExecutionAPIPath + "/" + qmsAuditValidGUID + "/transitions/itp-execution.start", tool: "UteamupQualityItpExecutionTransition"},
		"result-upsert":   {method: "PUT", path: itpExecutionAPIPath + "/" + qmsAuditValidGUID + "/results", tool: "UteamupQualityItpExecutionResultUpsert"},
		"point-event-add": {method: "POST", path: itpExecutionAPIPath + "/" + qmsAuditValidGUID + "/point-events", tool: "UteamupQualityItpExecutionPointEventAdd"},
		"evidence-add":    {method: "POST", path: itpExecutionAPIPath + "/" + qmsAuditValidGUID + "/evidence", tool: "UteamupQualityItpExecutionEvidenceAdd"},
		"evidence-revoke": {method: "POST", path: itpExecutionAPIPath + "/" + qmsAuditValidGUID + "/evidence/" + qmsAuditValidGUID + "/revoke", tool: "UteamupQualityItpExecutionEvidenceRevoke"},
	})
}

func TestItpExecutionSearchMirrorsControllerQueryNames(t *testing.T) {
	t.Parallel()
	search := qmsAuditAction(t, qmsAuditDomain(t, "itp-execution", itpExecutionAPIPath, 8), "search")
	qmsAuditAssertSearchFlags(t, search, map[string]FlagDef{
		"plan-guid":                 {QueryName: "inspectionTestPlanGuid", Type: "non-empty-uuid"},
		"revision-guid":             {QueryName: "inspectionTestPlanRevisionGuid", Type: "non-empty-uuid"},
		"status":                    {QueryName: "status", Type: "string"},
		"release-state":             {QueryName: "releaseState", Type: "string"},
		"scheduled-on-or-after-utc": {QueryName: "scheduledOnOrAfterUtc", Type: "string"},
		"scheduled-before-utc":      {QueryName: "scheduledBeforeUtc", Type: "string"},
		"query":                     {QueryName: "query", Type: "string"},
		"page":                      {QueryName: "page", Type: "int", Default: 1},
		"page-size":                 {QueryName: "pageSize", Type: "int", Default: 25},
	})
}

func TestItpExecutionMutationsRequireExactSafetyFlags(t *testing.T) {
	t.Parallel()
	domain := qmsAuditDomain(t, "itp-execution", itpExecutionAPIPath, 8)
	qmsAuditAssertMutationFlags(t, domain, map[string]qmsAuditMutationExpectation{
		"create":          {},
		"transition":      {concurrency: true, confirmation: true},
		"result-upsert":   {concurrency: true},
		"point-event-add": {concurrency: true},
		"evidence-add":    {concurrency: true},
		"evidence-revoke": {concurrency: true, confirmation: true},
	})
}

func TestItpExecutionTransitionMirrorsTheClosedLifecycle(t *testing.T) {
	t.Parallel()
	transition := qmsAuditAction(t, qmsAuditDomain(t, "itp-execution", itpExecutionAPIPath, 8), "transition")
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
		"itp-execution.start", "itp-execution.reach-hold", "itp-execution.release-hold", "itp-execution.waive-hold",
		"itp-execution.complete-steps", "itp-execution.approve-release", "itp-execution.return", "itp-execution.link-ncr",
		"itp-execution.reopen", "itp-execution.cancel", "itp-execution.abort", "itp-execution.review-partial",
	}
	if !reflect.DeepEqual(actionKey.AllowedValues, want) {
		t.Fatalf("execution action keys = %v, want %v", actionKey.AllowedValues, want)
	}
}
