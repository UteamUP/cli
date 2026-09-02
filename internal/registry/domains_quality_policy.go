package registry

// qualityPolicyRecordKinds are the record kinds that ship an enterprise default definition; the
// three unbuilt aggregates (CustomerQualityCase, QualityObjective, ManagementReview) have none.
var qualityPolicyRecordKinds = []string{
	"NonConformance",
	"CorrectivePreventiveAction",
	"QualityAuditProgram",
	"QualityAudit",
	"QualityAuditFinding",
	"InspectionTestPlan",
	"InspectionTestPlanExecution",
	"SupplierCorrectiveActionRequest",
}

func init() {
	Register(&Domain{
		Name:        "quality-policy",
		Aliases:     []string{"policy", "quality-workflow-policy", "workflow-policy"},
		Description: "Administer tenant Quality workflow-policy versions: defaults, drafts, review, publication and supersession",
		APIPath:     "/api/quality/policies",
		Actions: []Action{
			{
				Name:              "search",
				Description:       "Search one bounded page of tenant Quality workflow-policy versions",
				ToolName:          "UteamupQualityPolicySearch",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					{Name: "record-kind", QueryName: "recordKind", Description: "Optional Quality record kind", Type: "string"},
					{Name: "status", QueryName: "status", Description: "Optional policy status (Draft, InReview, Published, Superseded, Withdrawn)", Type: "string"},
					{Name: "scope", QueryName: "scope", Description: "Optional policy scope (Tenant or Site)", Type: "string"},
					qualityAuditPublicGUIDQueryFlag("site-location-guid", "siteLocationGuid", "Optional owning site location public GUID"),
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one policy version; the ETag response header is the concurrency token for later mutations",
				ToolName:    "UteamupQualityPolicyGet",
				HTTPMethod:  "GET",
				RESTPath:    "{policyGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("policyGuid", "Policy public GUID"),
				},
			},
			{
				Name:        "defaults",
				Description: "Get the enterprise default definition for one record kind plus the version the next draft must request",
				ToolName:    "UteamupQualityPolicyDefaultsGet",
				HTTPMethod:  "GET",
				RESTPath:    "defaults/{recordKind}",
				Args: []ArgDef{
					{
						Name:          "recordKind",
						Description:   "Quality record kind with a shipped enterprise default",
						Required:      true,
						Type:          "string",
						AllowedValues: qualityPolicyRecordKinds,
					},
				},
			},
			{
				Name:        "bootstrap-drafts",
				Description: "Create enterprise-default drafts for every supported record kind without a tenant-scope policy (never publishes)",
				ToolName:    "UteamupQualityPolicyBootstrapDrafts",
				HTTPMethod:  "POST",
				RESTPath:    "bootstrap-drafts",
				Flags: []FlagDef{
					qualityPolicyBootstrapIdempotencyFlag(),
					qualityAuditConfirmationFlag(),
				},
			},
			{
				Name:        "draft-create",
				Description: "Create one explicit policy draft from a complete schema-v2 request",
				ToolName:    "UteamupQualityPolicyDraftCreate",
				HTTPMethod:  "POST",
				Flags:       qualityAuditCreateMutationFlags("quality-policy"),
			},
			{
				Name:        "draft-update",
				Description: "Replace the complete authored content of one draft with optimistic concurrency",
				ToolName:    "UteamupQualityPolicyDraftUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{policyGuid}",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("policyGuid", "Policy public GUID"),
				},
				Flags: qualityAuditExistingMutationFlags("quality-policy", false),
			},
			{
				Name:        "submit-for-review",
				Description: "Submit one draft for independent review after explicit confirmation",
				ToolName:    "UteamupQualityPolicySubmitForReview",
				HTTPMethod:  "POST",
				RESTPath:    "{policyGuid}/submit-for-review",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("policyGuid", "Policy public GUID"),
				},
				Flags: qualityPolicyBodylessMutationFlags(),
			},
			{
				Name:        "return-to-draft",
				Description: "Return one in-review policy to Draft with a retained reason",
				ToolName:    "UteamupQualityPolicyReturnToDraft",
				HTTPMethod:  "POST",
				RESTPath:    "{policyGuid}/return-to-draft",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("policyGuid", "Policy public GUID"),
				},
				Flags: append(qualityPolicyBodylessMutationFlags(), qualityPolicyReasonFlag()),
			},
			{
				Name:        "publish",
				Description: "Publish one reviewed policy; the publisher must differ from the author, last editor and submitter",
				ToolName:    "UteamupQualityPolicyPublish",
				HTTPMethod:  "POST",
				RESTPath:    "{policyGuid}/publish",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("policyGuid", "Policy public GUID"),
				},
				Flags: qualityPolicyBodylessMutationFlags(),
			},
			{
				Name:        "supersede",
				Description: "Publish an approved successor and supersede the published predecessor atomically",
				ToolName:    "UteamupQualityPolicySupersede",
				HTTPMethod:  "POST",
				RESTPath:    "{predecessorPolicyGuid}/supersede",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("predecessorPolicyGuid", "Published predecessor policy public GUID"),
				},
				Flags: append(
					qualityPolicyBodylessMutationFlags(),
					FlagDef{
						Name:        "successor-policy-guid",
						BodyName:    "successorPolicyGuid",
						Description: "Approved successor policy public GUID",
						Required:    true,
						Type:        "non-empty-uuid",
					},
					qualityPolicyReasonFlag(),
					FlagDef{
						Name:        "predecessor-concurrency-token",
						Description: "Current opaque concurrency token from the predecessor policy response",
						Required:    true,
						Type:        "string",
						Sensitive:   true,
						HeaderName:  "X-Predecessor-If-Match",
						StrongETag:  true,
					},
				),
			},
			{
				Name:        "withdraw",
				Description: "Withdraw one published policy with a retained reason",
				ToolName:    "UteamupQualityPolicyWithdraw",
				HTTPMethod:  "POST",
				RESTPath:    "{policyGuid}/withdraw",
				Args: []ArgDef{
					qualityAuditPublicGUIDArgument("policyGuid", "Policy public GUID"),
				},
				Flags: append(qualityPolicyBodylessMutationFlags(), qualityPolicyReasonFlag()),
			},
		},
	})
}

// qualityPolicyBodylessMutationFlags covers lifecycle actions whose REST body is empty or a small
// reason object: idempotency + concurrency headers and the explicit confirmation gate.
func qualityPolicyBodylessMutationFlags() []FlagDef {
	return []FlagDef{
		qualityAuditIdempotencyFlag(),
		qualityAuditConcurrencyFlag("quality-policy"),
		qualityAuditConfirmationFlag(),
	}
}

func qualityPolicyReasonFlag() FlagDef {
	return FlagDef{
		Name:        "reason",
		BodyName:    "reason",
		Description: "Retained reason for the decision (1-2000 characters)",
		Required:    true,
		Type:        "string",
	}
}

// The bootstrap key is shorter than the generic 8-128 byte key because the server derives one
// key per record kind by appending ":{RecordKind}" and the longest kind name is 31 bytes.
func qualityPolicyBootstrapIdempotencyFlag() FlagDef {
	return FlagDef{
		Name:        "idempotency-key",
		Description: "Opaque 8-96 byte caller key reused only for the exact same bootstrap; each record kind derives its own key from it",
		Required:    true,
		Type:        "string",
		HeaderName:  "Idempotency-Key",
	}
}
