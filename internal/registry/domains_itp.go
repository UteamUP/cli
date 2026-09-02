package registry

var inspectionTestPlanRevisionTransitionActionKeys = []string{
	"itp-revision.submit",
	"itp-revision.approve",
	"itp-revision.return",
	"itp-revision.release",
	"itp-revision.approve-successor",
	"itp-revision.withdraw",
}

func init() {
	Register(&Domain{
		Name:        "itp",
		Aliases:     []string{"inspection-test-plan", "inspection-test-plans", "quality-itp"},
		Description: "Manage governed inspection and test plans, their controlled revisions, lifecycle and retained evidence",
		APIPath:     "/api/quality/inspection-test-plans",
		Actions: []Action{
			{
				Name:              "search",
				Description:       "Search the permission-filtered inspection and test plan collection",
				ToolName:          "UteamupQualityItpSearch",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					qualityAuditPublicGUIDQueryFlag("project-guid", "projectGuid", "Optional project public GUID"),
					qualityAuditPublicGUIDQueryFlag("contract-guid", "contractGuid", "Optional contract public GUID"),
					qualityAuditPublicGUIDQueryFlag("customer-guid", "customerGuid", "Optional customer public GUID"),
					qualityAuditPublicGUIDQueryFlag("vendor-guid", "vendorGuid", "Optional vendor public GUID"),
					qualityAuditPublicGUIDQueryFlag("part-guid", "partGuid", "Optional part public GUID"),
					qualityAuditPublicGUIDQueryFlag("owner-user-guid", "ownerUserGuid", "Optional plan-owner public GUID"),
					{Name: "status", QueryName: "status", Description: "Optional plan lifecycle status (Draft, Active, Withdrawn, Archived)", Type: "string"},
					{Name: "query", QueryName: "query", Description: "Bounded title, scope, or display-number search", Type: "string"},
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one plan with its revisions, retained evidence and the server-derived revision lifecycle availability",
				ToolName:    "UteamupQualityItpGet",
				HTTPMethod:  "GET",
				RESTPath:    "{planGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("planGuid", "Plan public GUID"),
				},
			},
			{
				Name:        "create",
				Description: "Create one governed draft plan with exactly one scope reference",
				ToolName:    "UteamupQualityItpCreate",
				HTTPMethod:  "POST",
				Flags:       qualityAuditCreateMutationFlags("inspection-test-plan"),
			},
			{
				Name:        "update",
				Description: "Update lifecycle-safe plan fields with optimistic concurrency",
				ToolName:    "UteamupQualityItpUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{planGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("planGuid", "Plan public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("inspection-test-plan", false),
			},
			{
				Name:        "revision-create",
				Description: "Add one draft revision (steps, parties, acceptance authority) to a plan",
				ToolName:    "UteamupQualityItpRevisionCreate",
				HTTPMethod:  "POST",
				RESTPath:    "{planGuid}/revisions",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("planGuid", "Plan public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("inspection-test-plan", false),
			},
			{
				Name:        "revision-transition",
				Description: "Run one server-governed revision lifecycle transition after explicit confirmation",
				ToolName:    "UteamupQualityItpRevisionTransition",
				HTTPMethod:  "POST",
				RESTPath:    "{planGuid}/revisions/{revisionGuid}/transitions/{actionKey}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("planGuid", "Plan public GUID"),
					qualityAuditPublicGUIDArgument("revisionGuid", "Revision public GUID"),
					{
						Name:          "actionKey",
						Description:   "Exact supported revision lifecycle action key",
						Required:      true,
						Type:          "string",
						AllowedValues: inspectionTestPlanRevisionTransitionActionKeys,
					},
				},
				Flags: qualityAuditExistingMutationFlags("inspection-test-plan", true),
			},
			{
				Name:        "evidence-add",
				Description: "Link one exact retained document version as plan evidence",
				ToolName:    "UteamupQualityItpEvidenceAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{planGuid}/evidence",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("planGuid", "Plan public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("inspection-test-plan", false),
			},
			{
				Name:        "evidence-revoke",
				Description: "Revoke only the plan evidence association while retaining history",
				ToolName:    "UteamupQualityItpEvidenceRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{planGuid}/evidence/{evidenceGuid}/revoke",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("planGuid", "Plan public GUID"),
					qualityAuditPublicGUIDArgument("evidenceGuid", "Retained evidence-link public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("inspection-test-plan", true),
			},
		},
	})
}
