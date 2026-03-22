package registry

func init() {
	Register(&Domain{
		Name:        "asset-type",
		Aliases:     []string{"assettypes", "at"},
		Description: "Manage asset types",
		Actions: []Action{
			{Name: "list", Description: "List asset types", ToolName: "UteamupAssetTypeList", Flags: paginationFlags()},
			{Name: "get", Description: "Get asset type by ID", ToolName: "UteamupAssetTypeGet", Args: idArg()},
			{Name: "create", Description: "Create an asset type", ToolName: "UteamupAssetTypeCreate", Flags: []FlagDef{{Name: "name", Description: "Asset type name", Required: true, Type: "string"}, jsonFlag()}},
			{Name: "update", Description: "Update an asset type", ToolName: "UteamupAssetTypeUpdate", Args: idArg(), Flags: []FlagDef{{Name: "name", Description: "New name", Type: "string"}, jsonFlag()}},
			{Name: "delete", Description: "Delete an asset type", ToolName: "UteamupAssetTypeDelete", Args: idArg()},
		},
	})
}
