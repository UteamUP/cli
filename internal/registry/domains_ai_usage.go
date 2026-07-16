package registry

// AI request usage summary for the caller's tenant.
// Requires AIUsage.Read permission on the backend.

func init() {
	Register(&Domain{
		Name:        "ai-usage",
		Aliases:     []string{"aiu"},
		Description: "View AI request usage (tokens, estimated cost) for the tenant",
		Actions: []Action{
			{
				Name:         "data-readiness",
				Description:  "Read deterministic AI readiness scores and actionable data gaps",
				ToolName:     "UteamupAIDataReadiness",
				RESTBasePath: "/api/aianalytics",
				RESTPath:     "data-readiness",
				HTTPMethod:   "GET",
			},
			{
				Name:        "summary",
				Description: "Aggregated usage grouped by provider and operation type",
				ToolName:    "UteamupAIUsageSummary",
				RESTPath:    "summary",
				Flags: []FlagDef{
					{Name: "from", Short: "f", Description: "Window start (UTC, ISO-8601). Defaults to 30 days ago.", Type: "string"},
					{Name: "to", Short: "t", Description: "Window end (UTC, ISO-8601). Defaults to now.", Type: "string"},
				},
			},
		},
	})
}
