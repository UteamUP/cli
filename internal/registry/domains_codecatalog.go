package registry

func init() {
	Register(&Domain{
		Name:        "codecatalog",
		Aliases:     []string{"cc", "codes"},
		Description: "Search, update, and deactivate industry code catalog entries",
		APIPath:     "/api/codecatalog",
		Actions: []Action{
			{
				Name:        "search",
				Description: "Search industry codes by text (mention search)",
				ToolName:    "UteamupCodeCatalogSearch",
				RESTPath:    "mention-search",
				Args:        []ArgDef{{Name: "query", Description: "Search term (e.g., 'pump', '1-HLA')", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "limit", Short: "l", Description: "Maximum results to return", Default: 10, Type: "int"},
				},
			},
			{
				Name:        "update-by-guid",
				Description: "Update a code catalog entry by stable ExternalGuid",
				ToolName:    "UteamupCodingsystemUpdateEntryByGuid",
				Args: []ArgDef{
					{Name: "guid", Description: "Entry ExternalGuid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "name", Description: "New display name", Type: "string"},
					{Name: "description", Description: "New description", Type: "string"},
					{Name: "component-type-code", Description: "New component type code", Type: "string"},
					{Name: "component-type-description", Description: "New component type description", Type: "string"},
					{Name: "drawing-reference", Description: "New drawing reference", Type: "string"},
				},
			},
			{
				Name:        "deactivate-by-guid",
				Description: "Deactivate a code catalog entry by ExternalGuid (cascades to descendants)",
				ToolName:    "UteamupCodingsystemDeactivateEntryByGuid",
				Args: []ArgDef{
					{Name: "guid", Description: "Entry ExternalGuid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "remove-asset-assignment",
				Description: "Remove the code assignment from an asset by its ExternalGuid — preserves audit log",
				ToolName:    "UteamupCodingsystemRemoveAssetAssignment",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset ExternalGuid", Required: true, Type: "string"},
				},
			},
		},
	})
}
