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
