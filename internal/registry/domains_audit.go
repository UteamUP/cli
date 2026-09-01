package registry

var qualityAuditTransitionActionKeys = []string{
	"quality-audit.approve-plan",
	"quality-audit.schedule",
	"quality-audit.accept-assignments",
	"quality-audit.cancel",
	"quality-audit.open",
	"quality-audit.close-fieldwork",
	"quality-audit.abort",
	"quality-audit.document-partial-evidence",
	"quality-audit.issue-report",
	"quality-audit.issue-report-no-follow-up",
	"quality-audit.verify-follow-up",
	"quality-audit.reopen",
	"quality-audit.resume-follow-up",
}

func init() {
	Register(&Domain{
		Name:        "audit",
		Aliases:     []string{"quality-audit", "quality-audits"},
		Description: "Manage governed Quality audit execution, assurance, checklist evaluation, and evidence",
		APIPath:     "/api/quality/audits",
		Actions: []Action{
			{
				Name:              "search",
				Description:       "Search the permission-filtered audit collection",
				ToolName:          "UteamupQualityAuditSearch",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					qualityAuditPublicGUIDQueryFlag("site-location-guid", "siteLocationGuid", "Optional owning-site public GUID"),
					qualityAuditPublicGUIDQueryFlag("quality-audit-program-guid", "qualityAuditProgramGuid", "Optional audit-program public GUID"),
					qualityAuditPublicGUIDQueryFlag("vendor-guid", "vendorGuid", "Optional supplier-vendor public GUID"),
					qualityAuditPublicGUIDQueryFlag("lead-auditor-user-guid", "leadAuditorUserGuid", "Optional lead-auditor public GUID"),
					{Name: "status", QueryName: "status", Description: "Optional audit lifecycle status", Type: "string"},
					{Name: "type", QueryName: "type", Description: "Optional audit type", Type: "string"},
					{Name: "scheduled-on-or-after-utc", QueryName: "scheduledOnOrAfterUtc", Description: "Optional inclusive UTC schedule start", Type: "string"},
					{Name: "scheduled-before-utc", QueryName: "scheduledBeforeUtc", Description: "Optional exclusive UTC schedule end", Type: "string"},
					{Name: "query", QueryName: "query", Description: "Bounded title, display-number, or scope search", Type: "string"},
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one audit with assignments, checklist, findings, and retained evidence",
				ToolName:    "UteamupQualityAuditGet",
				HTTPMethod:  "GET",
				RESTPath:    "{auditGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("auditGuid", "Audit public GUID"),
				},
			},
			{
				Name:        "create",
				Description: "Create one governed planned audit",
				ToolName:    "UteamupQualityAuditCreate",
				HTTPMethod:  "POST",
				Flags:       qualityAuditCreateMutationFlags("audit"),
			},
			{
				Name:        "update",
				Description: "Update lifecycle-safe audit fields with optimistic concurrency",
				ToolName:    "UteamupQualityAuditUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{auditGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("auditGuid", "Audit public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("audit", false),
			},
			{
				Name:        "transition",
				Description: "Run one server-governed audit lifecycle transition after explicit confirmation",
				ToolName:    "UteamupQualityAuditTransition",
				HTTPMethod:  "POST",
				RESTPath:    "{auditGuid}/transitions/{actionKey}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("auditGuid", "Audit public GUID"),
					{
						Name:          "actionKey",
						Description:   "Exact supported audit lifecycle action key",
						Required:      true,
						Type:          "string",
						AllowedValues: qualityAuditTransitionActionKeys,
					},
				},
				Flags: qualityAuditExistingMutationFlags("audit", true),
			},
			qualityAuditAssignmentAction(qualityAuditAssignmentActionDefinition{
				name:        "assignment-verify-competence",
				description: "Verify one audit assignment's competence with retained attestation evidence",
				toolName:    "UteamupQualityAuditAssignmentCompetenceVerify",
				routeAction: "verify-competence",
			}),
			qualityAuditAssignmentAction(qualityAuditAssignmentActionDefinition{
				name:        "assignment-review-independence",
				description: "Review one audit assignment's independence with retained attestation evidence",
				toolName:    "UteamupQualityAuditAssignmentIndependenceReview",
				routeAction: "review-independence",
			}),
			qualityAuditAssignmentAction(qualityAuditAssignmentActionDefinition{
				name:        "assignment-accept",
				description: "Accept one audit assignment with retained attestation evidence",
				toolName:    "UteamupQualityAuditAssignmentAccept",
				routeAction: "accept",
			}),
			qualityAuditAssignmentAction(qualityAuditAssignmentActionDefinition{
				name:        "assignment-remove",
				description: "Remove one active non-lead audit assignment while retaining history",
				toolName:    "UteamupQualityAuditAssignmentRemove",
				routeAction: "remove",
			}),
			{
				Name:        "checklist-evaluate",
				Description: "Evaluate one frozen audit checklist item after explicit confirmation",
				ToolName:    "UteamupQualityAuditChecklistEvaluate",
				HTTPMethod:  "POST",
				RESTPath:    "{auditGuid}/checklist-items/{checklistItemGuid}/evaluate",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("auditGuid", "Audit public GUID"),
					qualityAuditPublicGUIDArgument("checklistItemGuid", "Frozen checklist-item public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("audit", true),
			},
			{
				Name:        "evidence-add",
				Description: "Link one exact retained document version as audit evidence",
				ToolName:    "UteamupQualityAuditEvidenceAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{auditGuid}/evidence",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("auditGuid", "Audit public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("audit", false),
			},
			{
				Name:        "evidence-revoke",
				Description: "Revoke only the audit evidence association while retaining history",
				ToolName:    "UteamupQualityAuditEvidenceRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{auditGuid}/evidence/{evidenceGuid}/revoke",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("auditGuid", "Audit public GUID"),
					qualityAuditPublicGUIDArgument("evidenceGuid", "Retained evidence-link public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("audit", true),
			},
		},
	})
}

type qualityAuditAssignmentActionDefinition struct {
	name        string
	description string
	toolName    string
	routeAction string
}

func qualityAuditAssignmentAction(definition qualityAuditAssignmentActionDefinition) Action {
	return Action{
		Name:        definition.name,
		Description: definition.description,
		ToolName:    definition.toolName,
		HTTPMethod:  "POST",
		RESTPath:    "{auditGuid}/assignments/{assignmentGuid}/" + definition.routeAction,
		Args: []ArgDef{
			qualityAuditPublicGUIDArgument("auditGuid", "Audit public GUID"),
			qualityAuditPublicGUIDArgument("assignmentGuid", "Audit-assignment public GUID"),
		},
		Flags: qualityAuditExistingMutationFlags("audit", true),
	}
}
