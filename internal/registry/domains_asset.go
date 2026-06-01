package registry

func init() {
	Register(&Domain{
		Name:        "asset",
		Aliases:     []string{"assets"},
		Description: "Manage assets and equipment inventory",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List all assets with pagination and filtering",
				ToolName:    "UteamupAssetList",
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
					{Name: "filter", Short: "f", Description: "Filter by name", Type: "string"},
					{Name: "sort-by", Description: "Sort field (Name, CreatedAt, etc.)", Default: "Name", Type: "string"},
					{Name: "sort-order", Description: "Sort direction (asc or desc)", Default: "asc", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get asset details by ID",
				ToolName:    "UteamupAssetGet",
				Args:        []ArgDef{{Name: "id", Description: "Asset ID", Required: true, Type: "int"}},
			},
			{
				Name:        "get-by-guid",
				Description: "Get asset details by stable ExternalGuid (URL-safe, survives migrations)",
				ToolName:    "UteamupAssetGetByGuid",
				Args:        []ArgDef{{Name: "guid", Description: "Asset ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "get-assigned-stock",
				Description: "Get parts/tools/chemicals assigned to an asset, with the stock locations each item sits in (quantity + low-stock state)",
				ToolName:    "UteamupAssetGetAssignedStock",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "get-documents-aggregated",
				Description: "Get ALL documents for an asset grouped by source — the asset's own documents plus its assigned parts', tools', chemicals', and industry code's documents (each tagged with sourceType/sourceGuid/sourceName)",
				ToolName:    "UteamupAssetGetAggregatedDocuments",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "create",
				Description: "Create a new asset",
				ToolName:    "UteamupAssetCreate",
				Flags: []FlagDef{
					{Name: "name", Description: "Asset name", Required: true, Type: "string"},
					{Name: "serial", Description: "Serial number", Type: "string"},
					// Deprecated single-type flag — still accepted and mapped onto the primary type server-side.
					{Name: "asset-type-id", Description: "Asset type ID (deprecated, use --asset-type-ids + --primary-asset-type-id)", Type: "int"},
					// Many-to-many: comma-separated list of type ids plus an explicit primary.
					{Name: "asset-type-ids", Description: "Comma-separated list of asset type ids (many-to-many)", Type: "string"},
					{Name: "primary-asset-type-id", Description: "Primary asset type id (must be one of --asset-type-ids)", Type: "int"},
					{Name: "location-id", Description: "Location ID", Type: "int"},
					{Name: "from-json", Description: "JSON file with asset data", Type: "string"},
				},
			},
			{
				Name:        "update",
				Description: "Update an existing asset",
				ToolName:    "UteamupAssetUpdate",
				Args:        []ArgDef{{Name: "id", Description: "Asset ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "name", Description: "New asset name", Type: "string"},
					{Name: "serial", Description: "New serial number", Type: "string"},
					{Name: "asset-type-ids", Description: "Comma-separated list of asset type ids (replaces current many-to-many set)", Type: "string"},
					{Name: "primary-asset-type-id", Description: "Primary asset type id (must be one of --asset-type-ids)", Type: "int"},
					{Name: "from-json", Description: "JSON file with update data", Type: "string"},
				},
			},
			{
				Name:        "get-specs",
				Description: "Get the effective attribute definitions (operating specs) for an asset, grouped per asset type",
				ToolName:    "UteamupAssetGetEffectiveAttributeDefinitions",
				Args:        []ArgDef{{Name: "id", Description: "Asset ID", Required: true, Type: "int"}},
			},
			{
				Name:        "delete",
				Description: "Delete an asset by ID",
				ToolName:    "UteamupAssetDelete",
				Args:        []ArgDef{{Name: "id", Description: "Asset ID", Required: true, Type: "int"}},
			},
			{
				Name:        "search",
				Description: "Search assets by name or serial number",
				ToolName:    "UteamupAssetSearch",
				Args:        []ArgDef{{Name: "query", Description: "Search term", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "set-responsible-owners",
				Description: "Set the responsible owners of an asset (replace-set; pass all owner user ids).",
				ToolName:    "UteamupAssetSetResponsibleOwners",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{assetGuid}/responsible-owners",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "user-ids", Description: "Responsible owner user ids — repeatable or comma-separated (replaces the current set)", Type: "stringSlice", BodyName: "userIds"},
				},
			},
		},
	})
}
