package registry

import "testing"

func TestStockAssignedAssetsUsesGuidAndPagedRead(t *testing.T) {
	action := findStockAction(t, "item-assigned-assets")
	if action.ToolName != "UteamupStockGetItemAssignedAssets" || action.HTTPMethod != "GET" {
		t.Fatalf("unexpected assigned assets contract: %+v", action)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "itemGuid" || action.Args[0].Type != "uuid" || !action.Args[0].Required {
		t.Fatalf("expected required public item GUID: %+v", action.Args)
	}
	guid := "11111111-1111-4111-8111-111111111111"
	path, consumed := buildRESTPath(findStockDomain(t), *action, map[string]any{"itemGuid": guid})
	if path != "/api/stock/items/"+guid+"/assigned-assets" || len(consumed) != 1 {
		t.Fatalf("unexpected URL %q, consumed %v", path, consumed)
	}
	pageNames := map[string]string{}
	for _, flag := range action.Flags {
		pageNames[flag.Name] = flag.QueryName
	}
	if pageNames["page"] != "page" || pageNames["page-size"] != "pageSize" {
		t.Fatalf("pagination must use the backend query names: %v", pageNames)
	}
}
