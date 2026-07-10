package registry

func init() {
	Register(&Domain{
		Name:        "asset-type",
		Aliases:     []string{"assettypes", "at"},
		Description: "Manage asset types",
		// The controller routes at api/asset-type (hyphenated). Without this override the
		// base path would derive to /api/assettype (hyphens stripped), which the backend
		// does not serve.
		APIPath: "/api/asset-type",
		Actions: []Action{
			{Name: "list", Description: "List asset types", ToolName: "UteamupAssetTypeList", Flags: paginationFlags()},
			{Name: "get", Description: "Get asset type by GUID", ToolName: "UteamupAssetTypeGet", Args: externalGuidArg()},
			{Name: "create", Description: "Create an asset type", ToolName: "UteamupAssetTypeCreate", Flags: []FlagDef{{Name: "name", Description: "Asset type name", Required: true, Type: "string"}, jsonFlag()}},
			{Name: "update", Description: "Update an asset type", ToolName: "UteamupAssetTypeUpdate", Args: externalGuidArg(), Flags: []FlagDef{{Name: "name", Description: "New name", Type: "string"}, jsonFlag()}},
			{Name: "delete", Description: "Delete an asset type", ToolName: "UteamupAssetTypeDelete", Args: externalGuidArg()},
			// --- Reseller catalog: reverse fitment lookup (stock-reseller-catalog §12) ---
			{
				Name:        "compatible-parts",
				Description: "List the parts declared compatible with (that fit) an asset type",
				ToolName:    "UteamupAssetTypeListCompatibleParts",
				RESTPath:    "by-guid/{guid}/compatible-parts",
				Args:        []ArgDef{{Name: "guid", Description: "Asset type GUID", Required: true, Type: "string"}},
			},
		},
	})
}
