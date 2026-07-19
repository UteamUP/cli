package registry

func init() {
	Register(&Domain{
		Name:        "reliability",
		Aliases:     []string{"rel"},
		Description: "Review GUID-first reliability evidence and governed strategies",
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
			{
				Name:        "assessments",
				Description: "List latest or historical failure-mode assessment versions",
				ToolName:    "UteamupReliabilityAssessmentsList",
				HTTPMethod:  "GET",
				RESTPath:    "assessments",
				Flags: []FlagDef{
					{Name: "asset-guid", BodyName: "assetGuid", Description: "Public asset GUID", Required: true, Type: "uuid"},
					{Name: "include-history", BodyName: "includeHistory", Description: "Include superseded versions", Type: "bool", Default: false},
				},
			},
			{
				Name:        "assessment-create",
				Description: "Create an idempotent draft failure-mode assessment",
				ToolName:    "UteamupReliabilityAssessmentCreate",
				HTTPMethod:  "POST",
				RESTPath:    "assessments",
				Flags: []FlagDef{
					{Name: "request-guid", BodyName: "requestGuid", Description: "Caller-generated idempotency GUID", Required: true, Type: "uuid"},
					{Name: "asset-guid", BodyName: "assetGuid", Description: "Public asset GUID", Required: true, Type: "uuid"},
					{Name: "failure-code-guid", BodyName: "failureCodeGuid", Description: "Optional public failure-code GUID", Type: "uuid"},
					{Name: "failure-mode-key", BodyName: "failureModeKey", Description: "Stable failure-mode key", Required: true, Type: "string"},
					{Name: "failure-mode-name", BodyName: "failureModeName", Description: "Human-readable failure-mode name", Required: true, Type: "string"},
					{Name: "severity-score", BodyName: "severityScore", Description: "Severity score, 1-10", Required: true, Type: "int"},
					{Name: "detectability-score", BodyName: "detectabilityScore", Description: "Detectability score, 1-10", Required: true, Type: "int"},
					{Name: "consequence-score", BodyName: "consequenceScore", Description: "Optional consequence score, 1-10", Type: "int"},
					{Name: "consequence-category", BodyName: "consequenceCategory", Description: "Optional consequence category", Type: "string"},
					{Name: "evidence-reference-guids", BodyName: "evidenceReferenceGuids", Description: "Comma-separated public evidence GUIDs", Type: "stringSlice"},
					{Name: "notes", BodyName: "notes", Description: "Optional reviewer notes", Type: "string"},
				},
			},
			{
				Name:        "assessment-approve",
				Description: "Approve an exact reviewed failure-mode assessment version",
				ToolName:    "UteamupReliabilityAssessmentApprove",
				HTTPMethod:  "POST",
				RESTPath:    "assessments/{versionGuid}/approve",
				Args: []ArgDef{
					{Name: "versionGuid", Description: "Assessment version GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "request-guid", BodyName: "requestGuid", Description: "Caller-generated idempotency GUID", Required: true, Type: "uuid"},
					{Name: "expected-version", BodyName: "expectedVersion", Description: "Exact version being approved", Required: true, Type: "int"},
				},
			},
			{
				Name:        "prepare-run",
				Description: "Revalidate an eligible strategy and prepare its governed durable run",
				ToolName:    "UteamupReliabilityStrategyPrepareRun",
				HTTPMethod:  "POST",
				RESTPath:    "strategies/prepare-run",
				Flags: []FlagDef{
					{Name: "request-guid", BodyName: "requestGuid", Description: "Caller-generated plan idempotency GUID", Required: true, Type: "uuid"},
					{Name: "run-request-guid", BodyName: "runRequestGuid", Description: "Caller-generated run idempotency GUID", Required: true, Type: "uuid"},
					{Name: "asset-guid", BodyName: "assetGuid", Description: "Public asset GUID", Required: true, Type: "uuid"},
					{Name: "objective", BodyName: "objectiveKey", Description: "availability, downtime, cost, or safety", Default: "availability", Type: "string"},
					{Name: "from-utc", BodyName: "fromUtc", Description: "Optional UTC evidence-window start", Type: "string"},
					{Name: "to-utc", BodyName: "toUtc", Description: "Optional UTC evidence-window end", Type: "string"},
				},
			},
		},
	})
}
