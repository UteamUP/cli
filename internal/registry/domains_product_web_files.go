package registry

func init() {
	Register(&Domain{
		Name:        "product-web-files",
		Aliases:     []string{"web-files"},
		Description: "Find product photos and documents for an asset, part, tool or chemical on the web and attach them",
		APIPath:     "/api/productwebfiles",
		Actions: []Action{
			{
				Name:        "search",
				Description: "Search the web for photos and documents of an item (charged once, refunded when nothing is found)",
				ToolName:    "uteamup_product_web_files_search",
				HTTPMethod:  "POST",
				RESTPath:    "search",
				Flags: []FlagDef{
					{Name: "item-kind", Description: "asset, part, tool or chemical", Required: true, Type: "string", AllowedValues: model3DItemKinds},
					{Name: "item-guid", Description: "GUID of the asset, part, tool or chemical", Required: true, Type: "non-empty-uuid"},
					{Name: "query", Description: "Search words (defaults to the item's manufacturer and model number)", Type: "string"},
				},
			},
			{
				Name:        "import",
				Description: "Fetch one photo or document URL from a search into Documents and attach it to the item",
				ToolName:    "uteamup_product_web_files_import",
				HTTPMethod:  "POST",
				RESTPath:    "import",
				Flags: []FlagDef{
					{Name: "item-kind", Description: "asset, part, tool or chemical", Required: true, Type: "string", AllowedValues: model3DItemKinds},
					{Name: "item-guid", Description: "GUID of the asset, part, tool or chemical", Required: true, Type: "non-empty-uuid"},
					{Name: "url", Description: "HTTPS URL of the photo or document", Required: true, Type: "string"},
					{Name: "page-url", Description: "The page the photo was found on", Type: "string"},
				},
			},
		},
	})
}
