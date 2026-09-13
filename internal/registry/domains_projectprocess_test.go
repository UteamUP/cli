package registry

import "testing"

func TestProjectProcessRoutesKeepExactRevisionsAndCompleteReviewedPayloads(t *testing.T) {
	for _, scenario := range []struct{ action, method, path, tool string }{
		{"defaults", "GET", "{projectGuid}/process/defaults", "UteamupProjectProcessDefaults"},
		{"get", "GET", "{projectGuid}/process", "UteamupProjectProcessGet"},
		{"revision", "GET", "{projectGuid}/process/revisions/{revisionGuid}", "UteamupProjectProcessRevision"},
		{"history", "GET", "{projectGuid}/process/history", "UteamupProjectProcessHistory"},
		{"preview", "POST", "{projectGuid}/process/preview", "UteamupProjectProcessPreview"},
		{"adopt", "POST", "{projectGuid}/process/adopt", "UteamupProjectProcessAdopt"},
	} {
		action := findDomainAction(t, "project-process", scenario.action)
		if action.ToolName != scenario.tool || action.RESTPath != scenario.path || action.HTTPMethod != scenario.method {
			t.Fatalf("wrong process route for %s: %+v", scenario.action, action)
		}
		if scenario.method == "POST" && (len(action.Flags) != 1 || action.Flags[0].Name != "from-json") {
			t.Fatalf("%s must preserve the full definition, request identity and review fingerprint", scenario.action)
		}
	}
}
