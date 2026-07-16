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
				Name:         "outcomes",
				Description:  "List predicted, pending and measured operational outcomes",
				ToolName:     "UteamupAIOutcomeMeasurements",
				RESTBasePath: "/api/aianalytics",
				RESTPath:     "outcomes",
				HTTPMethod:   "GET",
				Flags: []FlagDef{
					{Name: "domain", Description: "Filter by operational domain", Type: "string"},
					{Name: "capability-key", Description: "Filter by UPMate capability key", Type: "string"},
					{Name: "metric-key", Description: "Filter by outcome metric key", Type: "string"},
					{Name: "status", Description: "Pending, BaselineCaptured, Measured, Unavailable or Invalidated", Type: "string"},
					{Name: "from-utc", Description: "Inclusive creation-window start (ISO-8601 UTC)", Type: "string"},
					{Name: "to-utc", Description: "Exclusive creation-window end (ISO-8601 UTC)", Type: "string"},
					{Name: "limit", Description: "Maximum rows (1-200)", Type: "int", Default: 100},
				},
			},
			{
				Name:         "knowledge-tutorial-trust",
				Description:  "Read Knowledge citation readiness, gaps and tutorial adoption telemetry",
				ToolName:     "UteamupAIKnowledgeTutorialTrust",
				RESTBasePath: "/api/aianalytics",
				RESTPath:     "knowledge-tutorial-trust",
				HTTPMethod:   "GET",
			},
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
