package registry

import "testing"

func TestHostingOverviewReadsTheProtectedSnapshotRoute(t *testing.T) {
	t.Parallel()
	domain := findDomain("hosting")
	if domain == nil || len(domain.Actions) != 3 {
		t.Fatal("hosting must expose saved overview, permission repair and fresh collection")
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

func TestHostingPermissionRetryUsesTheProtectedOperatorRoute(t *testing.T) {
	t.Parallel()
	domain := findDomain("hosting")
	if domain == nil || len(domain.Actions) != 3 {
		t.Fatal("hosting operator action missing")
	}
	action := domain.Actions[1]
	path, _ := buildRESTPath(domain, action, map[string]any{})
	if path != "/api/globaladmin/hosting/retry-permissions" || action.HTTPMethod != "POST" || action.ToolName != "UteamupHostingRetryPermissions" {
		t.Fatalf("unexpected hosting recovery contract: %s %+v", path, action)
	}
	if len(action.Args) != 0 || len(action.Flags) != 0 {
		t.Fatal("permission repair must not accept arbitrary scope or force flags")
	}
}

func TestHostingRefreshUsesProtectedCollectionRoute(t *testing.T) {
	t.Parallel()
	domain := findDomain("hosting")
	if domain == nil || len(domain.Actions) != 3 {
		t.Fatal("hosting refresh action missing")
	}
	action := domain.Actions[2]
	path, _ := buildRESTPath(domain, action, map[string]any{})
	if path != "/api/globaladmin/hosting/refresh" || action.HTTPMethod != "POST" || action.ToolName != "UteamupHostingRefresh" {
		t.Fatalf("unexpected hosting refresh contract: %s %+v", path, action)
	}
	if len(action.Args) != 0 || len(action.Flags) != 1 || action.Flags[0].QueryName != "month" {
		t.Fatal("refresh may select a month but must not accept arbitrary scope or force flags")
	}
}
