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
					{Name: "asset-id", Description: "Asset ID", Required: true, Type: "int"},
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
					{Name: "asset-id", Description: "Asset ID", Required: true, Type: "int"},
				},
			},
			{
				Name:        "history",
				Description: "Get criticality assessment history for an asset",
				ToolName:    "UteamupCriticalityHistory",
				Flags: append(paginationFlags(),
					FlagDef{Name: "asset-id", Description: "Asset ID", Required: true, Type: "int"},
				),
			},
			{
				Name:        "matrix",
				Description: "Get the criticality matrix",
				ToolName:    "UteamupCriticalityMatrix",
				Flags: []FlagDef{
					{Name: "location-id", Description: "Filter by location ID", Type: "int"},
					{Name: "asset-type-id", Description: "Filter by asset type ID", Type: "int"},
				},
			},
		},
	})
}
