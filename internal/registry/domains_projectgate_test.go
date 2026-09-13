package registry

import "testing"

func TestProjectGatePreservesReviewedPayloadsAndExactScopeRoutes(t *testing.T) {
	for _, scenario := range []struct{ action, method, suffix, tool string }{
		{"preview", "POST", "/submissions/preview", "UteamupProjectGateSubmissionPreview"},
		{"submit", "POST", "/submissions", "UteamupProjectGateSubmissionSubmit"},
		{"get", "GET", "/submissions/{submissionGuid}", "UteamupProjectGateSubmissionGet"},
		{"history", "GET", "/submissions", "UteamupProjectGateSubmissionList"},
		{"decisions", "GET", "", "UteamupProjectGateDecisionList"},
		{"decide", "POST", "", "UteamupProjectGateDecisionRecord"},
	} {
		action := findDomainAction(t, "project-gate", scenario.action)
		if action.ToolName != scenario.tool || action.HTTPMethod != scenario.method || action.RESTPath != "{projectGuid}/stages/{stageGuid}/decisions"+scenario.suffix {
			t.Fatalf("wrong gate route for %s: %+v", scenario.action, action)
		}
		if len(action.Args) < 2 || action.Args[0].Name != "projectGuid" || action.Args[1].Name != "stageGuid" {
			t.Fatalf("%s lost exact project/stage ownership", scenario.action)
		}
		if scenario.method == "POST" && (len(action.Flags) != 1 || action.Flags[0].Name != "from-json") {
			t.Fatalf("%s must preserve complete reviewed terms, source identities and retry identity", scenario.action)
		}
	}
}
