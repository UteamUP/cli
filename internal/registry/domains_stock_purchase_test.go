package registry

import "testing"

func TestStockPurchaseRoutesUseGuidIdentityAndExistingPurchaseLedger(t *testing.T) {
	guid := "11111111-1111-4111-8111-111111111111"
	for _, entry := range []struct{ name, tool, path, arg string }{
		{"purchase-get", "UteamupStockGetPurchaseTransaction", "purchase-transactions/" + guid, "purchaseGuid"},
		{"purchase-list", "UteamupStockGetPurchaseTransactions", "items/" + guid + "/purchase-transactions", "stockItemGuid"},
	} {
		action := findStockAction(t, entry.name)
		if action.HTTPMethod != "GET" || action.ToolName != entry.tool || len(action.Args) != 1 || action.Args[0].Type != "non-empty-uuid" || !action.RejectExtraArgs {
			t.Fatalf("incorrect purchase read contract: %+v", action)
		}
		path, consumed := buildRESTPath(findStockDomain(t), *action, map[string]any{entry.arg: guid})
		if path != "/api/stock/"+entry.path || len(consumed) != 1 {
			t.Fatalf("%s: wrong purchase route %q", entry.name, path)
		}
	}
}

func TestStockPurchaseCreateMirrorsReceiptGuidsAndExactDecimalFlags(t *testing.T) {
	action := findStockAction(t, "purchase-add")
	if action.ToolName != "UteamupStockAddTransaction" || action.HTTPMethod != "POST" || action.RESTPath != "purchase-transactions" || !action.RejectExtraArgs {
		t.Fatalf("must use the existing purchase operation: %+v", action)
	}
	flags := make(map[string]FlagDef)
	for _, flag := range action.Flags {
		if flag.Name == "tenant-guid" || flag.Name == "actor" || flag.Name == "id" {
			t.Fatalf("server-owned or numeric identity must not be public: %+v", flag)
		}
		flags[flag.Name] = flag
	}
	for _, name := range []string{"stock-guid", "stock-item-guid"} {
		if flags[name].Type != "non-empty-uuid" || !flags[name].Required {
			t.Fatalf("%s must be a required public GUID", name)
		}
	}
	for _, name := range []string{"amount", "purchase-price"} {
		if flags[name].Type != "decimal" {
			t.Fatalf("%s must preserve decimal input", name)
		}
	}
	if !flags["amount"].Required || !flags["purchase-date"].Required || flags["document-guids"].Type != "stringSlice" || flags["document-guids"].BodyName != "documentGuids" || flags["document-guids"].Required {
		t.Fatal("purchase create must preserve optional receipt attachments and required purchase inputs")
	}
}
