package registry

func init() {
	Register(&Domain{
		Name:        "condition",
		Description: "Manage asset condition assessments",
		Actions: []Action{
			{
				Name:        "assess",
				Description: "Create a condition assessment for an asset",
				ToolName:    "UteamupConditionAssess",
				Flags: []FlagDef{
					{Name: "asset-id", Description: "Asset ID", Required: true, Type: "int"},
					{Name: "overall-grade", Description: "Overall condition grade", Required: true, Type: "int"},
					{Name: "structural-grade", Description: "Structural condition grade", Type: "int"},
					{Name: "performance-grade", Description: "Performance condition grade", Type: "int"},
					{Name: "safety-grade", Description: "Safety condition grade", Type: "int"},
					{Name: "compliance-grade", Description: "Compliance condition grade", Type: "int"},
					{Name: "notes", Description: "Assessment notes", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get the latest condition assessment for an asset",
				ToolName:    "UteamupConditionGet",
				Flags: []FlagDef{
					{Name: "asset-id", Description: "Asset ID", Required: true, Type: "int"},
				},
			},
			{
				Name:        "history",
				Description: "Get condition assessment history for an asset",
				ToolName:    "UteamupConditionHistory",
				Flags: append(paginationFlags(),
					FlagDef{Name: "asset-id", Description: "Asset ID", Required: true, Type: "int"},
				),
			},
			{
				Name:        "heat-map",
				Description: "Get the condition heat map",
				ToolName:    "UteamupConditionHeatMap",
				Flags: []FlagDef{
					{Name: "location-id", Description: "Filter by location ID", Type: "int"},
					{Name: "max-grade", Description: "Maximum grade to include", Type: "int"},
				},
			},
			{
				Name:        "overdue",
				Description: "Get overdue condition assessments",
				ToolName:    "UteamupConditionOverdue",
				Flags:       paginationFlags(),
			},
		},
	})
}
