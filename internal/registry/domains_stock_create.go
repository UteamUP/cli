package registry

func stockCreateItemAction() Action {
	return Action{Name: "item-create", Description: "Create one stock row through MCP; check location items before retrying an uncertain result", ToolName: "UteamupStockCreateItem", MCPOnly: true, RejectExtraArgs: true,
		Flags: []FlagDef{
			{Name: "stock-guid", BodyName: "stockGuid", Description: "Writable stock location public GUID", Required: true, Type: "non-empty-uuid"},
			{Name: "name", Description: "Stock row name, up to 200 characters", Required: true, Type: "string"},
			{Name: "type", Description: "Catalog type", Required: true, Type: "string", AllowedValues: []string{"Part", "Tool", "Chemical"}},
			{Name: "part-guid", BodyName: "partGuid", Description: "Existing part GUID when type is Part", Type: "non-empty-uuid"},
			{Name: "tool-guid", BodyName: "toolGuid", Description: "Existing tool GUID when type is Tool", Type: "non-empty-uuid"},
			{Name: "chemical-guid", BodyName: "chemicalGuid", Description: "Existing chemical GUID when type is Chemical", Type: "non-empty-uuid"},
			{Name: "minimum-amount", BodyName: "minimumAmount", Description: "Minimum threshold in catalog units", Type: "decimal", Default: "0"},
			{Name: "maximum-amount", BodyName: "maximumAmount", Description: "Maximum threshold in catalog units; zero leaves unset", Type: "decimal", Default: "0"},
			{Name: "initial-stock", BodyName: "initialStock", Description: "Opening on-hand quantity; positive quantity posts an opening Add transaction", Type: "decimal", Default: "0"},
			{Name: "description", Description: "Description, up to 2,000 characters", Type: "string"},
			{Name: "unit-of-measure", BodyName: "unitOfMeasure", Description: "Unit for a new catalog item; existing catalog units are preserved", Type: "string"},
			{Name: "sku", Description: "SKU for a new catalog item; existing catalog SKUs are preserved", Type: "string"},
			{Name: "reorder-point", BodyName: "reorderPoint", Description: "Reorder point in catalog units", Type: "decimal"},
			{Name: "safety-stock", BodyName: "safetyStock", Description: "Safety stock in catalog units", Type: "decimal"},
			{Name: "lead-time-days", BodyName: "leadTimeDays", Description: "Lead time, 0-3650 days", Type: "int"},
			{Name: "criticality-override", BodyName: "criticalityOverride", Description: "MRO criticality A, B or C", Type: "string", AllowedValues: []string{"A", "B", "C"}},
			{Name: "service-level-target-percent", BodyName: "serviceLevelTargetPercent", Description: "Service target percent, 80-99.9", Type: "decimal"},
			{Name: "condition", Description: "New, Used, Refurbished or OpenBox", Type: "string", AllowedValues: []string{"New", "Used", "Refurbished", "OpenBox"}},
			{Name: "bin-guid", BodyName: "binGuid", Description: "Bin public GUID inside the chosen location", Type: "non-empty-uuid"},
		}}
}
