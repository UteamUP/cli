package registry

import "testing"

func TestBusinessCentralStockReadUsesTheNativeGuidRoute(t *testing.T) {
	action := findStockAction(t, "business-central")
	guid := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	path, consumed := buildRESTPath(findStockDomain(t), *action, map[string]any{"stockItemGuid": guid})
	if path != "/api/businesscentral/stock/"+guid || len(consumed) != 1 {
		t.Fatalf("unexpected BC stock route: %q, consumed %v", path, consumed)
	}
	if action.ToolName != "UteamupStockGetBusinessCentral" || len(action.Args) != 1 || action.Args[0].Type != "uuid" || !action.Args[0].Required {
		t.Fatalf("BC stock read must require a local stock GUID: %+v", action)
	}
}
