package registry

func init() {
	Register(&Domain{
		Name:        "asset-certification",
		Description: "Manage asset certifications",
		Actions: []Action{
			{Name: "list", Description: "List certifications", ToolName: "UteamupAssetCertificationList", Flags: append(paginationFlags(), FlagDef{Name: "asset-id", Description: "Asset ID", Type: "int"})},
			{Name: "get", Description: "Get certification by ID", ToolName: "UteamupAssetCertificationGet", Args: idArg()},
			{Name: "create", Description: "Create a certification", ToolName: "UteamupAssetCertificationCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a certification", ToolName: "UteamupAssetCertificationUpdate", Args: idArg(), Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete a certification", ToolName: "UteamupAssetCertificationDelete", Args: idArg()},
		},
	})
}
