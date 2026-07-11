package registry

func init() {
	Register(&Domain{
		Name:        "labour-contractor-analytics",
		Aliases:     []string{"contractor-analytics"},
		Description: "View the authenticated user's private contractor portfolio and personal dispatch analytics",
		APIPath:     "/api/labour-marketplace",
		Actions: []Action{
			{
				Name:        "me",
				Description: "Show current-user contractor funnel, personal dispatches, and per-currency approved earnings",
				ToolName:    "UteamupLabourContractorAnalyticsMe",
				RESTPath:    "analytics/me",
				HTTPMethod:  "GET",
			},
		},
	})
}
