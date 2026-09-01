package registry

var qualityAuditFindingTransitionActionKeys = []string{
	"quality-audit-finding.issue",
	"quality-audit-finding.request-response",
	"quality-audit-finding.cancel",
	"quality-audit-finding.submit-response",
	"quality-audit-finding.close-without-response",
	"quality-audit-finding.verify-close",
	"quality-audit-finding.return-response",
	"quality-audit-finding.reopen",
	"quality-audit-finding.resume-response",
}

func init() {
	Register(&Domain{
		Name:        "audit-finding",
		Aliases:     []string{"quality-audit-finding", "audit-findings", "quality-audit-findings"},
		Description: "Manage governed Quality audit findings, responses, evidence, and retained NCR/CAPA links",
		APIPath:     "/api/quality/audit-findings",
		Actions: []Action{
			{
				Name:              "search",
				Description:       "Search the permission-filtered audit-finding collection",
				ToolName:          "UteamupQualityAuditFindingSearch",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					qualityAuditPublicGUIDQueryFlag("site-location-guid", "siteLocationGuid", "Optional owning-site public GUID"),
					qualityAuditPublicGUIDQueryFlag("quality-audit-guid", "qualityAuditGuid", "Optional audit public GUID"),
					qualityAuditPublicGUIDQueryFlag("quality-audit-checklist-item-guid", "qualityAuditChecklistItemGuid", "Optional frozen checklist-item public GUID"),
					qualityAuditPublicGUIDQueryFlag("owner-user-guid", "ownerUserGuid", "Optional finding-owner public GUID"),
					{Name: "classification", QueryName: "classification", Description: "Optional finding classification", Type: "string"},
					{Name: "status", QueryName: "status", Description: "Optional finding lifecycle status", Type: "string"},
					{Name: "due-on-or-after-utc", QueryName: "dueOnOrAfterUtc", Description: "Optional inclusive UTC due-time start", Type: "string"},
					{Name: "due-before-utc", QueryName: "dueBeforeUtc", Description: "Optional exclusive UTC due-time end", Type: "string"},
					{Name: "query", QueryName: "query", Description: "Bounded title, display-number, or requirement search", Type: "string"},
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one audit finding with retained evidence and NCR/CAPA links",
				ToolName:    "UteamupQualityAuditFindingGet",
				HTTPMethod:  "GET",
				RESTPath:    "{findingGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("findingGuid", "Audit-finding public GUID"),
				},
			},
			{
				Name:        "create",
				Description: "Create one governed draft audit finding",
				ToolName:    "UteamupQualityAuditFindingCreate",
				HTTPMethod:  "POST",
				Flags:       qualityAuditCreateMutationFlags("audit-finding"),
			},
			{
				Name:        "update",
				Description: "Update lifecycle-safe audit-finding fields with optimistic concurrency",
				ToolName:    "UteamupQualityAuditFindingUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{findingGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("findingGuid", "Audit-finding public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("audit-finding", false),
			},
			{
				Name:        "transition",
				Description: "Run one server-governed audit-finding lifecycle transition after explicit confirmation",
				ToolName:    "UteamupQualityAuditFindingTransition",
				HTTPMethod:  "POST",
				RESTPath:    "{findingGuid}/transitions/{actionKey}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("findingGuid", "Audit-finding public GUID"),
					{
						Name:          "actionKey",
						Description:   "Exact supported audit-finding lifecycle action key",
						Required:      true,
						Type:          "string",
						AllowedValues: qualityAuditFindingTransitionActionKeys,
					},
				},
				Flags: qualityAuditExistingMutationFlags("audit-finding", true),
			},
			{
				Name:        "evidence-add",
				Description: "Link one exact retained document version as audit-finding evidence",
				ToolName:    "UteamupQualityAuditFindingEvidenceAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{findingGuid}/evidence",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("findingGuid", "Audit-finding public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("audit-finding", false),
			},
			{
				Name:        "evidence-revoke",
				Description: "Revoke only the audit-finding evidence association while retaining history",
				ToolName:    "UteamupQualityAuditFindingEvidenceRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{findingGuid}/evidence/{evidenceGuid}/revoke",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("findingGuid", "Audit-finding public GUID"),
					qualityAuditPublicGUIDArgument("evidenceGuid", "Retained evidence-link public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("audit-finding", true),
			},
			qualityAuditFindingLinkAction(qualityAuditFindingLinkActionDefinition{
				name:               "ncr-link-add",
				description:        "Link one existing NCR to the audit finding",
				toolName:           "UteamupQualityAuditFindingNonConformanceAdd",
				resourcePath:       "non-conformances",
				targetArgumentName: "nonConformanceGuid",
				targetDescription:  "NCR public GUID",
			}),
			qualityAuditFindingLinkAction(qualityAuditFindingLinkActionDefinition{
				name:               "ncr-link-revoke",
				description:        "Revoke only the audit-finding NCR association while retaining history",
				toolName:           "UteamupQualityAuditFindingNonConformanceRevoke",
				resourcePath:       "non-conformances",
				targetArgumentName: "linkGuid",
				targetDescription:  "Retained NCR-link public GUID",
				revoke:             true,
			}),
			qualityAuditFindingLinkAction(qualityAuditFindingLinkActionDefinition{
				name:               "capa-link-add",
				description:        "Link one existing CAPA to the audit finding",
				toolName:           "UteamupQualityAuditFindingCapaAdd",
				resourcePath:       "corrective-preventive-actions",
				targetArgumentName: "capaGuid",
				targetDescription:  "CAPA public GUID",
			}),
			qualityAuditFindingLinkAction(qualityAuditFindingLinkActionDefinition{
				name:               "capa-link-revoke",
				description:        "Revoke only the audit-finding CAPA association while retaining history",
				toolName:           "UteamupQualityAuditFindingCapaRevoke",
				resourcePath:       "corrective-preventive-actions",
				targetArgumentName: "linkGuid",
				targetDescription:  "Retained CAPA-link public GUID",
				revoke:             true,
			}),
		},
	})
}

type qualityAuditFindingLinkActionDefinition struct {
	name               string
	description        string
	toolName           string
	resourcePath       string
	targetArgumentName string
	targetDescription  string
	revoke             bool
}

func qualityAuditFindingLinkAction(definition qualityAuditFindingLinkActionDefinition) Action {
	restPath := "{findingGuid}/" + definition.resourcePath + "/{" + definition.targetArgumentName + "}"
	if definition.revoke {
		restPath += "/revoke"
	}
	return Action{
		Name:        definition.name,
		Description: definition.description,
		ToolName:    definition.toolName,
		HTTPMethod:  "POST",
		RESTPath:    restPath,
		Args: []ArgDef{
			qualityAuditPublicGUIDArgument("findingGuid", "Audit-finding public GUID"),
			qualityAuditPublicGUIDArgument(definition.targetArgumentName, definition.targetDescription),
		},
		Flags: qualityAuditExistingMutationFlags("audit-finding", definition.revoke),
	}
}
