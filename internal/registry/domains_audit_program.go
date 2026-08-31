package registry

var qualityAuditProgramTransitionActionKeys = []string{
	"quality-audit-program.submit",
	"quality-audit-program.return",
	"quality-audit-program.approve",
	"quality-audit-program.activate",
	"quality-audit-program.cancel",
	"quality-audit-program.complete",
}

func init() {
	Register(&Domain{
		Name:        "audit-program",
		Aliases:     []string{"quality-audit-program", "audit-programme", "quality-audit-programme"},
		Description: "Manage governed Quality audit programs, lifecycle, and retained evidence",
		APIPath:     "/api/quality/audit-programs",
		Actions: []Action{
			{
				Name:              "search",
				Description:       "Search the permission-filtered audit-program collection",
				ToolName:          "UteamupQualityAuditProgramSearch",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					qualityAuditPublicGUIDQueryFlag("project-guid", "projectGuid", "Optional project public GUID"),
					qualityAuditPublicGUIDQueryFlag("owner-user-guid", "ownerUserGuid", "Optional program-owner public GUID"),
					{Name: "status", QueryName: "status", Description: "Optional audit-program lifecycle status", Type: "string"},
					{Name: "period-starts-on-or-after-utc", QueryName: "periodStartsOnOrAfterUtc", Description: "Optional inclusive UTC period start", Type: "string"},
					{Name: "period-ends-on-or-before-utc", QueryName: "periodEndsOnOrBeforeUtc", Description: "Optional inclusive UTC period end", Type: "string"},
					{Name: "query", QueryName: "query", Description: "Bounded title, display-number, or description search", Type: "string"},
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one audit program with retained evidence and audit counts",
				ToolName:    "UteamupQualityAuditProgramGet",
				HTTPMethod:  "GET",
				RESTPath:    "{programmeGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("programmeGuid", "Audit-program public GUID"),
				},
			},
			{
				Name:        "create",
				Description: "Create one governed draft audit program",
				ToolName:    "UteamupQualityAuditProgramCreate",
				HTTPMethod:  "POST",
				Flags:       qualityAuditCreateMutationFlags("audit-program"),
			},
			{
				Name:        "update",
				Description: "Update lifecycle-safe audit-program fields with optimistic concurrency",
				ToolName:    "UteamupQualityAuditProgramUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{programmeGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("programmeGuid", "Audit-program public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("audit-program", false),
			},
			{
				Name:        "transition",
				Description: "Run one server-governed audit-program lifecycle transition after explicit confirmation",
				ToolName:    "UteamupQualityAuditProgramTransition",
				HTTPMethod:  "POST",
				RESTPath:    "{programmeGuid}/transitions/{actionKey}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("programmeGuid", "Audit-program public GUID"),
					{
						Name:          "actionKey",
						Description:   "Exact supported audit-program lifecycle action key",
						Required:      true,
						Type:          "string",
						AllowedValues: qualityAuditProgramTransitionActionKeys,
					},
				},
				Flags: qualityAuditExistingMutationFlags("audit-program", true),
			},
			{
				Name:        "evidence-add",
				Description: "Link one exact retained document version as audit-program evidence",
				ToolName:    "UteamupQualityAuditProgramEvidenceAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{programmeGuid}/evidence",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("programmeGuid", "Audit-program public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("audit-program", false),
			},
			{
				Name:        "evidence-revoke",
				Description: "Revoke only the audit-program evidence association while retaining history",
				ToolName:    "UteamupQualityAuditProgramEvidenceRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{programmeGuid}/evidence/{evidenceGuid}/revoke",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("programmeGuid", "Audit-program public GUID"),
					qualityAuditPublicGUIDArgument("evidenceGuid", "Retained evidence-link public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("audit-program", true),
			},
		},
	})
}
