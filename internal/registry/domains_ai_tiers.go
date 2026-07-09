package registry

func init() {
	Register(&Domain{
		Name:        "workforce-ai",
		Aliases:     []string{"wfai"},
		Description: "Run workforce AI assist operations",
		APIPath:     "/api/workforceai",
		Actions: []Action{
			{
				Name:        "daily-brief",
				Description: "Generate the current user's My Day AI daily brief",
				ToolName:    "UteamupWorkforceAiDailyBrief",
				RESTPath:    "daily-brief",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "date", Description: "Brief date (UTC ISO-8601 or YYYY-MM-DD). Defaults to today.", Type: "string"},
					{Name: "regenerate", Description: "Force a fresh brief and re-charge credits", Type: "bool"},
					{Name: "currentLatitude", Description: "Optional current GPS latitude for route ordering", Type: "float"},
					{Name: "currentLongitude", Description: "Optional current GPS longitude for route ordering", Type: "float"},
				},
			},
		},
	})

	Register(&Domain{
		Name:        "work-permit-ai",
		Aliases:     []string{"wpai"},
		Description: "Run review-only AI assist operations for work permits",
		APIPath:     "/api/workpermit",
		Actions: []Action{
			{
				Name:        "prefill",
				Description: "Suggest permit hazards, PPE, isolation steps, and rescue notes",
				ToolName:    "UteamupWorkPermitAiPrefill",
				RESTPath:    "by-guid/{work-permit-guid}/ai-prefill",
				HTTPMethod:  "POST",
				Args: []ArgDef{
					{Name: "work-permit-guid", Description: "Work permit external Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "workorder-guid", Description: "Optional related workorder Guid", Type: "string"},
					{Name: "asset-guid", Description: "Optional related asset Guid", Type: "string"},
					{Name: "context", Description: "Optional extra field context for the AI prompt", Type: "string"},
				},
			},
		},
	})
}
