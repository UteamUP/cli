package registry

func init() {
	Register(&Domain{
		Name:        "asset-type",
		APIPath:     "/api/asset-type",
		Aliases:     []string{"assettypes", "at"},
		Description: "Manage asset types",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List asset types",
				ToolName:    "UteamupAssetTypeList",
				Flags: []FlagDef{
					{Name: "include-inactive", QueryName: "includeInactive", Description: "Include inactive asset types", Type: "bool"},
				},
			},
			{
				Name:        "get",
				RESTPath:    "{assetTypeGuid}",
				HTTPMethod:  "GET",
				Description: "Get an asset type by public GUID",
				ToolName:    "UteamupAssetTypeGet",
				Args: []ArgDef{
					{Name: "assetTypeGuid", Description: "Asset type public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create an asset type from reviewed JSON",
				ToolName:    "UteamupAssetTypeCreate",
				Flags: []FlagDef{
					{Name: "request-file", Short: "f", Description: "JSON file containing the GUID-only asset type creation model", Required: true, Type: "string", RootJSONObjectFile: true},
				},
			},
			{
				Name:        "update",
				RESTPath:    "{assetTypeGuid}",
				HTTPMethod:  "PUT",
				Description: "Update an asset type by public GUID from reviewed JSON",
				ToolName:    "UteamupAssetTypeUpdate",
				Args: []ArgDef{
					{Name: "assetTypeGuid", Description: "Asset type public GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "request-file", Short: "f", Description: "JSON file containing the asset type update model", Required: true, Type: "string", RootJSONObjectFile: true},
				},
			},
			{
				Name:        "delete",
				RESTPath:    "{assetTypeGuid}",
				HTTPMethod:  "DELETE",
				Description: "Delete an asset type by public GUID",
				ToolName:    "UteamupAssetTypeDelete",
				Args: []ArgDef{
					{Name: "assetTypeGuid", Description: "Asset type public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "fill-existing-assets",
				RESTPath:    "attributes/{attributeGuid}/fill-existing",
				HTTPMethod:  "POST",
				Description: "Write one value for a field on every existing asset of its type that has none yet",
				ToolName:    "UteamupAssetTypeFillExistingAssets",
				Args: []ArgDef{
					{Name: "attributeGuid", Description: "Attribute definition public GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "value", BodyName: "rawValue", Description: "Value parsed by the field's data type (e.g. 12.5, true, 2026-10-05)", Required: true, Type: "string"},
				},
			},
			// --- Reseller catalog: reverse fitment lookup (stock-reseller-catalog §12) ---
			{
				Name:        "compatible-parts",
				RESTPath:    "by-guid/{assetTypeGuid}/compatible-parts",
				HTTPMethod:  "GET",
				Description: "List the parts declared compatible with (that fit) an asset type",
				ToolName:    "UteamupAssetTypeListCompatibleParts",
				Args: []ArgDef{
					{Name: "assetTypeGuid", Description: "Asset type public GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
