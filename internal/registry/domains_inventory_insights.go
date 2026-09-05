package registry

func init() {
	Register(&Domain{
		Name:        "insights",
		Aliases:     []string{"inventory-insights"},
		Description: "Inventory Insights: GUID-first work-order history per asset, part, tool or chemical, and the UPMate root cause analysis draft",
		APIPath:     "/api/inventory/insights",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Work-order totals, monthly timeline, status mix, recent work orders and (assets) linked root cause analyses",
				ToolName:    "UteamupAssetInsightsGet",
				HTTPMethod:  "GET",
				RESTPath:    "{entityType}/by-guid/{entityGuid}",
				Args: []ArgDef{
					{
						Name: "entityType", Description: "asset, part, tool or chemical", Required: true, Type: "string",
						AllowedValues: []string{"asset", "part", "tool", "chemical"},
					},
					{Name: "entityGuid", Description: "Public GUID of the inventory item", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "months-back", Description: "Months of history to bucket (1-36, default 12)", Default: 12, Type: "int"},
				},
			},
			{
				Name:        "rca-draft",
				Description: "Ask UPMate for a 5 Whys root cause analysis draft grounded in the asset's work orders (charges AI credits, saves nothing)",
				ToolName:    "UteamupAssetRcaDraft",
				HTTPMethod:  "POST",
				RESTPath:    "asset/by-guid/{assetGuid}/root-cause-analysis/draft",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "months-back", BodyName: "monthsBack", Description: "Months of evidence to analyze (1-36, default 12)", Default: 12, Type: "int"},
					{Name: "focus", BodyName: "focus", Description: "Optional problem focus (max 500 chars)", Type: "string"},
				},
			},
			{
				Name:        "rca-create",
				Description: "Persist a reviewed UPMate draft as a root cause analysis linked to the asset (JSON body: draft + workorderGuids)",
				ToolName:    "UteamupAssetRcaCreateFromDraft",
				HTTPMethod:  "POST",
				RESTPath:    "asset/by-guid/{assetGuid}/root-cause-analysis",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "uuid"}},
				Flags:       []FlagDef{jsonFlag()},
			},
		},
	})
}
