package registry

import "testing"

func TestHostingOverviewReadsTheProtectedSnapshotRoute(t *testing.T) {
	t.Parallel()
	domain := findDomain("hosting")
	if domain == nil || len(domain.Actions) != 1 {
		t.Fatal("hosting must expose exactly one read-only overview")
	}
	action := domain.Actions[0]
	path, _ := buildRESTPath(domain, action, map[string]any{})
	if path != "/api/globaladmin/hosting" || action.HTTPMethod != "GET" || action.ToolName != "UteamupHostingCosts" {
		t.Fatalf("unexpected hosting contract: %s %+v", path, action)
	}
	if len(action.Args) != 0 || len(action.Flags) != 1 || action.Flags[0].QueryName != "month" {
		t.Fatalf("hosting only accepts a month filter, never a caller-selected subscription or refresh: %+v", action)
	}
}
