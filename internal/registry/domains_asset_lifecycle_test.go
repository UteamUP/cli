package registry

import "testing"

func TestAssetLifecycleFilteredReadsExposeNoIdentifierArguments(t *testing.T) {
	domain := findDomain("asset-lifecycle")
	if domain == nil {
		t.Fatal("asset-lifecycle domain is not registered")
	}
	actions := map[string]*Action{}
	for index := range domain.Actions {
		actions[domain.Actions[index].Name] = &domain.Actions[index]
	}
	byType := actions["by-type"]
	if byType == nil || byType.ToolName != "UteamupAssetlifecycleGetByType" ||
		len(byType.Args) != 0 || len(byType.Flags) != 1 ||
		byType.Flags[0].Name != "event-type" {
		t.Fatalf("unexpected lifecycle by-type action: %+v", byType)
	}
	byDate := actions["by-date"]
	if byDate == nil || byDate.ToolName != "UteamupAssetlifecycleGetByDateRange" ||
		len(byDate.Args) != 0 || len(byDate.Flags) != 2 {
		t.Fatalf("unexpected lifecycle by-date action: %+v", byDate)
	}
}

func TestAssetLifecycleWorkspaceAndSetupUseGuidFirstRoutes(t *testing.T) {
	domain := findDomain("asset-lifecycle")
	if domain == nil {
		t.Fatal("asset-lifecycle domain is not registered")
	}
	actions := map[string]*Action{}
	for index := range domain.Actions {
		actions[domain.Actions[index].Name] = &domain.Actions[index]
	}

	workspace := actions["workspace"]
	if workspace == nil || workspace.RESTBasePath != "/api/asset-lifecycle-workspace" ||
		workspace.HTTPMethod != "GET" || !workspace.UseDomainBasePath {
		t.Fatalf("unexpected lifecycle workspace action: %+v", workspace)
	}

	for _, name := range []string{"setup-get", "setup-apply"} {
		action := actions[name]
		if action == nil || action.RESTBasePath != "/api/assetlifecyclesetup" ||
			action.RESTPath != "asset/{assetGuid}" || len(action.Args) != 1 ||
			action.Args[0].Name != "assetGuid" || action.Args[0].Type != "uuid" {
			t.Fatalf("unexpected %s action: %+v", name, action)
		}
	}

	for _, name := range []string{"setup-apply", "setup-bulk-apply"} {
		action := actions[name]
		if action == nil || len(action.Flags) != 1 || !action.Flags[0].RootJSONObjectFile {
			t.Fatalf("%s must accept one root JSON request file: %+v", name, action)
		}
	}
}
