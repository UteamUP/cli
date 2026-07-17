package registry

import "testing"

func TestAssetRentalAvailableActionHasNoIdentifierBoundary(t *testing.T) {
	domain := findDomain("asset-rental")
	if domain == nil {
		t.Fatal("asset-rental domain is not registered")
	}
	for index := range domain.Actions {
		action := &domain.Actions[index]
		if action.Name != "available" {
			continue
		}
		if action.ToolName != "UteamupAssetrentalGetAvailable" ||
			action.HTTPMethod != "GET" || action.RESTPath != "available" ||
			len(action.Args) != 0 || len(action.Flags) != 0 {
			t.Fatalf("unexpected available-rental action: %+v", action)
		}
		return
	}
	t.Fatal("asset-rental available action is not registered")
}
