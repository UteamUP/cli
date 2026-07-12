package registry

func init() {
	Register(&Domain{
		Name:        "labour-marketplace-workspace",
		Aliases:     []string{"labour-workspace", "offers-contracts"},
		Description: "View the authenticated user's private buyer jobs, provider applications, agreements, and dispatches",
		APIPath:     "/api/labour-marketplace",
		Actions: []Action{
			{
				Name:        "me",
				Description: "Show the current user's combined labour marketplace workspace",
				ToolName:    "UteamupLabourMarketplaceWorkspaceMe",
				RESTPath:    "workspace/me",
				HTTPMethod:  "GET",
			},
		},
	})
}
