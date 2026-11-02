package registry

func stockPurchaseActions() []Action {
	return []Action{
		{Name: "purchase-add", Description: "Record one purchase receipt; check purchase history before retrying an uncertain result. Requires StockTransaction.Create and Document.View when linking receipts", ToolName: "UteamupStockAddTransaction", HTTPMethod: "POST", RESTPath: "purchase-transactions", RejectExtraArgs: true,
			Flags: []FlagDef{
				{Name: "stock-guid", BodyName: "stockGuid", Description: "Owning stock location public GUID", Required: true, Type: "non-empty-uuid"},
				{Name: "stock-item-guid", BodyName: "stockItemGuid", Description: "Stock item public GUID", Required: true, Type: "non-empty-uuid"},
				{Name: "amount", Description: "Exact positive received quantity in stock units", Required: true, Type: "decimal"},
				{Name: "purchase-date", BodyName: "purchaseDate", Description: "Purchase date in ISO 8601 format", Required: true, Type: "string"},
				{Name: "purchase-price", BodyName: "purchasePrice", Description: "Exact unit purchase price", Type: "decimal"},
				{Name: "currency-guid", BodyName: "currencyGuid", Description: "Currency public GUID", Type: "non-empty-uuid"},
				{Name: "vendor-guid", BodyName: "vendorGuid", Description: "Vendor public GUID", Type: "non-empty-uuid"},
				{Name: "reason", Description: "Purchase reason, up to 1,000 characters", Type: "string"},
				{Name: "purchase-order-number", BodyName: "purchaseOrderNumber", Description: "Purchase order reference", Type: "string"},
				{Name: "invoice-number", BodyName: "invoiceNumber", Description: "Invoice reference", Type: "string"},
				{Name: "document-guids", BodyName: "documentGuids", Description: "Up to 20 existing active tenant receipt document GUIDs, repeatable or comma-separated; the complete set is validated before stock changes", Type: "stringSlice"},
			}},
		{Name: "purchase-get", Description: "Read a purchase by GUID; receipt metadata requires current Document.View", ToolName: "UteamupStockGetPurchaseTransaction", HTTPMethod: "GET", RESTPath: "purchase-transactions/{purchaseGuid}", RejectExtraArgs: true,
			Args: []ArgDef{{Name: "purchaseGuid", Description: "Purchase public GUID", Required: true, Type: "non-empty-uuid"}}},
		{Name: "purchase-list", Description: "List a stock item's purchases newest first; receipt metadata requires current Document.View", ToolName: "UteamupStockGetPurchaseTransactions", HTTPMethod: "GET", RESTPath: "items/{stockItemGuid}/purchase-transactions", RejectExtraArgs: true,
			Args: []ArgDef{{Name: "stockItemGuid", Description: "Stock item public GUID", Required: true, Type: "non-empty-uuid"}}},
	}
}
