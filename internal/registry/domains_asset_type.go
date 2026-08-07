package registry

func init() {
	Register(&Domain{
		Name:        "asset-type",
		Aliases:     []string{"assettypes", "at"},
		Description: "Manage asset types",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List asset types",
				ToolName:    "UteamupAssetTypeList",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "include-inactive", BodyName: "includeInactive", Description: "Include inactive asset types", Type: "bool"},
				},
			},
			{
				Name:        "get",
				Description: "Get an asset type by public GUID",
				ToolName:    "UteamupAssetTypeGet",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "assetTypeGuid", Description: "Asset type public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create an asset type from reviewed JSON",
				ToolName:    "UteamupAssetTypeCreate",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "request-file", Short: "f", BodyName: "model", Description: "JSON file containing the GUID-only asset type creation model", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "update",
				Description: "Update an asset type by public GUID from reviewed JSON",
				ToolName:    "UteamupAssetTypeUpdate",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "assetTypeGuid", Description: "Asset type public GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "request-file", Short: "f", BodyName: "model", Description: "JSON file containing the asset type update model", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "delete",
				Description: "Delete an asset type by public GUID",
				ToolName:    "UteamupAssetTypeDelete",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "assetTypeGuid", Description: "Asset type public GUID", Required: true, Type: "uuid"},
				},
			},
			// --- Reseller catalog: reverse fitment lookup (stock-reseller-catalog §12) ---
			{
				Name:        "compatible-parts",
				Description: "List the parts declared compatible with (that fit) an asset type",
				ToolName:    "UteamupAssetTypeListCompatibleParts",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "assetTypeGuid", Description: "Asset type public GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
