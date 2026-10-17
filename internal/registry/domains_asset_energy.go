package registry

func init() {
	filters := []FlagDef{
		{Name: "asset-guid", BodyName: "assetGuid", Description: "Filter by public asset GUID", Type: "non-empty-uuid"},
		{Name: "from", Description: "Inclusive first record date (YYYY-MM-DD)", Type: "string"},
		{Name: "to", Description: "Inclusive last record date (YYYY-MM-DD)", Type: "string"},
	}
	recordArg := []ArgDef{{Name: "record-guid", BodyName: "recordGuid", Description: "Public energy record GUID", Required: true, Type: "non-empty-uuid"}}
	model := []FlagDef{{Name: "request-file", Short: "f", BodyName: "model", Description: "JSON file containing the reviewed energy record model", Required: true, Type: "string", JSONFile: true}}
	Register(&Domain{
		Name:        "asset-energy",
		Description: "Manage asset energy records",
		APIPath:     "/api/assets",
		Actions: []Action{
			{Name: "list", Description: "List one bounded page of tenant energy records", ToolName: "UteamupAssetEnergyRecordList", RESTPath: "energy", RejectExtraArgs: true,
				Flags: append(append([]FlagDef{}, filters...),
					FlagDef{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					FlagDef{Name: "page-size", Short: "s", Description: "Items per page (1-200)", Default: 100, Type: "int"},
					FlagDef{Name: "sort-by", Description: "Global record sort field", Default: "recordDate", Type: "string", AllowedValues: []string{"recordDate", "fuelConsumedLitres", "energyConsumedKWh", "co2EmittedGrams", "distanceKm"}},
					FlagDef{Name: "sort-order", Description: "Global record sort order", Default: "desc", Type: "string", AllowedValues: []string{"asc", "desc"}})},
			{Name: "summary", Description: "Summarize tenant energy records using the same GUID and date filters", ToolName: "UteamupAssetEnergyRecordGetSummary", RESTPath: "energy/summary", HTTPMethod: "GET", Flags: filters, RejectExtraArgs: true},
			{Name: "get", Description: "Get an energy record by public GUID", ToolName: "UteamupAssetEnergyRecordGet", MCPOnly: true, Args: recordArg, RejectExtraArgs: true},
			{Name: "create", Description: "Create an energy record for a public asset GUID", ToolName: "UteamupAssetEnergyRecordCreate", MCPOnly: true,
				Args: []ArgDef{{Name: "asset-guid", BodyName: "assetGuid", Description: "Public asset GUID", Required: true, Type: "non-empty-uuid"}}, Flags: model, RejectExtraArgs: true},
			{Name: "update", Description: "Update an energy record by public GUID", ToolName: "UteamupAssetEnergyRecordUpdate", MCPOnly: true, Args: recordArg, Flags: model, RejectExtraArgs: true},
			{Name: "delete", Description: "Delete an energy record by public GUID", ToolName: "UteamupAssetEnergyRecordDelete", MCPOnly: true, Args: recordArg, RejectExtraArgs: true},
		},
	})
}
