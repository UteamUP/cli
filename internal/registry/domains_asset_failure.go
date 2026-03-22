package registry

func init() {
	Register(&Domain{
		Name:        "asset-failure",
		Description: "Manage asset failure records",
		Actions: []Action{
			{Name: "list", Description: "List failure records", ToolName: "UteamupAssetFailureList", Flags: append(paginationFlags(), FlagDef{Name: "asset-id", Description: "Asset ID", Type: "int"})},
			{Name: "get", Description: "Get failure record by ID", ToolName: "UteamupAssetFailureGet", Args: idArg()},
			{Name: "create", Description: "Create a failure record", ToolName: "UteamupAssetFailureCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete a failure record", ToolName: "UteamupAssetFailureDelete", Args: idArg()},
		},
	})
}
