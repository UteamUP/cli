package registry

// Tool and energy usage recorded on work orders. The backend MCP tools apply the
// WorkorderToolUsage.* and WorkorderEnergyUsage.* permissions of the REST routes.
func init() {
	page := []FlagDef{
		{Name: "page", Description: "Page number, starting at 1", Type: "int", Default: 1},
		{Name: "page-size", Description: "Records per page (1-100)", Type: "int", Default: 25},
	}
	usageGuid := []ArgDef{{Name: "usage-guid", BodyName: "usageGuid", Description: "Usage record public GUID", Required: true, Type: "non-empty-uuid"}}
	Register(&Domain{
		Name:        "workorder-usage",
		Aliases:     []string{"workorder-usages"},
		Description: "Review and correct tool and energy usage recorded on work orders",
		Actions: []Action{
			{Name: "tool-list", Description: "List tool usage records, newest first", ToolName: "UteamupWorkorderToolUsageList", MCPOnly: true, RejectExtraArgs: true, Flags: page},
			{Name: "tool-get", Description: "Get a tool usage record", ToolName: "UteamupWorkorderToolUsageGet", MCPOnly: true, RejectExtraArgs: true, Args: usageGuid},
			{Name: "tool-update", Description: "Replace a tool usage record (send every field)", ToolName: "UteamupWorkorderToolUsageUpdate", MCPOnly: true, RejectExtraArgs: true, Args: usageGuid,
				Flags: []FlagDef{
					{Name: "workorder-guid", Description: "Work order public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "tool-guid", Description: "Tool public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "start-time", Description: "Start time (ISO 8601, UTC)", Required: true, Type: "string"},
					{Name: "end-time", Description: "End time (ISO 8601, UTC)", Type: "string"},
					{Name: "duration-minutes", Description: "Duration in minutes", Type: "int"},
					{Name: "rate-type", Description: "Rate type, for example Hourly or Daily", Type: "string"},
					{Name: "rental-rate", Description: "Rental rate per rate type unit", Type: "decimal"},
					{Name: "notes", Description: "Notes", Type: "string"},
					{Name: "user-id", Description: "User id of the tenant member who used the tool", Type: "string"},
				}},
			{Name: "energy-list", Description: "List energy usage records, newest first", ToolName: "UteamupWorkorderEnergyUsageList", MCPOnly: true, RejectExtraArgs: true, Flags: page},
			{Name: "energy-get", Description: "Get an energy usage record", ToolName: "UteamupWorkorderEnergyUsageGet", MCPOnly: true, RejectExtraArgs: true, Args: usageGuid},
			{Name: "energy-update", Description: "Replace an energy usage record (send every field)", ToolName: "UteamupWorkorderEnergyUsageUpdate", MCPOnly: true, RejectExtraArgs: true, Args: usageGuid,
				Flags: []FlagDef{
					{Name: "workorder-guid", Description: "Work order public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "utility-type", Description: "electricity, naturalGas, water, steam, compressedAir, diesel, propane or fuelOil", Required: true, Type: "string"},
					{Name: "consumption", Description: "Amount consumed", Required: true, Type: "decimal"},
					{Name: "asset-guid", Description: "Asset public GUID", Type: "uuid"},
					{Name: "unit", Description: "Unit, for example kWh", Type: "string"},
					{Name: "cost-per-unit", Description: "Cost per unit", Type: "decimal"},
					{Name: "total-cost", Description: "Total cost", Type: "decimal"},
					{Name: "notes", Description: "Notes", Type: "string"},
				}},
		},
	})
}
