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
		},
	})
}
