package registry

var model3DCategories = []string{"asset", "part", "tool", "chemical", "other"}

// The catalogue's Global Admin review routes (/api/admin/model3d-catalogue) are
// deliberately not exposed here: moderation stays in the admin web UI.
func init() {
	Register(&Domain{
		Name:        "model3d-catalogue",
		Aliases:     []string{"models3d-catalogue", "3d-catalogue"},
		Description: "The cross-company 3D model catalogue: search reviewed models, copy one into the library, share and withdraw your own",
		APIPath:     "/api/model3d/catalogue",
		Actions: []Action{
			{
				Name:        "search",
				Description: "Search published catalogue models by name, manufacturer, model or part number, or GTIN; best match first",
				ToolName:    "uteamup_model3d_catalogue_search",
				Flags: []FlagDef{
					{Name: "search", Description: "Part of the name, or a model, part or GTIN number", Type: "string"},
					{Name: "manufacturer", Description: "Manufacturer name", Type: "string"},
					{Name: "model-number", Description: "Model number", Type: "string"},
					{Name: "part-number", Description: "Part number", Type: "string"},
					{Name: "gtin", Description: "GTIN / EAN barcode", Type: "string"},
					{Name: "category", Description: "asset, part, tool, chemical or other", Type: "string", AllowedValues: model3DCategories},
					{Name: "page", Short: "p", Description: "Page number (1-based)", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page (max 50)", Default: 20, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one published catalogue entry: product identity, size in metres, triangles and version",
				ToolName:    "uteamup_model3d_catalogue_get",
				HTTPMethod:  "GET",
				RESTPath:    "{entryGuid}",
				Args: []ArgDef{
					{Name: "entryGuid", Description: "Catalogue entry GUID", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:        "copy",
				Description: "Copy a catalogue model into your 3D model library, optionally linking it to an asset, part, tool or chemical",
				ToolName:    "uteamup_model3d_catalogue_copy",
				HTTPMethod:  "POST",
				RESTPath:    "copy",
				Flags: []FlagDef{
					{Name: "entry-guid", Description: "Catalogue entry GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "item-kind", Description: "asset, part, tool or chemical to link the copy to", Type: "string", AllowedValues: model3DItemKinds},
					{Name: "item-guid", Description: "GUID of the item to link the copy to", Type: "non-empty-uuid"},
				},
			},
			{
				Name:        "share",
				Description: "Share a copy of one of your 3D models to the catalogue; a UteamUP admin reviews it and your company is never named",
				ToolName:    "uteamup_model3d_catalogue_share",
				HTTPMethod:  "POST",
				RESTPath:    "share",
				Flags: []FlagDef{
					{Name: "model-guid", BodyName: "model3DGuid", Description: "3D model GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "name", Description: "Name shown in the catalogue", Required: true, Type: "string"},
					{Name: "category", Description: "asset, part, tool, chemical or other", Default: "other", Type: "string", AllowedValues: model3DCategories},
					{Name: "manufacturer", Description: "Manufacturer name", Type: "string"},
					{Name: "model-number", Description: "Model number", Type: "string"},
					{Name: "part-number", Description: "Part number", Type: "string"},
					{Name: "gtin", Description: "GTIN / EAN barcode", Type: "string"},
					{Name: "rights-confirmed", Description: "Confirm your company may share this model and its photos with every UteamUP customer", Required: true, MustBeTrue: true, Type: "bool"},
					{Name: "contains-online-photos", Description: "Some photos on the model came from the web", Default: false, Type: "bool"},
				},
			},
			{
				Name:        "withdraw",
				Description: "Take one of your shared models out of the catalogue; copies other companies made are kept",
				ToolName:    "uteamup_model3d_catalogue_withdraw",
				HTTPMethod:  "POST",
				RESTPath:    "{entryGuid}/withdraw",
				Args: []ArgDef{
					{Name: "entryGuid", Description: "Catalogue entry GUID", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:        "mine",
				Description: "List the models your company shared with their review status",
				ToolName:    "uteamup_model3d_catalogue_mine",
				HTTPMethod:  "GET",
				RESTPath:    "mine",
			},
		},
	})
}
