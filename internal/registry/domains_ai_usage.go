package registry

// AI request usage summary for the caller's tenant.
// Requires AIUsage.ViewTenant permission on the backend.

func init() {
	Register(&Domain{
		Name:        "ai-usage",
		Aliases:     []string{"aiu"},
		Description: "View AI request usage (tokens, estimated cost) for the tenant",
		Actions: []Action{
			{
				Name:         "requests",
				Description:  "Page through tenant AI requests, actors, net credits and known costs",
				ToolName:     "UteamupAIUsageRequests",
				RESTBasePath: "/api/aiusage",
				RESTPath:     "requests",
				HTTPMethod:   "GET",
				Flags: []FlagDef{
					{Name: "from", Description: "Inclusive UTC timestamp ending in Z", Type: "string"},
					{Name: "to", Description: "Exclusive UTC timestamp ending in Z (maximum window 366 days)", Type: "string"},
					{Name: "page", Description: "Page number starting at 1", Type: "int", Default: 1},
					{Name: "page-size", Description: "Rows per page (capped at 100)", Type: "int", Default: 25},
					{Name: "task-key", Description: "Filter by task key", Type: "string"},
					{Name: "operation-key", Description: "Filter by operation key", Type: "string"},
					{Name: "credential-source", Description: "LegacyManaged, GlobalManaged, TenantByok, ManagedRealtimeTranscription or DeviceLocal", Type: "string"},
					{Name: "user-guid", Description: "Filter by a public user GUID in this tenant", Type: "string"},
				},
			},
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
