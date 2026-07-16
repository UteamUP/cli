package registry

func init() {
	createFlags := []FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped write idempotency UUID", Required: true, Type: "string"},
		{Name: "contract-guid", BodyName: "contractGuid", Description: "Existing Contract external GUID", Required: true, Type: "string"},
		{Name: "customer-guid", BodyName: "customerGuid", Description: "Covered customer external GUID", Required: true, Type: "string"},
		{Name: "effective-from", BodyName: "effectiveFrom", Description: "Agreement start in ISO-8601 UTC", Required: true, Type: "string"},
		{Name: "effective-to", BodyName: "effectiveTo", Description: "Optional agreement end in ISO-8601 UTC", Type: "string"},
		{Name: "default-response-minutes", BodyName: "defaultResponseMinutes", Description: "Default response target in minutes", Required: true, Type: "int"},
		{Name: "default-resolution-minutes", BodyName: "defaultResolutionMinutes", Description: "Default resolution target in minutes", Required: true, Type: "int"},
		{Name: "pause-sla-outside-business-hours", BodyName: "pauseSlaOutsideBusinessHours", Description: "Pause SLA clocks outside configured business hours", Type: "bool"},
		{Name: "coverage-notes", BodyName: "coverageNotes", Description: "Optional reviewed scope notes", Type: "string"},
		{Name: "coverage-json", BodyName: "coverage", Description: "Path to a JSON array of GUID-only coverage rules", Type: "string", JSONFile: true},
	}

	updateFlags := []FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped write idempotency UUID", Required: true, Type: "string"},
		{Name: "customer-guid", BodyName: "customerGuid", Description: "Covered customer external GUID", Required: true, Type: "string"},
		{Name: "effective-from", BodyName: "effectiveFrom", Description: "Agreement start in ISO-8601 UTC", Required: true, Type: "string"},
		{Name: "effective-to", BodyName: "effectiveTo", Description: "Optional agreement end in ISO-8601 UTC", Type: "string"},
		{Name: "default-response-minutes", BodyName: "defaultResponseMinutes", Description: "Default response target in minutes", Required: true, Type: "int"},
		{Name: "default-resolution-minutes", BodyName: "defaultResolutionMinutes", Description: "Default resolution target in minutes", Required: true, Type: "int"},
		{Name: "pause-sla-outside-business-hours", BodyName: "pauseSlaOutsideBusinessHours", Description: "Pause SLA clocks outside configured business hours", Type: "bool"},
		{Name: "coverage-notes", BodyName: "coverageNotes", Description: "Optional reviewed scope notes", Type: "string"},
		{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Exact reviewed UpdatedAt timestamp", Required: true, Type: "string"},
		{Name: "coverage-json", BodyName: "coverage", Description: "Path to a JSON array of GUID-only coverage rules", Type: "string", JSONFile: true},
	}

	Register(&Domain{
		Name:        "service-agreement",
		Aliases:     []string{"service-agreements", "agreement-terms"},
		Description: "Manage operational service terms attached to existing contracts",
		APIPath:     "/api/service-agreements",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List service agreements and coverage evidence",
				ToolName:    "UteamupServiceAgreementList",
				Flags: []FlagDef{
					{Name: "active-only", BodyName: "activeOnly", Description: "Return only currently effective active-contract terms", Type: "bool"},
				},
			},
			{
				Name:        "get",
				Description: "Get a service agreement by external GUID",
				ToolName:    "UteamupServiceAgreementGet",
				RESTPath:    "{agreementGuid}",
				Args: []ArgDef{
					{Name: "agreementGuid", Description: "Service agreement external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create reviewable terms on an existing Contract",
				ToolName:    "UteamupServiceAgreementCreate",
				HTTPMethod:  "POST",
				Flags:       createFlags,
			},
			{
				Name:        "update",
				Description: "Update reviewed terms; the server clears prior approval",
				ToolName:    "UteamupServiceAgreementUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{agreementGuid}",
				Args: []ArgDef{
					{Name: "agreementGuid", Description: "Service agreement external GUID", Required: true, Type: "uuid"},
				},
				Flags: updateFlags,
			},
			{
				Name:        "approve",
				Description: "Approve the exact reviewed service agreement version",
				ToolName:    "UteamupServiceAgreementApprove",
				HTTPMethod:  "POST",
				RESTPath:    "{agreementGuid}/approve",
				Args: []ArgDef{
					{Name: "agreementGuid", Description: "Service agreement external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped write idempotency UUID", Required: true, Type: "string"},
					{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Exact reviewed UpdatedAt timestamp", Required: true, Type: "string"},
				},
			},
		},
	})
}
