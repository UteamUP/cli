package registry

func init() {
	Register(&Domain{
		Name:        "asset-energy",
		Description: "Manage asset energy records",
		Actions: []Action{
			{Name: "list", Description: "List energy records", ToolName: "UteamupAssetEnergyRecordList", Flags: append(paginationFlags(), FlagDef{Name: "asset-id", Description: "Asset ID", Type: "int"})},
			{Name: "get", Description: "Get energy record by ID", ToolName: "UteamupAssetEnergyRecordGet", Args: idArg()},
			{Name: "create", Description: "Create an energy record", ToolName: "UteamupAssetEnergyRecordCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete an energy record", ToolName: "UteamupAssetEnergyRecordDelete", Args: idArg()},
		},
	})
}
