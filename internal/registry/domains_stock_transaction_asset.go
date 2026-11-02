package registry

func stockTransactionAssetAction() Action {
	return Action{
		Name: "transaction-link-asset", Description: "Attribute an applied removal to equipment without withdrawing stock again",
		ToolName: "UteamupStockTransactionLinkAsset", HTTPMethod: "POST", RESTPath: "transactions/{transactionGuid}/asset",
		Args:  []ArgDef{{Name: "transactionGuid", Description: "Applied Remove movement public GUID", Required: true, Type: "uuid"}},
		Flags: []FlagDef{{Name: "asset-guid", BodyName: "assetGuid", Description: "Equipment public GUID; same-asset retries are idempotent", Required: true, Type: "non-empty-uuid"}},
	}
}
