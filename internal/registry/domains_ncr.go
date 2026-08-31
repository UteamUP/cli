package registry

var nonConformanceTransitionActionKeys = []string{
	"ncr.submit",
	"ncr.accept-triage",
	"ncr.return-for-correction",
	"ncr.withdraw",
	"ncr.revise",
	"ncr.assign-containment",
	"ncr.approve-not-substantiated",
	"ncr.link-canonical-duplicate",
	"ncr.cancel",
	"ncr.complete-containment",
	"ncr.approve-analysis",
	"ncr.approve-disposition",
	"ncr.complete-work",
	"ncr.verify-close",
	"ncr.fail-verification",
	"ncr.reopen",
	"ncr.unlink-reopen",
	"ncr.resume-triage",
}

var nonConformanceRelationshipKinds = []string{
	"workorder",
	"project",
	"asset",
	"location",
	"part",
	"stock-item",
	"vendor",
	"vehicle-inspection",
	"root-cause-analysis",
	"operational-route-execution",
}

func init() {
	Register(&Domain{
		Name:        "ncr",
		Aliases:     []string{"nonconformance", "non-conformance"},
		Description: "Manage governed non-conformance records, evidence, and typed links",
		APIPath:     "/api/quality/non-conformances",
		Actions: []Action{
			{
				Name:              "search",
				Description:       "Search the permission-filtered NCR collection",
				ToolName:          "UteamupNonConformanceSearch",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					{Name: "owning-site-location-guid", QueryName: "owningSiteLocationGuid", Description: "Optional owning-site public GUID", Type: "uuid"},
					{Name: "status", QueryName: "status", Description: "Optional NCR lifecycle status", Type: "string"},
					{Name: "severity", QueryName: "severity", Description: "Optional Minor, Major, or Critical severity", Type: "string"},
					{Name: "risk-level", QueryName: "riskLevel", Description: "Optional Low, Medium, High, or Critical risk", Type: "string"},
					{Name: "query", QueryName: "query", Description: "Bounded title, number, or requirement search", Type: "string"},
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one NCR with retained evidence and server-derived transitions",
				ToolName:    "UteamupNonConformanceGet",
				HTTPMethod:  "GET",
				RESTPath:    "{nonConformanceGuid}",
				Args:        []ArgDef{nonConformanceGUIDArgument()},
			},
			{
				Name:        "create",
				Description: "Create one governed draft NCR from a root-object JSON request",
				ToolName:    "UteamupNonConformanceCreate",
				HTTPMethod:  "POST",
				Flags:       nonConformanceCreateFlags(),
			},
			{
				Name:        "update",
				Description: "Update lifecycle-safe NCR fields with optimistic concurrency",
				ToolName:    "UteamupNonConformanceUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{nonConformanceGuid}",
				Args:        []ArgDef{nonConformanceGUIDArgument()},
				Flags:       nonConformanceExistingMutationFlags(false),
			},
			{
				Name:        "transition",
				Description: "Run one server-governed NCR lifecycle transition after explicit confirmation",
				ToolName:    "UteamupNonConformanceTransition",
				HTTPMethod:  "POST",
				RESTPath:    "{nonConformanceGuid}/transitions/{actionKey}",
				Args: []ArgDef{
					nonConformanceGUIDArgument(),
					{
						Name:          "actionKey",
						Description:   "Exact action key returned by transitionAvailability",
						Required:      true,
						Type:          "string",
						AllowedValues: nonConformanceTransitionActionKeys,
					},
				},
				Flags: nonConformanceExistingMutationFlags(true),
			},
			{
				Name:        "evidence-add",
				Description: "Link one exact retained document version as NCR evidence",
				ToolName:    "UteamupNonConformanceEvidenceAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{nonConformanceGuid}/evidence",
				Args:        []ArgDef{nonConformanceGUIDArgument()},
				Flags:       nonConformanceExistingMutationFlags(false),
			},
			{
				Name:        "evidence-revoke",
				Description: "Revoke one NCR evidence link while retaining its audit history",
				ToolName:    "UteamupNonConformanceEvidenceRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{nonConformanceGuid}/evidence/{evidenceGuid}/revoke",
				Args: []ArgDef{
					nonConformanceGUIDArgument(),
					{Name: "evidenceGuid", Description: "Retained evidence-link public GUID", Required: true, Type: "uuid"},
				},
				Flags: nonConformanceExistingMutationFlags(true),
			},
			{
				Name:        "link-add",
				Description: "Add one explicit typed relationship to an authoritative record",
				ToolName:    "UteamupNonConformanceLinkAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{nonConformanceGuid}/links/{linkKind}/{targetGuid}",
				Args: []ArgDef{
					nonConformanceGUIDArgument(),
					nonConformanceLinkKindArgument(),
					{Name: "targetGuid", Description: "Authoritative target-record public GUID", Required: true, Type: "uuid"},
				},
				Flags: nonConformanceExistingMutationFlags(false),
			},
			{
				Name:        "link-revoke",
				Description: "Revoke one typed NCR relationship while retaining its audit history",
				ToolName:    "UteamupNonConformanceLinkRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{nonConformanceGuid}/links/{linkKind}/{linkGuid}/revoke",
				Args: []ArgDef{
					nonConformanceGUIDArgument(),
					nonConformanceLinkKindArgument(),
					{Name: "linkGuid", Description: "Retained relationship public GUID", Required: true, Type: "uuid"},
				},
				Flags: nonConformanceExistingMutationFlags(true),
			},
		},
	})
}

func nonConformanceGUIDArgument() ArgDef {
	return ArgDef{
		Name:        "nonConformanceGuid",
		Description: "NCR public GUID",
		Required:    true,
		Type:        "uuid",
	}
}

func nonConformanceLinkKindArgument() ArgDef {
	return ArgDef{
		Name:          "linkKind",
		Description:   "Supported typed NCR relationship kind",
		Required:      true,
		Type:          "string",
		AllowedValues: nonConformanceRelationshipKinds,
	}
}

func nonConformanceCreateFlags() []FlagDef {
	return []FlagDef{
		nonConformanceRequestFileFlag(),
		nonConformanceIdempotencyFlag(),
	}
}

func nonConformanceExistingMutationFlags(requireConfirmation bool) []FlagDef {
	flags := []FlagDef{
		nonConformanceRequestFileFlag(),
		nonConformanceIdempotencyFlag(),
		nonConformanceConcurrencyFlag(),
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

func nonConformanceRequestFileFlag() FlagDef {
	return FlagDef{
		Name:               "request-file",
		Short:              "f",
		Description:        "Path to the exact NCR request DTO as one root JSON object",
		Required:           true,
		Type:               "string",
		RootJSONObjectFile: true,
	}
}

func nonConformanceIdempotencyFlag() FlagDef {
	return FlagDef{
		Name:        "idempotency-key",
		Description: "Caller-generated GUID reused only for the exact same mutation",
		Required:    true,
		Type:        "uuid",
		HeaderName:  "Idempotency-Key",
	}
}

func nonConformanceConcurrencyFlag() FlagDef {
	return FlagDef{
		Name:        "concurrency-token",
		Description: "Current opaque concurrency token from the NCR response",
		Required:    true,
		Type:        "string",
		Sensitive:   true,
		HeaderName:  "If-Match",
		StrongETag:  true,
	}
}
