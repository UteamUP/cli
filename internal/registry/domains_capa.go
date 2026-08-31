package registry

var correctivePreventiveActionTransitionActionKeys = []string{
	"capa.submit",
	"capa.require-rca",
	"capa.complete-rca",
	"capa.complete-plan",
	"capa.approve-plan",
	"capa.return-plan",
	"capa.complete-actions",
	"capa.reach-effectiveness-window",
	"capa.approve-effective",
	"capa.mark-ineffective",
	"capa.resubmit-plan",
	"capa.reopen",
	"capa.resume-assessment",
}

func init() {
	Register(&Domain{
		Name:        "capa",
		Aliases:     []string{"corrective-preventive-action", "corrective-preventive-actions"},
		Description: "Manage governed corrective and preventive actions, exact evidence, and source NCR provenance",
		APIPath:     "/api/quality/corrective-preventive-actions",
		Actions: []Action{
			{
				Name:              "search",
				Description:       "Search the permission-filtered CAPA collection",
				ToolName:          "UteamupCorrectivePreventiveActionSearch",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					{Name: "owning-site-location-guid", QueryName: "owningSiteLocationGuid", Description: "Optional owning-site public GUID", Type: "uuid"},
					{Name: "status", QueryName: "status", Description: "Optional CAPA lifecycle status", Type: "string"},
					{Name: "type", QueryName: "type", Description: "Optional corrective, preventive, or combined CAPA type", Type: "string"},
					{Name: "initial-risk-level", QueryName: "initialRiskLevel", Description: "Optional Low, Medium, High, or Critical initial risk", Type: "string"},
					{Name: "owner-user-guid", QueryName: "ownerUserGuid", Description: "Optional assigned-owner public GUID", Type: "uuid"},
					{Name: "query", QueryName: "query", Description: "Bounded title, display-number, or problem-statement search", Type: "string"},
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one CAPA with retained evidence and source NCR provenance",
				ToolName:    "UteamupCorrectivePreventiveActionGet",
				HTTPMethod:  "GET",
				RESTPath:    "{correctivePreventiveActionGuid}",
				Args:        []ArgDef{correctivePreventiveActionGUIDArgument()},
			},
			{
				Name:        "create",
				Description: "Create one governed draft CAPA from a root-object JSON request",
				ToolName:    "UteamupCorrectivePreventiveActionCreate",
				HTTPMethod:  "POST",
				Flags:       correctivePreventiveActionCreateFlags(),
			},
			{
				Name:        "update",
				Description: "Update lifecycle-safe CAPA fields with optimistic concurrency",
				ToolName:    "UteamupCorrectivePreventiveActionUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{correctivePreventiveActionGuid}",
				Args:        []ArgDef{correctivePreventiveActionGUIDArgument()},
				Flags:       correctivePreventiveActionExistingMutationFlags(false),
			},
			{
				Name:        "transition",
				Description: "Run one server-governed CAPA lifecycle transition after explicit confirmation",
				ToolName:    "UteamupCorrectivePreventiveActionTransition",
				HTTPMethod:  "POST",
				RESTPath:    "{correctivePreventiveActionGuid}/transitions/{actionKey}",
				Args: []ArgDef{
					correctivePreventiveActionGUIDArgument(),
					{
						Name:          "actionKey",
						Description:   "Exact supported CAPA lifecycle action key",
						Required:      true,
						Type:          "string",
						AllowedValues: correctivePreventiveActionTransitionActionKeys,
					},
				},
				Flags: correctivePreventiveActionExistingMutationFlags(true),
			},
			{
				Name:        "evidence-add",
				Description: "Link one exact retained DocumentVersion as CAPA evidence without altering provider content",
				ToolName:    "UteamupCorrectivePreventiveActionEvidenceAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{correctivePreventiveActionGuid}/evidence",
				Args:        []ArgDef{correctivePreventiveActionGUIDArgument()},
				Flags:       correctivePreventiveActionExistingMutationFlags(false),
			},
			{
				Name:        "evidence-revoke",
				Description: "Revoke only the CAPA evidence association; this does not delete the Document, DocumentVersion, SharePoint item through Microsoft Graph, other provider content, or immutable history",
				ToolName:    "UteamupCorrectivePreventiveActionEvidenceRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{correctivePreventiveActionGuid}/evidence/{evidenceGuid}/revoke",
				Args: []ArgDef{
					correctivePreventiveActionGUIDArgument(),
					{Name: "evidenceGuid", Description: "Retained evidence-link public GUID", Required: true, Type: "uuid"},
				},
				Flags: correctivePreventiveActionExistingMutationFlags(true),
			},
			{
				Name:        "source-ncr-add",
				Description: "Add one exact NCR as a retained CAPA source association",
				ToolName:    "UteamupCorrectivePreventiveActionSourceNonConformanceAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{correctivePreventiveActionGuid}/source-non-conformances/{sourceNonConformanceGuid}",
				Args: []ArgDef{
					correctivePreventiveActionGUIDArgument(),
					{Name: "sourceNonConformanceGuid", Description: "Route source NCR public GUID; must match request JSON nonConformanceGuid", Required: true, Type: "uuid"},
				},
				Flags: correctivePreventiveActionExistingMutationFlags(false),
			},
			{
				Name:        "source-ncr-revoke",
				Description: "Revoke only the CAPA-to-NCR association; this does not delete the source NCR and preserves immutable link history",
				ToolName:    "UteamupCorrectivePreventiveActionSourceNonConformanceRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{correctivePreventiveActionGuid}/source-non-conformances/{sourceLinkGuid}/revoke",
				Args: []ArgDef{
					correctivePreventiveActionGUIDArgument(),
					{Name: "sourceLinkGuid", Description: "Retained CAPA source-link public GUID", Required: true, Type: "uuid"},
				},
				Flags: correctivePreventiveActionExistingMutationFlags(true),
			},
		},
	})
}

