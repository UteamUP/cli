package registry

// scarInternalTransitionActionKeys are the keys the tenant /transitions route accepts. Response
// dispositions and effectiveness verdicts have their own routes; supplier-only keys are never
// available to a tenant session.
var scarInternalTransitionActionKeys = []string{
	"scar.submit",
	"scar.approve-issue",
	"scar.return",
	"scar.cancel",
	"scar.withdraw",
	"scar.escalate",
	"scar.begin-internal-review",
	"scar.complete-implementation",
	"scar.reopen",
	"scar.resume-supplier-correction",
}

func init() {
	Register(&Domain{
		Name:        "scar",
		Aliases:     []string{"supplier-corrective-action-request", "supplier-corrective-action-requests", "supplier-quality"},
		Description: "Manage the tenant side of supplier corrective action requests: issue, supplier access grants, response reviews, effectiveness verdicts, links and evidence",
		APIPath:     "/api/quality/supplier-corrective-action-requests",
		Actions: []Action{
			{
				Name:              "search",
				Description:       "Search the permission-filtered SCAR collection",
				ToolName:          "UteamupQualityScarSearch",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					qualityAuditPublicGUIDQueryFlag("site-location-guid", "siteLocationGuid", "Optional owning site location public GUID"),
					qualityAuditPublicGUIDQueryFlag("vendor-guid", "vendorGuid", "Optional vendor public GUID"),
					qualityAuditPublicGUIDQueryFlag("owner-user-guid", "ownerUserGuid", "Optional owner public GUID"),
					{Name: "status", QueryName: "status", Description: "Optional SCAR lifecycle status", Type: "string"},
					{Name: "severity", QueryName: "severity", Description: "Optional severity (Minor, Major, Critical)", Type: "string"},
					{Name: "due-on-or-after-utc", QueryName: "dueOnOrAfterUtc", Description: "Optional inclusive UTC full-response due lower bound", Type: "string"},
					{Name: "due-before-utc", QueryName: "dueBeforeUtc", Description: "Optional exclusive UTC full-response due upper bound", Type: "string"},
					{Name: "query", QueryName: "query", Description: "Bounded number, title, or problem-statement search", Type: "string"},
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one SCAR with grants, supplier responses, evidence and the server-derived internal lifecycle availability",
				ToolName:    "UteamupQualityScarGet",
				HTTPMethod:  "GET",
				RESTPath:    "{scarGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
				},
			},
			{
				Name:        "create",
				Description: "Create one draft SCAR for a vendor with ordered UTC due dates and at least one context reference",
				ToolName:    "UteamupQualityScarCreate",
				HTTPMethod:  "POST",
				Flags:       qualityAuditCreateMutationFlags("scar"),
			},
			{
				Name:        "update",
				Description: "Update the editable fields of one draft SCAR with optimistic concurrency",
				ToolName:    "UteamupQualityScarUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{scarGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", false),
			},
			{
				Name:        "transition",
				Description: "Run one server-governed internal SCAR lifecycle transition after explicit confirmation",
				ToolName:    "UteamupQualityScarTransition",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/transitions/{actionKey}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
					{
						Name:          "actionKey",
						Description:   "Exact supported internal SCAR lifecycle action key",
						Required:      true,
						Type:          "string",
						AllowedValues: scarInternalTransitionActionKeys,
					},
				},
				Flags: qualityAuditExistingMutationFlags("scar", true),
			},
			{
				Name:        "evidence-add",
				Description: "Link one exact retained document version as SCAR evidence",
				ToolName:    "UteamupQualityScarEvidenceAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/evidence",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", false),
			},
			{
				Name:        "evidence-revoke",
				Description: "Revoke only the SCAR evidence association while retaining history",
				ToolName:    "UteamupQualityScarEvidenceRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/evidence/{evidenceGuid}/revoke",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
					qualityAuditPublicGUIDArgument("evidenceGuid", "Retained evidence-link public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", true),
			},
			{
				Name:        "grant-create",
				Description: "Grant one portal-active contractor of the vendor bounded supplier actions for a limited time",
				ToolName:    "UteamupQualityScarGrantCreate",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/external-grants",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", false),
			},
			{
				Name:        "grant-revoke",
				Description: "Revoke one supplier access grant with a retained reason after explicit confirmation",
				ToolName:    "UteamupQualityScarGrantRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/external-grants/{grantGuid}/revoke",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
					qualityAuditPublicGUIDArgument("grantGuid", "Grant public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", true),
			},
			{
				Name:        "response-review",
				Description: "Accept, return or reject one current-cycle supplier response revision after explicit confirmation",
				ToolName:    "UteamupQualityScarResponseReview",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/response-reviews",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", true),
			},
			{
				Name:        "effectiveness-verify",
				Description: "Record the independent effectiveness verdict that closes the SCAR or returns it to the supplier",
				ToolName:    "UteamupQualityScarEffectivenessVerify",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/effectiveness-reviews",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", true),
			},
			{
				Name:        "non-conformance-add",
				Description: "Link one source non-conformance to the SCAR",
				ToolName:    "UteamupQualityScarNonConformanceAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/non-conformances/{nonConformanceGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
					qualityAuditPublicGUIDArgument("nonConformanceGuid", "Non-conformance public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", false),
			},
			{
				Name:        "non-conformance-revoke",
				Description: "Revoke one source non-conformance link with a retained reason after explicit confirmation",
				ToolName:    "UteamupQualityScarLinkRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/non-conformances/{linkGuid}/revoke",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
					qualityAuditPublicGUIDArgument("linkGuid", "Link public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", true),
			},
			{
				Name:        "capa-add",
				Description: "Link one corrective and preventive action that follows from the SCAR",
				ToolName:    "UteamupQualityScarCapaAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/corrective-preventive-actions/{capaGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
					qualityAuditPublicGUIDArgument("capaGuid", "CAPA public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", false),
			},
			{
				Name:        "capa-revoke",
				Description: "Revoke one CAPA link with a retained reason after explicit confirmation",
				ToolName:    "UteamupQualityScarLinkRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/corrective-preventive-actions/{linkGuid}/revoke",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
					qualityAuditPublicGUIDArgument("linkGuid", "Link public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", true),
			},
			{
				Name:        "communication-add",
				Description: "Retain one vendor message as SCAR communication",
				ToolName:    "UteamupQualityScarCommunicationAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/communications",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", false),
			},
			{
				Name:        "communication-revoke",
				Description: "Revoke one retained communication link with a reason after explicit confirmation",
				ToolName:    "UteamupQualityScarLinkRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{scarGuid}/communications/{linkGuid}/revoke",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("scarGuid", "SCAR public GUID"),
					qualityAuditPublicGUIDArgument("linkGuid", "Link public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("scar", true),
			},
		},
	})
}
