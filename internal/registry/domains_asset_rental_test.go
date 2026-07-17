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

func TestAssetRentalActiveAndExpiringActionsUseNoIdentifierBoundary(t *testing.T) {
	domain := findDomain("asset-rental")
	if domain == nil {
		t.Fatal("asset-rental domain is not registered")
	}
	actions := map[string]*Action{}
	for index := range domain.Actions {
		actions[domain.Actions[index].Name] = &domain.Actions[index]
	}
	active := actions["active"]
	if active == nil || active.ToolName != "UteamupAssetrentalGetRented" ||
		len(active.Args) != 0 || len(active.Flags) != 0 {
		t.Fatalf("unexpected active-rental action: %+v", active)
	}
	expiring := actions["expiring"]
	if expiring == nil || expiring.ToolName != "UteamupAssetrentalGetExpiringSoon" ||
		len(expiring.Args) != 0 || len(expiring.Flags) != 1 ||
		expiring.Flags[0].Name != "days" {
		t.Fatalf("unexpected expiring-rental action: %+v", expiring)
	}
}

func TestAssetRentalRevenueActionUsesOnlyDateRange(t *testing.T) {
	domain := findDomain("asset-rental")
	if domain == nil {
		t.Fatal("asset-rental domain is not registered")
	}
	for index := range domain.Actions {
		action := &domain.Actions[index]
		if action.Name != "revenue" {
			continue
		}
		if action.ToolName != "UteamupAssetrentalRevenueSummary" ||
			len(action.Args) != 0 || len(action.Flags) != 2 ||
			action.Flags[0].Name != "start-date" || action.Flags[1].Name != "end-date" {
			t.Fatalf("unexpected rental revenue action: %+v", action)
		}
		return
	}
	t.Fatal("asset-rental revenue action is not registered")
}
