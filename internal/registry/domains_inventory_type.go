package registry

// Inventory type catalogue (UTE-1075). A type defines the fields; each individual
// part, tool or chemical owns its own values.
//
// The kind is a route segment on the backend (/api/inventory-type/{kind}), so it is a
// positional argument constrained to the three inventory kinds. The permission is
// derived from the kind on the server: {Part|Tool|Chemical}.View for reads and
// .Update for writes, so a part command never answers for a tool or a chemical.
//
// The routes are the ones the API actually serves:
//
//	GET    /api/inventory-type/{kind}
//	GET    /api/inventory-type/{kind}/{typeGuid}
//	POST   /api/inventory-type/{kind}
//	POST   /api/inventory-type/{kind}/catalogue
//	PUT    /api/inventory-type/{kind}/{typeGuid}
//	POST   /api/inventory-type/{kind}/{typeGuid}/fields
//	PUT    /api/inventory-type/{kind}/{typeGuid}/fields/{fieldGuid}
//	GET    /api/inventory-type/{kind}/items/{itemGuid}
//	PATCH  /api/inventory-type/{kind}/items/{itemGuid}
func init() {
	Register(&Domain{
		Name:        "inventory-type",
		Aliases:     []string{"inventorytype", "invtype"},
		Description: "Manage the part, tool and chemical type catalogue and the type selection plus per-item values of one inventory item",
		APIPath:     "/api/inventory-type",
		Actions: []Action{
			{
				Name:            "list",
				Description:     "List the type definitions of one inventory kind (part, tool or chemical)",
				ToolName:        "UteamupInventoryTypeList",
				HTTPMethod:      "GET",
				RESTPath:        "{kind}",
				RejectExtraArgs: true,
				Args:            []ArgDef{inventoryTypeKindArg()},
				Flags: []FlagDef{
					{Name: "include-inactive", Description: "Include archived types (default: false)", Default: false, Type: "bool"},
					{Name: "page", Description: "Page number, 1-based (default: 1)", Default: 1, Type: "int"},
					{Name: "page-size", Description: "Items per page, 1-200 (default: 50)", Default: 50, Type: "int"},
				},
			},
			{
				Name:            "get",
				Description:     "Get one type definition with its field definitions",
				ToolName:        "UteamupInventoryTypeGet",
				HTTPMethod:      "GET",
				RESTPath:        "{kind}/{typeGuid}",
				RejectExtraArgs: true,
				Args: []ArgDef{
					inventoryTypeKindArg(),
					{Name: "typeGuid", Description: "Type GUID", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:            "create",
				Description:     "Create a type definition for one inventory kind",
				ToolName:        "UteamupInventoryTypeCreate",
				HTTPMethod:      "POST",
				RESTPath:        "{kind}",
				RejectExtraArgs: true,
				Args:            []ArgDef{inventoryTypeKindArg()},
				Flags: []FlagDef{
					{Name: "name", Description: "Type name (max 200 characters)", Required: true, Type: "string"},
					{Name: "description", Description: "Type description (max 2000 characters, optional)", Type: "string"},
					{Name: "icon-name", Description: "Icon name (max 100 characters, optional)", Type: "string"},
					{Name: "is-active", Description: "Whether the type is selectable (default: true)", Default: true, Type: "bool"},
					{Name: "display-order", Description: "Sort order (default: 0)", Default: 0, Type: "int"},
				},
			},
			{
				Name:            "update",
				Description:     "Update a type definition, or archive it with --is-active=false",
				ToolName:        "UteamupInventoryTypeUpdate",
				HTTPMethod:      "PUT",
				RESTPath:        "{kind}/{typeGuid}",
				RejectExtraArgs: true,
				Args: []ArgDef{
					inventoryTypeKindArg(),
					{Name: "typeGuid", Description: "Type GUID", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{
					{Name: "name", Description: "Type name (max 200 characters)", Required: true, Type: "string"},
					{Name: "description", Description: "Type description (max 2000 characters, optional)", Type: "string"},
					{Name: "icon-name", Description: "Icon name (max 100 characters, optional)", Type: "string"},
					{Name: "is-active", Description: "Whether the type is selectable; false archives it", Default: true, Type: "bool"},
					{Name: "display-order", Description: "Sort order (default: 0)", Default: 0, Type: "int"},
				},
			},
			{
				Name:            "catalogue-apply",
				Description:     "Add the shipped type catalogue of one kind to the tenant; types the tenant already has are skipped",
				ToolName:        "UteamupInventoryTypeCatalogueApply",
				HTTPMethod:      "POST",
				RESTPath:        "{kind}/catalogue",
				RejectExtraArgs: true,
				Args:            []ArgDef{inventoryTypeKindArg()},
			},
			{
				Name:            "field-create",
				Description:     "Add a field definition to a type (the field key and data type are the field's immutable identity)",
				ToolName:        "UteamupInventoryTypeFieldCreate",
				HTTPMethod:      "POST",
				RESTPath:        "{kind}/{typeGuid}/fields",
				RejectExtraArgs: true,
				Args: []ArgDef{
					inventoryTypeKindArg(),
					{Name: "typeGuid", Description: "Type GUID", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{
					{Name: "field-key", Description: "Canonical field key matching ^[a-z][a-z0-9_]{0,99}$", Required: true, Type: "string"},
					{Name: "name", Description: "Field name (max 200 characters)", Required: true, Type: "string"},
					{Name: "description", Description: "Field description (max 1000 characters, optional)", Type: "string"},
					{Name: "data-type", Description: "Value type: string, integer, float, boolean, date or array", Required: true, Type: "string",
						AllowedValues: []string{"string", "integer", "float", "boolean", "date", "array"}},
					{Name: "unit", Description: "Unit of measure (max 50 characters, optional)", Type: "string"},
					{Name: "group-name", Description: "Field group (max 100 characters, optional)", Type: "string"},
					{Name: "is-required", Description: "Whether the item must carry a value for this field", Default: false, Type: "bool"},
					{Name: "is-active", Description: "Whether the field is selectable (default: true)", Default: true, Type: "bool"},
					{Name: "display-order", Description: "Sort order (default: 0)", Default: 0, Type: "int"},
				},
			},
			{
				Name:            "field-update",
				Description:     "Update a field definition, or archive it with --is-active=false; the field key and data type never change",
				ToolName:        "UteamupInventoryTypeFieldUpdate",
				HTTPMethod:      "PUT",
				RESTPath:        "{kind}/{typeGuid}/fields/{fieldGuid}",
				RejectExtraArgs: true,
				Args: []ArgDef{
					inventoryTypeKindArg(),
					{Name: "typeGuid", Description: "Type GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "fieldGuid", Description: "Field GUID", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{
					{Name: "field-key", Description: "Canonical field key matching ^[a-z][a-z0-9_]{0,99}$", Required: true, Type: "string"},
					{Name: "name", Description: "Field name (max 200 characters)", Required: true, Type: "string"},
					{Name: "description", Description: "Field description (max 1000 characters, optional)", Type: "string"},
					{Name: "data-type", Description: "Value type: string, integer, float, boolean, date or array", Required: true, Type: "string",
						AllowedValues: []string{"string", "integer", "float", "boolean", "date", "array"}},
					{Name: "unit", Description: "Unit of measure (max 50 characters, optional)", Type: "string"},
					{Name: "group-name", Description: "Field group (max 100 characters, optional)", Type: "string"},
					{Name: "is-required", Description: "Whether the item must carry a value for this field", Default: false, Type: "bool"},
					{Name: "is-active", Description: "Whether the field is selectable; false archives it", Default: true, Type: "bool"},
					{Name: "display-order", Description: "Sort order (default: 0)", Default: 0, Type: "int"},
				},
			},
			{
				Name:            "item-get",
				Description:     "Get the type selection and the per-item values of one part, tool or chemical",
				ToolName:        "UteamupInventoryTypeItemGet",
				HTTPMethod:      "GET",
				RESTPath:        "{kind}/items/{itemGuid}",
				RejectExtraArgs: true,
				Args: []ArgDef{
					inventoryTypeKindArg(),
					{Name: "itemGuid", Description: "Inventory item GUID", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:            "item-save",
				Description:     "Set the type selection and the per-item values of one part, tool or chemical (sparse patch: omitted fields keep their value, an explicit null removes it)",
				ToolName:        "UteamupInventoryTypeItemSave",
				HTTPMethod:      "PATCH",
				RESTPath:        "{kind}/items/{itemGuid}",
				RejectExtraArgs: true,
				Args: []ArgDef{
					inventoryTypeKindArg(),
					{Name: "itemGuid", Description: "Inventory item GUID", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{
					{Name: "file", Short: "f", Description: "Path to a JSON file with the selection patch: {\"typeGuids\":[\"guid\",...],\"primaryTypeGuid\":\"guid\",\"values\":[{\"fieldGuid\":\"guid\",\"value\":<native value>}]}. Omit typeGuids to preserve the assignment; [] clears it.", Required: true, Type: "string", RootJSONObjectFile: true},
				},
			},
		},
	})
}

// inventoryTypeKindArg is the route segment that selects which inventory kind the
// command addresses. It is constrained to the three kinds the API serves so a
// caller can never reshape the path.
func inventoryTypeKindArg() ArgDef {
	return ArgDef{
		Name: "kind", Description: "Inventory kind: part, tool or chemical", Required: true, Type: "string",
		AllowedValues: []string{"part", "tool", "chemical"},
	}
}
