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

func TestAssetRentalLifecycleUsesGuidArgumentsAndTypedRequestFiles(t *testing.T) {
	domain := findDomain("asset-rental")
	expected := map[string]string{"info": "GetByAsset", "history": "GetHistory", "configure": "CreateOrUpdate", "reserve": "Start", "extend": "Extend", "cancel": "Cancel", "checkout": "Checkout", "return": "End", "ready": "Ready"}
	for _, action := range domain.Actions {
		suffix, exists := expected[action.Name]
		if !exists {
			continue
		}
		if action.ToolName != "UteamupAssetrental"+suffix || len(action.Args) != 1 || action.Args[0].Type != "uuid" {
			t.Fatalf("unsafe or broken rental action: %+v", action)
		}
		if action.HTTPMethod == "POST" && (len(action.Flags) != 1 || !action.Flags[0].JSONFile || action.Flags[0].BodyName != "model") {
			t.Fatalf("rental action must preserve structured evidence: %+v", action)
		}
		delete(expected, action.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing rental actions: %v", expected)
	}
}

func TestAssetReadinessUsesGuidEvidenceActions(t *testing.T) {
	domain := findDomain("asset-readiness")
	if domain == nil || len(domain.Actions) != 6 {
		t.Fatal("six readiness actions must be registered")
	}
	for _, action := range domain.Actions {
		if len(action.Args) != 1 || action.Args[0].Type != "uuid" {
			t.Fatalf("readiness action leaks an internal identifier: %+v", action)
		}
		if action.HTTPMethod == "POST" && (len(action.Flags) != 1 || !action.Flags[0].JSONFile || action.Flags[0].BodyName != "model") {
			t.Fatalf("readiness evidence must remain structured: %+v", action)
		}
	}
}
