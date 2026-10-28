package registry

// Planned incoming is metadata; even Received does not post an inventory movement.
func stockPlannedIncomingActions() []Action {
	item := ArgDef{Name: "stockItemGuid", Description: "Stock item public GUID", Required: true, Type: "uuid"}
	plan := ArgDef{Name: "plannedIncomingGuid", Description: "Planned incoming public GUID", Required: true, Type: "uuid"}
	version := FlagDef{Name: "expected-version-guid", BodyName: "expectedVersionGuid", Description: "Version GUID from the reviewed plan; reuse for uncertain conversion retries", Required: true, Type: "uuid"}
	fields := func() []FlagDef {
		return []FlagDef{
			{Name: "quantity", BodyName: "quantity", Description: "Positive quantity in the stock item's unit", Required: true, Type: "decimal"},
			{Name: "expected-date", BodyName: "expectedDate", Description: "YYYY-MM-DD; use day 01 for Month precision", Required: true, Type: "string"},
			{Name: "date-precision", BodyName: "datePrecision", Description: "Day or Month", Default: "Day", Type: "string"},
			{Name: "vendor-guid", BodyName: "vendorGuid", Description: "Optional readable tenant vendor; omitted on update clears", Type: "uuid"},
			{Name: "project-guid", BodyName: "projectGuid", Description: "Optional readable tenant project; omitted on update clears", Type: "uuid"},
			{Name: "stage-guid", BodyName: "stageGuid", Description: "Optional stage of the project; omitted on update clears", Type: "uuid"},
			{Name: "notes", BodyName: "notes", Description: "Optional notes, up to 2000 characters; omitted on update clears", Type: "string"},
		}
	}
	return []Action{
		{Name: "planned-incoming-list", Description: "List planned incoming separately from actual and ordered stock", ToolName: "UteamupStockListPlannedIncoming", HTTPMethod: "GET", RESTPath: "items/{stockItemGuid}/planned-incoming", Args: []ArgDef{item}, Flags: []FlagDef{
			{Name: "page", QueryName: "page", Description: "Page starting at 1", Default: 1, Type: "int"},
			{Name: "page-size", QueryName: "pageSize", Description: "Page size 1-200", Default: 50, Type: "int"},
		}},
		{Name: "planned-incoming-get", Description: "Read one plan and its current version", ToolName: "UteamupStockGetPlannedIncoming", HTTPMethod: "GET", RESTPath: "planned-incoming/{plannedIncomingGuid}", Args: []ArgDef{plan}},
		{Name: "planned-incoming-create", Description: "Create planning metadata without changing inventory or sending an order", ToolName: "UteamupStockCreatePlannedIncoming", HTTPMethod: "POST", RESTPath: "items/{stockItemGuid}/planned-incoming", Args: []ArgDef{item}, Flags: fields()},
		{Name: "planned-incoming-update", Description: "Replace unconverted planning metadata; Received status never changes inventory", ToolName: "UteamupStockUpdatePlannedIncoming", HTTPMethod: "PUT", RESTPath: "planned-incoming/{plannedIncomingGuid}", Args: []ArgDef{plan}, Flags: append(fields(), version,
			FlagDef{Name: "status", BodyName: "status", Description: "Planned, Ordered, Received or Cancelled", Required: true, Type: "string"})},
		{Name: "planned-incoming-convert", Description: "Convert one reviewed plan into an internal Draft PO; never send or receive", ToolName: "UteamupStockConvertPlannedIncoming", HTTPMethod: "POST", RESTPath: "planned-incoming/{plannedIncomingGuid}/purchase-order-draft", Args: []ArgDef{plan}, Flags: []FlagDef{version}},
	}
}
