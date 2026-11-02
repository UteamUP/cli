package registry

import "testing"

func TestStockTransactionAssetLinkUsesMetadataRouteAndGuidBody(t *testing.T) {
	action := findStockAction(t, "transaction-link-asset")
	if action.ToolName != "UteamupStockTransactionLinkAsset" || action.HTTPMethod != "POST" {
		t.Fatalf("unexpected action: %+v", action)
	}
	guid := "11111111-1111-4111-8111-111111111111"
	path, consumed := buildRESTPath(findStockDomain(t), *action, map[string]any{"transactionGuid": guid})
	if path != "/api/stock/transactions/"+guid+"/asset" || len(consumed) != 1 {
		t.Fatalf("wrong metadata route: %q %+v", path, consumed)
	}
	if len(action.Flags) != 1 || action.Flags[0].BodyName != "assetGuid" || !action.Flags[0].Required {
		t.Fatalf("must carry only the target equipment reference, never a second quantity: %+v", action.Flags)
	}
}
