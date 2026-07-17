package registry

func init() {
	Register(&Domain{
		Name:        "reliability",
		Aliases:     []string{"rel"},
		Description: "Read GUID-first reliability intelligence",
		APIPath:     "/api/analytics/reliability",
		Actions: []Action{
			{
				Name:        "risk",
				Description: "Rank evidence-backed asset reliability risks and bad actors",
				ToolName:    "UteamupReliabilityRiskGet",
				HTTPMethod:  "GET",
				RESTPath:    "risks",
				Flags: []FlagDef{
					{Name: "asset-guid", Description: "Optional public asset GUID", Type: "string"},
					{Name: "from-utc", Description: "Optional UTC evidence-window start", Type: "string"},
					{Name: "to-utc", Description: "Optional UTC evidence-window end", Type: "string"},
					{Name: "limit", Description: "Maximum bad actors to return (1-100)", Default: 20, Type: "int"},
				},
			},
			{
				Name:        "strategy",
				Description: "Compare review-only reliability strategy alternatives for one asset",
				ToolName:    "UteamupReliabilityStrategyPropose",
				HTTPMethod:  "POST",
				RESTPath:    "strategies/propose",
				Flags: []FlagDef{
					{Name: "asset-guid", BodyName: "assetGuid", Description: "Public asset GUID", Required: true, Type: "string"},
					{Name: "objective", BodyName: "objectiveKey", Description: "availability, downtime, cost, or safety", Default: "availability", Type: "string"},
					{Name: "from-utc", BodyName: "fromUtc", Description: "Optional UTC evidence-window start", Type: "string"},
					{Name: "to-utc", BodyName: "toUtc", Description: "Optional UTC evidence-window end", Type: "string"},
				},
			},
		},
	})
}
