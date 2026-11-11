package registry

import "testing"

func TestGlobalSearchPreservesRESTAndMcpContracts(t *testing.T) {
	action := findDomainAction(t, "search", "global")
	if action.HTTPMethod != "GET" || action.RESTPath != "global" || action.ToolName != "uteamup_global_search" {
		t.Fatalf("Unexpected global search transport: %+v", action)
	}
	if len(action.Args) != 1 || action.Args[0].QueryName != "query" || !action.Args[0].Required {
		t.Fatalf("Search text must be a query parameter: %+v", action.Args)
	}
	for name, query := range map[string]string{"max-results": "maxResults", "entity-types": "entityTypes"} {
		flag := actionFlagByName(t, action, name)
		if flag.QueryName != query {
			t.Errorf("%s binds %s, want %s", name, flag.QueryName, query)
		}
	}
}
