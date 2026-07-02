package registry

func init() {
	Register(&Domain{
		Name:        "plan-analytics",
		Aliases:     []string{"plananalytics"},
		Description: "Per-plan revenue, discount, and grandfathering analytics",
		APIPath:     "/api/plananalytics",
		Actions: []Action{
			{
				Name:        "summary",
				Description: "Plan-level summary: subscribers, grandfathered count, revenue, discount load",
				ToolName:    "UteamupPlanAnalyticsSummary",
				RESTPath:    "summary",
				Flags: []FlagDef{
					{Name: "from-date", Description: "Window start (ISO 8601 datetime; backend defaults to 90 days back)", Type: "string"},
				},
			},
			{
				Name:        "insights",
				Description: "AI narrative over the analytics summary (503 when the AI service is unavailable)",
				ToolName:    "UteamupPlanAnalyticsInsights",
				RESTPath:    "insights",
				Flags: []FlagDef{
					{Name: "from-date", Description: "Window start (ISO 8601 datetime; backend defaults to 90 days back)", Type: "string"},
					{Name: "provider", Description: "AI provider (backend defaults to openai)", Type: "string"},
				},
			},
		},
	})
}
