package registry

func init() {
	commonTerms := []FlagDef{
		{Name: "name", BodyName: "name", Description: "Entitlement display name", Required: true, Type: "string"},
		{Name: "unit", BodyName: "unit", Description: "laborMinutes, materialAmount, visits, travelMinutes, or quantity", Required: true, Type: "string"},
		{Name: "included-quantity", BodyName: "includedQuantity", Description: "Allowance available in each period", Required: true, Type: "float"},
		{Name: "period", BodyName: "period", Description: "agreement, calendarMonth, calendarQuarter, or calendarYear", Required: true, Type: "string"},
		{Name: "effective-from", BodyName: "effectiveFrom", Description: "Entitlement start in ISO-8601 UTC", Required: true, Type: "string"},
		{Name: "effective-to", BodyName: "effectiveTo", Description: "Optional entitlement end in ISO-8601 UTC", Type: "string"},
		{Name: "currency", BodyName: "currency", Description: "Three-letter ISO code required for materialAmount", Type: "string"},
		{Name: "notes", BodyName: "notes", Description: "Optional reviewed entitlement notes", Type: "string"},
	}

	createFlags := append([]FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped write idempotency UUID", Required: true, Type: "string"},
		{Name: "agreement-guid", BodyName: "agreementGuid", Description: "Service agreement external GUID", Required: true, Type: "string"},
	}, commonTerms...)
	createFlags = append(createFlags,
		FlagDef{Name: "is-active", BodyName: "isActive", Description: "Create as active; defaults to true", Default: true, Type: "bool"})

	updateFlags := append([]FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped write idempotency UUID", Required: true, Type: "string"},
	}, commonTerms...)
	updateFlags = append(updateFlags,
		FlagDef{Name: "is-active", BodyName: "isActive", Description: "Whether the entitlement remains active", Required: true, Type: "bool"},
		FlagDef{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Exact reviewed UpdatedAt timestamp", Required: true, Type: "string"})

	Register(&Domain{
		Name:        "service-entitlement",
		Aliases:     []string{"service-entitlements", "agreement-allowance"},
		Description: "Manage service-agreement allowances and immutable usage evidence",
		APIPath:     "/api/service-entitlements",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List current or historical period-aware entitlement balances",
				ToolName:    "UteamupServiceEntitlementList",
				Flags: []FlagDef{
					{Name: "agreement-guid", BodyName: "agreementGuid", Description: "Optional service agreement external GUID", Type: "string"},
					{Name: "as-of", BodyName: "asOf", Description: "Optional ISO-8601 balance timestamp", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get an entitlement and its immutable usage evidence by external GUID",
				ToolName:    "UteamupServiceEntitlementGet",
				RESTPath:    "{entitlementGuid}",
				Args: []ArgDef{
					{Name: "entitlementGuid", Description: "Service entitlement external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "as-of", BodyName: "asOf", Description: "Optional ISO-8601 balance timestamp", Type: "string"},
				},
			},
			{
				Name:        "create",
				Description: "Create reviewable terms on an existing service agreement",
				ToolName:    "UteamupServiceEntitlementCreate",
				HTTPMethod:  "POST",
				Flags:       createFlags,
			},
			{
				Name:        "update",
				Description: "Update unused reviewed terms and clear prior agreement approval",
				ToolName:    "UteamupServiceEntitlementUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{entitlementGuid}",
				Args: []ArgDef{
					{Name: "entitlementGuid", Description: "Service entitlement external GUID", Required: true, Type: "uuid"},
				},
				Flags: updateFlags,
			},
			{
				Name:        "usage-record",
				Description: "Record one deterministic source against an approved entitlement",
				ToolName:    "UteamupServiceEntitlementUsageRecord",
				HTTPMethod:  "POST",
				RESTPath:    "{entitlementGuid}/usage",
				Args: []ArgDef{
					{Name: "entitlementGuid", Description: "Service entitlement external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped write idempotency UUID", Required: true, Type: "string"},
					{Name: "source-type", BodyName: "sourceType", Description: "workorder, timeEntry, stockMovement, recurringCharge, or manualAdjustment", Required: true, Type: "string"},
					{Name: "source-guid", BodyName: "sourceGuid", Description: "Deterministic source external GUID", Required: true, Type: "string"},
					{Name: "quantity", BodyName: "quantity", Description: "Positive quantity consumed", Required: true, Type: "float"},
					{Name: "occurred-at", BodyName: "occurredAt", Description: "Usage occurrence in ISO-8601 UTC", Required: true, Type: "string"},
					{Name: "description", BodyName: "description", Description: "Optional usage evidence description", Type: "string"},
				},
			},
			{
				Name:        "usage-reverse",
				Description: "Restore allowance by appending an immutable signed reversal",
				ToolName:    "UteamupServiceEntitlementUsageReverse",
				HTTPMethod:  "POST",
				RESTPath:    "usage/{usageGuid}/reverse",
				Args: []ArgDef{
					{Name: "usageGuid", Description: "Service entitlement usage external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped write idempotency UUID", Required: true, Type: "string"},
					{Name: "reason", BodyName: "reason", Description: "Auditable reversal reason", Required: true, Type: "string"},
				},
			},
		},
	})
}
