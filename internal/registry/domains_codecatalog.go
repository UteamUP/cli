package registry

func init() {
	Register(&Domain{
		Name:        "codecatalog",
		Aliases:     []string{"cc", "codes"},
		Description: "Search and browse the industry code catalog",
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
		},
	})
}
