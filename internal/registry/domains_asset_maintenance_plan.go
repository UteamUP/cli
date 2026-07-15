package registry

// Asset maintenance plans use their existing external GUIDs at every public
// boundary. The routes mirror AssetMaintenancePlanController exactly; the
// generic CRUD registry is intentionally not used because it assumes an
// integer `id` argument and a different API base path.
func init() {
	planFlags := []FlagDef{
		{Name: "name", Description: "Maintenance plan name", Required: true, Type: "string"},
		{Name: "description", Description: "Optional maintenance plan description", Type: "string"},
		{Name: "is-active", BodyName: "isActive", Description: "Whether the maintenance plan is active", Default: true, Type: "bool"},
	}

	itemFlags := []FlagDef{
		{Name: "name", Description: "Maintenance plan item name", Required: true, Type: "string"},
		{Name: "trigger-type", BodyName: "triggerType", Description: "Trigger type: 0=calendar, 1=meter, 2=inspection result, 3=IoT alert", Default: 0, Type: "int"},
		{Name: "calendar-interval-days", BodyName: "calendarIntervalDays", Description: "Calendar interval in days (0-3650)", Type: "int"},
		{Name: "meter-interval-value", BodyName: "meterIntervalValue", Description: "Meter usage interval", Type: "float"},
		{Name: "required-chemical-items-json", BodyName: "requiredChemicalItemsJson", Description: "Required chemicals as JSON text", Type: "string"},
		{Name: "required-parts-json", BodyName: "requiredPartsJson", Description: "Required parts as JSON text", Type: "string"},
		{Name: "required-tools-json", BodyName: "requiredToolsJson", Description: "Required tools as JSON text", Type: "string"},
		{Name: "workorder-template-external-guid", BodyName: "workorderTemplateExternalGuid", Description: "Optional workorder template external GUID", Type: "string"},
		{Name: "required-for-warranty", BodyName: "isRequiredForWarranty", Description: "Whether completion is required for warranty", Type: "bool"},
		{Name: "required-for-certification", BodyName: "isRequiredForCertification", Description: "Whether completion is required for certification", Type: "bool"},
	}

	Register(&Domain{
		Name:        "asset-maintenance-plan",
		Aliases:     []string{"amp"},
		Description: "Manage asset maintenance plans and their items by external GUID",
		APIPath:     "/api/v1/maintenanceplans",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List maintenance plans, optionally filtered by asset external GUID",
				ToolName:    "UteamupAssetMaintenancePlanList",
				Flags: []FlagDef{
					{Name: "asset-external-guid", BodyName: "assetExternalGuid", Description: "Optional asset external GUID filter", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get a maintenance plan by external GUID",
				ToolName:    "UteamupAssetMaintenancePlanGet",
				RESTPath:    "{planExternalGuid}",
				Args: []ArgDef{
					{Name: "planExternalGuid", Description: "Maintenance plan external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create a maintenance plan for an asset external GUID",
				ToolName:    "UteamupAssetMaintenancePlanCreate",
				HTTPMethod:  "POST",
				RESTPath:    "asset/{assetExternalGuid}",
				Args: []ArgDef{
					{Name: "assetExternalGuid", Description: "Asset external GUID", Required: true, Type: "uuid"},
				},
				Flags: planFlags,
			},
			{
				Name:        "update",
				Description: "Update a maintenance plan by external GUID",
				ToolName:    "UteamupAssetMaintenancePlanUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{planExternalGuid}",
				Args: []ArgDef{
					{Name: "planExternalGuid", Description: "Maintenance plan external GUID", Required: true, Type: "uuid"},
				},
				Flags: planFlags,
			},
			{
				Name:        "delete",
				Description: "Delete a maintenance plan by external GUID",
				ToolName:    "UteamupAssetMaintenancePlanDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "{planExternalGuid}",
				Args: []ArgDef{
					{Name: "planExternalGuid", Description: "Maintenance plan external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "items",
				Description: "List the items in a maintenance plan",
				ToolName:    "UteamupAssetMaintenancePlanGetItems",
				HTTPMethod:  "GET",
				RESTPath:    "{planExternalGuid}/items",
				Args: []ArgDef{
					{Name: "planExternalGuid", Description: "Maintenance plan external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "item-add",
				Description: "Add an item to a maintenance plan",
				ToolName:    "UteamupAssetMaintenancePlanAddItem",
				HTTPMethod:  "POST",
				RESTPath:    "{planExternalGuid}/items",
				Args: []ArgDef{
					{Name: "planExternalGuid", Description: "Maintenance plan external GUID", Required: true, Type: "uuid"},
				},
				Flags: itemFlags,
			},
			{
				Name:        "item-update",
				Description: "Update a maintenance plan item by external GUID",
				ToolName:    "UteamupAssetMaintenancePlanUpdateItem",
				HTTPMethod:  "PUT",
				RESTPath:    "items/{itemExternalGuid}",
				Args: []ArgDef{
					{Name: "itemExternalGuid", Description: "Maintenance plan item external GUID", Required: true, Type: "uuid"},
				},
				Flags: itemFlags,
			},
			{
				Name:        "item-delete",
				Description: "Delete a maintenance plan item by external GUID",
				ToolName:    "UteamupAssetMaintenancePlanDeleteItem",
				HTTPMethod:  "DELETE",
				RESTPath:    "items/{itemExternalGuid}",
				Args: []ArgDef{
					{Name: "itemExternalGuid", Description: "Maintenance plan item external GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
