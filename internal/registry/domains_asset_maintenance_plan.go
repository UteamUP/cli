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
		{Name: "template-external-guid", BodyName: "templateExternalGuid", Description: "Optional reusable template version external GUID", Type: "string"},
		{Name: "effective-date", BodyName: "effectiveDate", Description: "Mid-life calendar anchor in ISO-8601 UTC", Type: "string"},
		{Name: "baseline-meter-value", BodyName: "baselineMeterValue", Description: "Known meter value at onboarding", Type: "float"},
		{Name: "baseline-meter-date", BodyName: "baselineMeterDate", Description: "Meter baseline timestamp in ISO-8601 UTC", Type: "string"},
		{Name: "baseline-meter-attribute-external-guid", BodyName: "baselineMeterAttributeExternalGuid", Description: "Meter attribute external GUID for the onboarding baseline", Type: "string"},
		{Name: "consolidation-window-days", BodyName: "consolidationWindowDays", Description: "Compatible-work consolidation window in days", Type: "int"},
		{Name: "due-trigger-policy", BodyName: "dueTriggerPolicy", Description: "Due policy: 0=earliest valid trigger wins", Default: 0, Type: "int"},
	}

	itemFlags := []FlagDef{
		{Name: "name", Description: "Maintenance plan item name", Required: true, Type: "string"},
		{Name: "trigger-type", BodyName: "triggerType", Description: "Trigger type: 0=calendar, 1=meter, 2=inspection result, 3=IoT alert", Default: 0, Type: "int"},
		{Name: "calendar-interval-days", BodyName: "calendarIntervalDays", Description: "Calendar interval in days (1-3650)", Type: "int"},
		{Name: "meter-interval-value", BodyName: "meterIntervalValue", Description: "Meter usage interval", Type: "float"},
		{Name: "meter-attribute-definition-external-guid", BodyName: "meterAttributeDefinitionExternalGuid", Description: "Meter attribute definition external GUID", Type: "string"},
		{Name: "required-chemical-items-json", BodyName: "requiredChemicalItemsJson", Description: "Required chemicals as JSON text", Type: "string"},
		{Name: "required-parts-json", BodyName: "requiredPartsJson", Description: "Required parts as JSON text", Type: "string"},
		{Name: "required-tools-json", BodyName: "requiredToolsJson", Description: "Required tools as JSON text", Type: "string"},
		{Name: "workorder-template-external-guid", BodyName: "workorderTemplateExternalGuid", Description: "Optional workorder template external GUID", Type: "string"},
		{Name: "required-for-warranty", BodyName: "isRequiredForWarranty", Description: "Whether completion is required for warranty", Type: "bool"},
		{Name: "required-for-certification", BodyName: "isRequiredForCertification", Description: "Whether completion is required for certification", Type: "bool"},
	}

	templateFlags := []FlagDef{
		{Name: "name", Description: "Reusable maintenance template name", Required: true, Type: "string"},
		{Name: "description", Description: "Optional template description", Type: "string"},
		{Name: "consolidation-window-days", BodyName: "consolidationWindowDays", Description: "Compatible-work consolidation window in days", Type: "int"},
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
			{
				Name:        "template-list",
				Description: "List current or historical reusable maintenance template versions",
				ToolName:    "UteamupAssetMaintenancePlanTemplateList",
				HTTPMethod:  "GET",
				RESTPath:    "templates",
				Flags: []FlagDef{
					{Name: "include-history", BodyName: "includeHistory", Description: "Include superseded template versions", Type: "bool"},
				},
			},
			{
				Name:        "template-get",
				Description: "Get one reusable maintenance template version by external GUID",
				ToolName:    "UteamupAssetMaintenancePlanTemplateGet",
				RESTPath:    "templates/{templateExternalGuid}",
				Args: []ArgDef{
					{Name: "templateExternalGuid", Description: "Template version external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "template-create",
				Description: "Create a reusable template; use --from-json to include item definitions",
				ToolName:    "UteamupAssetMaintenancePlanTemplateCreate",
				HTTPMethod:  "POST",
				RESTPath:    "templates",
				Flags:       templateFlags,
			},
			{
				Name:        "template-version",
				Description: "Create an immutable template version; use --from-json to include item definitions",
				ToolName:    "UteamupAssetMaintenancePlanTemplateCreateVersion",
				HTTPMethod:  "POST",
				RESTPath:    "templates/{templateExternalGuid}/versions",
				Args: []ArgDef{
					{Name: "templateExternalGuid", Description: "Source template version external GUID", Required: true, Type: "uuid"},
				},
				Flags: templateFlags,
			},
			{
				Name:        "due-projection",
				Description: "Preview deterministic due evidence without writing",
				ToolName:    "UteamupAssetMaintenancePlanDueProjection",
				HTTPMethod:  "GET",
				RESTPath:    "{planExternalGuid}/due-projection",
				Args: []ArgDef{
					{Name: "planExternalGuid", Description: "Maintenance plan external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "as-of", BodyName: "asOf", Description: "Optional UTC evidence cutoff", Type: "string"},
				},
			},
		},
	})
}
