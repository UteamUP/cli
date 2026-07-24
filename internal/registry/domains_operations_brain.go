package registry

func init() {
	Register(&Domain{
		Name:        "operations-brain",
		Aliases:     []string{"upmate-operations", "op-brain"},
		Description: "Review cross-domain risks, compare evidence-backed plans, and prepare governed runs",
		APIPath:     "/api/upmateassistant/operations",
		Actions: []Action{
			{
				Name:        "risks",
				Description: "Read bounded tenant operational risks with freshness and citation evidence",
				ToolName:    "UteamupOperationsGetRisks",
				HTTPMethod:  "GET",
				RESTPath:    "risks",
				Flags: []FlagDef{
					{Name: "domains", BodyName: "domains", Description: "Evidence domains, maximum nine", Type: "stringSlice"},
					{Name: "affected-entity-guids", BodyName: "affectedEntityGuids", Description: "Optional affected entity GUIDs", Type: "stringSlice"},
					{Name: "evidence-max-age-minutes", BodyName: "evidenceMaxAgeMinutes", Description: "Accepted evidence age in minutes, 5-10080", Type: "int", Default: 1440},
				},
			},
			{
				Name:        "create-plan",
				Description: "Create an idempotent evidence snapshot and compare deterministic scenarios",
				ToolName:    "UteamupOperationsCreatePlan",
				HTTPMethod:  "POST",
				RESTPath:    "plans",
				Flags: []FlagDef{
					{Name: "request-guid", BodyName: "requestGuid", Description: "Caller-generated idempotency GUID", Required: true, Type: "uuid"},
					{Name: "objective-key", BodyName: "objectiveKey", Description: "Stable planning objective key", Required: true, Type: "string"},
					{Name: "title", BodyName: "title", Description: "Human-readable plan title", Required: true, Type: "string"},
					{Name: "summary", BodyName: "summary", Description: "Optional evidence-backed summary", Type: "string"},
					{Name: "success-measure", BodyName: "successMeasure", Description: "Optional measurable success criterion", Type: "string"},
					{Name: "domains", BodyName: "domains", Description: "Evidence domains, maximum nine", Required: true, Type: "stringSlice"},
					{Name: "affected-entity-guids", BodyName: "affectedEntityGuids", Description: "Optional affected entity GUIDs", Type: "stringSlice"},
					{Name: "site-guids", BodyName: "siteGuids", Description: "Reserved site GUIDs (site scoping is not supported yet — the API rejects non-empty values)", Type: "stringSlice"},
					{Name: "evidence-max-age-minutes", BodyName: "evidenceMaxAgeMinutes", Description: "Accepted evidence age in minutes, 5-10080", Type: "int", Default: 1440},
				},
			},
			{
				Name:        "get-plan",
				Description: "Read one versioned plan with readiness, scenarios, constraints, and receipts",
				ToolName:    "UteamupOperationsGetPlan",
				HTTPMethod:  "GET",
				RESTPath:    "plans/{planGuid}",
				Args: []ArgDef{
					{Name: "planGuid", Description: "Operational plan GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "select-scenario",
				Description: "Select one reviewed scenario against the exact current plan version",
				ToolName:    "UteamupOperationsSelectScenario",
				HTTPMethod:  "POST",
				RESTPath:    "plans/{planGuid}/scenarios/{scenarioGuid}/select",
				Args: []ArgDef{
					{Name: "planGuid", Description: "Operational plan GUID", Required: true, Type: "uuid"},
					{Name: "scenarioGuid", Description: "Scenario GUID from the plan", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "request-guid", BodyName: "requestGuid", Description: "Caller-generated idempotency GUID", Required: true, Type: "uuid"},
					{Name: "expected-version", BodyName: "expectedVersion", Description: "Exact plan version under review", Required: true, Type: "int"},
				},
			},
			{
				Name:        "prepare-run",
				Description: "Revalidate the selected plan and prepare its approval-bound durable run",
				ToolName:    "UteamupOperationsPrepareRun",
				HTTPMethod:  "POST",
				RESTPath:    "plans/{planGuid}/runs",
				Args: []ArgDef{
					{Name: "planGuid", Description: "Operational plan GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "request-guid", BodyName: "requestGuid", Description: "Caller-generated idempotency GUID", Required: true, Type: "uuid"},
					{Name: "expected-version", BodyName: "expectedVersion", Description: "Exact selected plan version", Required: true, Type: "int"},
				},
			},
		},
	})
}
