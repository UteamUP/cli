package registry

func init() {
	Register(&Domain{
		Name:        "condition",
		Description: "Manage asset condition assessments",
		// Routes mirror AssetConditionAssessmentController ([Route("api/[controller]")]). Without
		// APIPath the REST fallback derived /api/condition — no such controller, so every verb 404'd.
		APIPath: "/api/assetconditionassessment",
		Actions: []Action{
			{
				Name:        "assess",
				Description: "Create a condition assessment for an asset",
				ToolName:    "UteamupConditionAssess",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "asset-guid", Description: "Public asset GUID", Required: true, Type: "string"},
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
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{assetGuid}/latest",
				Flags: []FlagDef{
					{Name: "asset-guid", Description: "Public asset GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "history",
				Description: "Get condition assessment history for an asset",
				ToolName:    "UteamupConditionHistory",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{assetGuid}/history",
				Flags: append(paginationFlags(),
					FlagDef{Name: "asset-guid", Description: "Public asset GUID", Required: true, Type: "string"},
				),
			},
			{
				Name:        "heat-map",
				Description: "Get the condition heat map",
				ToolName:    "UteamupConditionHeatMap",
				HTTPMethod:  "GET",
				RESTPath:    "heat-map",
				Flags: []FlagDef{
					{Name: "location-guid", Description: "Filter by public location GUID", Type: "string"},
					{Name: "max-grade", Description: "Maximum grade to include", Type: "int"},
				},
			},
			{
				Name:        "overdue",
				Description: "Get overdue condition assessments",
				ToolName:    "UteamupConditionOverdue",
				HTTPMethod:  "GET",
				RESTPath:    "due",
				Flags:       paginationFlags(),
			},
		},
	})
}
