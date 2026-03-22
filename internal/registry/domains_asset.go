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
				Name:        "create",
				Description: "Create a new asset",
				ToolName:    "UteamupAssetCreate",
				Flags: []FlagDef{
					{Name: "name", Description: "Asset name", Required: true, Type: "string"},
					{Name: "serial", Description: "Serial number", Type: "string"},
					{Name: "asset-type-id", Description: "Asset type ID", Type: "int"},
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
					{Name: "from-json", Description: "JSON file with update data", Type: "string"},
				},
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
		},
	})
}
