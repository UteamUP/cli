package registry

func init() {
	Register(&Domain{
		Name:        "criticality",
		Description: "Manage asset criticality assessments",
		Actions: []Action{
			{
				Name:        "assess",
				Description: "Create a criticality assessment for an asset",
				ToolName:    "UteamupCriticalityAssess",
				Flags: []FlagDef{
					{Name: "asset-guid", Description: "Public asset GUID", Required: true, Type: "string"},
					{Name: "consequence-score", Description: "Consequence score", Required: true, Type: "int"},
					{Name: "probability-score", Description: "Probability score", Required: true, Type: "int"},
					{Name: "notes", Description: "Assessment notes", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get the latest criticality assessment for an asset",
				ToolName:    "UteamupCriticalityGet",
				Flags: []FlagDef{
					{Name: "asset-guid", Description: "Public asset GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "history",
				Description: "Get criticality assessment history for an asset",
				ToolName:    "UteamupCriticalityHistory",
				Flags: append(paginationFlags(),
					FlagDef{Name: "asset-guid", Description: "Public asset GUID", Required: true, Type: "string"},
				),
			},
			{
				Name:        "matrix",
				Description: "Get the criticality matrix",
				ToolName:    "UteamupCriticalityMatrix",
				Flags: []FlagDef{
					{Name: "location-guid", Description: "Filter by public location GUID", Type: "string"},
					{Name: "asset-type-guid", Description: "Filter by public asset type GUID", Type: "string"},
				},
			},
		},
	})
}
