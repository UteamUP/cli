package registry

func init() {
	Register(&Domain{
		Name:        "labour-ai-monitor",
		Aliases:     []string{"job-monitor"},
		Description: "Manage private contractor AI job monitors and credit budgets",
		APIPath:     "/api/labour-marketplace/ai/job-monitors",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List your private AI job monitors",
				ToolName:    "UteamupLabourMarketplaceAiJobMonitorsList",
			},
			{
				Name:        "cost",
				Description: "Preview the live AI-credit cost and your remaining balance",
				ToolName:    "UteamupLabourMarketplaceAiJobMonitorCost",
				RESTPath:    "cost",
				HTTPMethod:  "GET",
			},
			{
				Name:        "create",
				Description: "Create a private monitor for an owned contractor profile",
				ToolName:    "UteamupLabourMarketplaceAiJobMonitorCreate",
				Flags:       monitorFlags(),
			},
			{
				Name:        "update",
				Description: "Update one owned monitor by GUID",
				ToolName:    "UteamupLabourMarketplaceAiJobMonitorUpdate",
				RESTPath:    "{monitorGuid}",
				Args: []ArgDef{
					{Name: "monitorGuid", Description: "Monitor GUID", Required: true, Type: "uuid"},
				},
				Flags: monitorFlags(),
			},
			{
				Name:        "delete",
				Description: "Delete one owned monitor and its private digest history",
				ToolName:    "UteamupLabourMarketplaceAiJobMonitorDelete",
				RESTPath:    "{monitorGuid}",
				Args: []ArgDef{
					{Name: "monitorGuid", Description: "Monitor GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "run",
				Description: "Run a monitor now; may consume the previewed credits when new matches exist",
				ToolName:    "UteamupLabourMarketplaceAiJobMonitorRun",
				RESTPath:    "{monitorGuid}/run",
				HTTPMethod:  "POST",
				Args: []ArgDef{
					{Name: "monitorGuid", Description: "Monitor GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}

func monitorFlags() []FlagDef {
	return []FlagDef{
		{Name: "provider-party-guid", Description: "Owned contractor party profile GUID", Required: true, Type: "string"},
		{Name: "name", Short: "n", Description: "Private monitor name", Required: true, Type: "string"},
		{Name: "search-term", Description: "Skills or keywords", Type: "string"},
		{Name: "industry", Description: "Industry filter", Type: "string"},
		{Name: "location", Description: "Location or service-area filter", Type: "string"},
		{Name: "maximum-budget", Description: "Maximum job budget", Type: "float"},
		{Name: "frequency-hours", Description: "Run frequency in hours (1-168)", Type: "int", Default: 24},
		{Name: "max-credits-per-run", Description: "Maximum AI credits per run", Type: "int", Default: 2},
		{Name: "monthly-credit-budget", Description: "Maximum AI credits per month", Type: "int", Default: 20},
		{Name: "is-active", Description: "Enable scheduled monitoring", Type: "bool", Default: true},
	}
}
