package registry

func init() {
	Register(&Domain{
		Name:        "asset-failure",
		Description: "Manage asset failure records",
		Actions: []Action{
			{
				Name:        "by-severity",
				Description: "List tenant asset failures by severity",
				ToolName:    "UteamupAssetfailureGetBySeverity",
				HTTPMethod:  "GET",
				RESTPath:    "by-severity",
				Flags: []FlagDef{
					{Name: "severity", Description: "Critical, High, Medium, or Low", Type: "string", Required: true},
				},
			},
			{
				Name:        "open",
				Description: "List unresolved tenant asset failures",
				ToolName:    "UteamupAssetfailureGetOpen",
				HTTPMethod:  "GET",
				RESTPath:    "open",
			},
			{
				Name:        "by-asset",
				Description: "List failure records for an asset GUID",
				ToolName:    "UteamupAssetfailureGetByAsset",
				Args: []ArgDef{{
					Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "uuid",
				}},
			},
			{
				Name:        "statistics",
				Description: "Get failure statistics for an asset GUID",
				ToolName:    "UteamupAssetfailureGetStatistics",
				Args: []ArgDef{{
					Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "uuid",
				}},
			},
			{
				Name:        "get",
				Description: "Get a failure record by GUID",
				ToolName:    "UteamupAssetfailureGet",
				Args: []ArgDef{{
					Name: "failureGuid", Description: "Public failure GUID", Required: true, Type: "uuid",
				}},
			},
			{
				Name:        "evidence-pack",
				Description: "Get the facts-only incident evidence pack for the hours around a failure (no AI, nothing stored)",
				ToolName:    "UteamupAssetFailureEvidencePack",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{failureGuid}/evidence-pack",
				Flags: []FlagDef{
					{
						Name: "guid", BodyName: "failureGuid",
						Description: "Public failure GUID", Required: true, Type: "non-empty-uuid",
					},
					{
						Name: "hours-before", QueryName: "hoursBefore",
						Description: "Hours before the failure time, 1-168 (default 24)", Type: "int",
					},
					{
						Name: "hours-after", QueryName: "hoursAfter",
						Description: "Hours after the failure time, 1-168 (default 6)", Type: "int",
					},
				},
			},
			{Name: "create", Description: "Create a failure record", ToolName: "UteamupAssetfailureCreate", Flags: []FlagDef{jsonFlag()}},
			{
				Name:        "update",
				Description: "Update a failure record by GUID",
				ToolName:    "UteamupAssetfailureUpdate",
				Args: []ArgDef{{
					Name: "failureGuid", Description: "Public failure GUID", Required: true, Type: "uuid",
				}},
				Flags: []FlagDef{jsonFlag()},
			},
			{Name: "resolve", Description: "Resolve a failure record", ToolName: "UteamupAssetfailureResolve", Flags: []FlagDef{jsonFlag()}},
			{
				Name:        "delete",
				Description: "Delete a failure record by GUID",
				ToolName:    "UteamupAssetfailureDelete",
				Args: []ArgDef{{
					Name: "failureGuid", Description: "Public failure GUID", Required: true, Type: "uuid",
				}},
			},
			{
				Name:        "classify",
				Description: "Classify a failure record by GUID",
				ToolName:    "UteamupAssetfailureClassify",
				Args: []ArgDef{{
					Name: "failureGuid", Description: "Public failure GUID", Required: true, Type: "uuid",
				}},
				Flags: []FlagDef{{
					Name: "failure-classification", BodyName: "failureClassification",
					Description: "EN 13306 failure classification", Required: true, Type: "string",
				}},
			},
		},
	})
}
