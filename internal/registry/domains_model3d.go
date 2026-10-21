package registry

var model3DItemKinds = []string{"asset", "part", "tool", "chemical"}

func init() {
	Register(&Domain{
		Name:        "model3d",
		Aliases:     []string{"models3d", "3d-model"},
		Description: "The tenant's 3D model library and the models assets, parts, tools and chemicals draw",
		APIPath:     "/api/model3d",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List 3D models (name search, category, paging)",
				ToolName:    "uteamup_model3d_list",
				Flags: []FlagDef{
					{Name: "search", Description: "Part of the model name", Type: "string"},
					{Name: "category", Description: "asset, part, tool, chemical or other", Type: "string", AllowedValues: []string{"asset", "part", "tool", "chemical", "other"}},
					{Name: "page", Short: "p", Description: "Page number (1-based)", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page (max 100)", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one 3D model with its editor source and document GUIDs",
				ToolName:    "uteamup_model3d_get",
				Args: []ArgDef{
					{Name: "externalGuid", Description: "3D model GUID", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:        "duplicate",
				Description: "Copy a 3D model into a new, independent model that can be edited without changing the original",
				ToolName:    "uteamup_model3d_duplicate",
				HTTPMethod:  "POST",
				RESTPath:    "{externalGuid}/duplicate",
				Args: []ArgDef{
					{Name: "externalGuid", Description: "GUID of the 3D model to copy", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{
					{Name: "name", Description: "Name of the copy (1-256 characters)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a 3D model; one still used by items is refused unless --force unlinks them first",
				ToolName:    "uteamup_model3d_delete",
				Args: []ArgDef{
					{Name: "externalGuid", Description: "3D model GUID", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{
					{Name: "force", QueryName: "force", Description: "Unlink the items that use the model, then delete it", Default: false, Type: "bool"},
				},
			},
			{
				Name:        "link",
				Description: "Point an asset, part, tool or chemical at a library 3D model",
				ToolName:    "uteamup_model3d_link",
				HTTPMethod:  "PUT",
				RESTPath:    "link",
				Flags: []FlagDef{
					{Name: "item-kind", Description: "asset, part, tool or chemical", Required: true, Type: "string", AllowedValues: model3DItemKinds},
					{Name: "item-guid", Description: "GUID of the asset, part, tool or chemical", Required: true, Type: "non-empty-uuid"},
					{Name: "model-guid", BodyName: "model3DGuid", Description: "3D model GUID", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				// Same route as link with no model: the item goes back to its digital twin or a box.
				Name:        "unlink",
				Description: "Clear the library 3D model of an asset, part, tool or chemical",
				ToolName:    "uteamup_model3d_link",
				HTTPMethod:  "PUT",
				RESTPath:    "link",
				Flags: []FlagDef{
					{Name: "item-kind", Description: "asset, part, tool or chemical", Required: true, Type: "string", AllowedValues: model3DItemKinds},
					{Name: "item-guid", Description: "GUID of the asset, part, tool or chemical", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:        "resolve",
				Description: "Show the 3D model an item draws: its library model, else an asset's digital twin, else none",
				ToolName:    "uteamup_model3d_resolve",
				HTTPMethod:  "GET",
				RESTPath:    "item/{itemKind}/{itemGuid}",
				Args: []ArgDef{
					{Name: "itemKind", Description: "asset, part, tool or chemical", Required: true, Type: "string", AllowedValues: model3DItemKinds},
					{Name: "itemGuid", Description: "GUID of the asset, part, tool or chemical", Required: true, Type: "non-empty-uuid"},
				},
			},
		},
	})
}
