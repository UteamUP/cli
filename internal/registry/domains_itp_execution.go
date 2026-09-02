package registry

var inspectionTestPlanExecutionTransitionActionKeys = []string{
	"itp-execution.start",
	"itp-execution.reach-hold",
	"itp-execution.release-hold",
	"itp-execution.waive-hold",
	"itp-execution.complete-steps",
	"itp-execution.approve-release",
	"itp-execution.return",
	"itp-execution.link-ncr",
	"itp-execution.reopen",
	"itp-execution.cancel",
	"itp-execution.abort",
	"itp-execution.review-partial",
}

func init() {
	Register(&Domain{
		Name:        "itp-execution",
		Aliases:     []string{"inspection-test-plan-execution", "inspection-test-plan-executions", "itp-executions"},
		Description: "Manage executions of released inspection and test plan revisions: schedule, step results, hold and witness points, lifecycle and evidence",
		APIPath:     "/api/quality/inspection-test-plan-executions",
		Actions: []Action{
			{
				Name:              "search",
				Description:       "Search the permission-filtered execution collection",
				ToolName:          "UteamupQualityItpExecutionSearch",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					qualityAuditPublicGUIDQueryFlag("plan-guid", "inspectionTestPlanGuid", "Optional plan public GUID"),
					qualityAuditPublicGUIDQueryFlag("revision-guid", "inspectionTestPlanRevisionGuid", "Optional released revision public GUID"),
					{Name: "status", QueryName: "status", Description: "Optional execution lifecycle status", Type: "string"},
					{Name: "release-state", QueryName: "releaseState", Description: "Optional release state", Type: "string"},
					{Name: "scheduled-on-or-after-utc", QueryName: "scheduledOnOrAfterUtc", Description: "Optional inclusive UTC scheduled-start lower bound", Type: "string"},
					{Name: "scheduled-before-utc", QueryName: "scheduledBeforeUtc", Description: "Optional exclusive UTC scheduled-start upper bound", Type: "string"},
					{Name: "query", QueryName: "query", Description: "Bounded number or reference search", Type: "string"},
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one execution with frozen steps, results, point events, evidence and the server-derived lifecycle availability",
				ToolName:    "UteamupQualityItpExecutionGet",
				HTTPMethod:  "GET",
				RESTPath:    "{executionGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("executionGuid", "Execution public GUID"),
				},
			},
			{
				Name:        "create",
				Description: "Plan one execution of a released revision with a UTC schedule and participants",
				ToolName:    "UteamupQualityItpExecutionCreate",
				HTTPMethod:  "POST",
				Flags:       qualityAuditCreateMutationFlags("inspection-test-plan-execution"),
			},
			{
				Name:        "transition",
				Description: "Run one server-governed execution lifecycle transition after explicit confirmation",
				ToolName:    "UteamupQualityItpExecutionTransition",
				HTTPMethod:  "POST",
				RESTPath:    "{executionGuid}/transitions/{actionKey}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("executionGuid", "Execution public GUID"),
					{
						Name:          "actionKey",
						Description:   "Exact supported execution lifecycle action key",
						Required:      true,
						Type:          "string",
						AllowedValues: inspectionTestPlanExecutionTransitionActionKeys,
					},
				},
				Flags: qualityAuditExistingMutationFlags("inspection-test-plan-execution", true),
			},
			{
				Name:        "result-upsert",
				Description: "Record or update the result of one frozen step",
				ToolName:    "UteamupQualityItpExecutionResultUpsert",
				HTTPMethod:  "PUT",
				RESTPath:    "{executionGuid}/results",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("executionGuid", "Execution public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("inspection-test-plan-execution", false),
			},
			{
				Name:        "point-event-add",
				Description: "Record one witness or hold point event (attendance, waiver, rejection) on a step",
				ToolName:    "UteamupQualityItpExecutionPointEventAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{executionGuid}/point-events",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("executionGuid", "Execution public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("inspection-test-plan-execution", false),
			},
			{
				Name:        "evidence-add",
				Description: "Link one exact retained document version as execution evidence",
				ToolName:    "UteamupQualityItpExecutionEvidenceAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{executionGuid}/evidence",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("executionGuid", "Execution public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("inspection-test-plan-execution", false),
			},
			{
				Name:        "evidence-revoke",
				Description: "Revoke only the execution evidence association while retaining history",
				ToolName:    "UteamupQualityItpExecutionEvidenceRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{executionGuid}/evidence/{evidenceGuid}/revoke",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("executionGuid", "Execution public GUID"),
					qualityAuditPublicGUIDArgument("evidenceGuid", "Retained evidence-link public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("inspection-test-plan-execution", true),
			},
		},
	})
}
