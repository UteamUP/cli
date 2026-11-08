package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHostingOverviewReadsTheProtectedSnapshotRoute(t *testing.T) {
	t.Parallel()
	domain := findDomain("hosting")
	if domain == nil || len(domain.Actions) != 5 {
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
	if domain == nil || len(domain.Actions) != 5 {
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
	if domain == nil || len(domain.Actions) != 5 {
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

func TestHostingBudgetUsesProtectedReadAndCompleteReviewedPatch(t *testing.T) {
	domain := findDomain("hosting")
	for _, name := range []string{"budget-alert", "budget-alert-update"} {
		action := findAction(domain, name)
		path, _ := buildRESTPath(domain, *action, map[string]any{})
		if path != "/api/globaladmin/hosting/budget-alert" || len(action.Args) != 0 || action.MCPOnly {
			t.Fatalf("invalid budget route: %s %+v", path, action)
		}
	}
	action := findAction(domain, "budget-alert-update")
	if action.HTTPMethod != "PATCH" || action.ToolName != "UteamupHostingBudgetAlertUpdate" || len(action.Flags) != 1 || !action.Flags[0].Required || !action.Flags[0].RootJSONObjectFile {
		t.Fatalf("budget mutation must send reviewed complete settings: %+v", action)
	}
	input := filepath.Join(t.TempDir(), "budget.json")
	if err := os.WriteFile(input, []byte(`{"enabled":true,"monthlyLimitEur":300,"recipientEmails":["one@example.test","two@example.test"],"language":"is","monthLimits":{"2026-11-01":500}}`), 0600); err != nil {
		t.Fatal(err)
	}
	requests, err := runRegisteredDomainCommand(t, "hosting", "budget-alert-update", "--from-json", input)
	if err != nil || len(requests) != 1 {
		t.Fatalf("budget patch failed: %v %+v", err, requests)
	}
	body := requests[0].body
	if body["enabled"] != true || body["monthlyLimitEur"] != float64(300) || len(body["recipientEmails"].([]any)) != 2 || body["monthLimits"].(map[string]any)["2026-11-01"] != float64(500) {
		t.Fatalf("reviewed settings changed: %+v", body)
	}
	if _, wrapped := body["settings"]; wrapped {
		t.Fatal("REST body must not be nested")
	}
}