func correctivePreventiveActionGUIDArgument() ArgDef {
	return ArgDef{
		Name:        "correctivePreventiveActionGuid",
		Description: "CAPA public GUID",
		Required:    true,
		Type:        "uuid",
	}
}

func correctivePreventiveActionCreateFlags() []FlagDef {
	return []FlagDef{
		correctivePreventiveActionRequestFileFlag(),
		correctivePreventiveActionIdempotencyFlag(),
	}
}

func correctivePreventiveActionExistingMutationFlags(requireConfirmation bool) []FlagDef {
	flags := []FlagDef{
		correctivePreventiveActionRequestFileFlag(),
		correctivePreventiveActionIdempotencyFlag(),
		correctivePreventiveActionConcurrencyFlag(),
	}
	if requireConfirmation {
		flags = append(flags, FlagDef{
			Name:        "confirm",
			Description: "Explicitly confirm the reviewed retained mutation",
			Required:    true,
			Type:        "bool",
			MustBeTrue:  true,
			LocalOnly:   true,
		})
	}
	return flags
}

func correctivePreventiveActionRequestFileFlag() FlagDef {
	return FlagDef{
		Name:               "request-file",
		Short:              "f",
		Description:        "Path to the exact CAPA request DTO as one root JSON object",
		Required:           true,
		Type:               "string",
		RootJSONObjectFile: true,
	}
}

func correctivePreventiveActionIdempotencyFlag() FlagDef {
	return FlagDef{
		Name:        "idempotency-key",
		Description: "Caller-generated GUID reused only for the exact same mutation",
		Required:    true,
		Type:        "uuid",
		HeaderName:  "Idempotency-Key",
	}
}

func correctivePreventiveActionConcurrencyFlag() FlagDef {
	return FlagDef{
		Name:        "concurrency-token",
		Description: "Current opaque concurrency token from the CAPA response",
		Required:    true,
		Type:        "string",
		Sensitive:   true,
		HeaderName:  "If-Match",
		StrongETag:  true,
	}
}
