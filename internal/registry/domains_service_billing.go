package registry

func serviceBillingListFlags() []FlagDef {
	return []FlagDef{
		{Name: "agreement-guid", BodyName: "agreementGuid", Description: "Optional service agreement external GUID", Type: "string"},
		{Name: "status", BodyName: "status", Description: "Optional deterministic billing status", Type: "string"},
		{Name: "page", BodyName: "page", Description: "One-based page number", Default: 1, Type: "int"},
		{Name: "page-size", BodyName: "pageSize", Description: "Results per page, maximum 200", Default: 50, Type: "int"},
	}
}

func serviceBillingTransitionFlags(reasonRequired bool) []FlagDef {
	return []FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped transition idempotency UUID", Required: true, Type: "string"},
		{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Exact reviewed UpdatedAt timestamp", Required: true, Type: "string"},
		{Name: "reason", BodyName: "reason", Description: "Reviewed transition reason", Required: reasonRequired, Type: "string"},
	}
}

func serviceInvoiceTransitionFlags() []FlagDef {
	return []FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped transition idempotency UUID", Required: true, Type: "string"},
		{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Exact reviewed UpdatedAt timestamp", Required: true, Type: "string"},
		{Name: "occurred-at", BodyName: "occurredAt", Description: "Optional ISO-8601 transition timestamp", Type: "string"},
	}
}

func serviceInvoiceIssueFlags() []FlagDef {
	return []FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped issue idempotency UUID", Required: true, Type: "string"},
		{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Exact reviewed UpdatedAt timestamp", Required: true, Type: "string"},
		{Name: "invoice-number", BodyName: "invoiceNumber", Description: "Explicit customer invoice number", Required: true, Type: "string"},
		{Name: "issued-at", BodyName: "issuedAt", Description: "Optional ISO-8601 issue timestamp", Type: "string"},
	}
}

func serviceInvoiceCorrectionFlags() []FlagDef {
	return append(serviceInvoiceTransitionFlags(), FlagDef{
		Name: "reason", BodyName: "reason", Description: "Explicit correction reason", Required: true, Type: "string",
	})
}

func init() {
	Register(&Domain{
		Name:        "service-billing",
		Aliases:     []string{"service-billing-runs", "billing-run"},
		Description: "Inspect and explicitly manage operational service-billing evidence",
		APIPath:     "/api/service-billing-runs",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List deterministic billing totals, exceptions, freshness, and approval state",
				ToolName:    "UteamupServiceBillingRunList",
				Flags:       serviceBillingListFlags(),
			},
			{
				Name:        "get",
				Description: "Get one billing run with exact source lines and anomaly evidence",
				ToolName:    "UteamupServiceBillingRunGet",
				RESTPath:    "{runGuid}",
				Args: []ArgDef{
					{Name: "runGuid", Description: "Service billing run external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Generate a review-first run from approved field evidence",
				ToolName:    "UteamupServiceBillingRunCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped generation idempotency UUID", Required: true, Type: "string"},
					{Name: "agreement-guid", BodyName: "agreementGuid", Description: "Approved service agreement external GUID", Required: true, Type: "string"},
					{Name: "period-start", BodyName: "periodStart", Description: "Billing period start in ISO-8601 UTC", Required: true, Type: "string"},
					{Name: "period-end", BodyName: "periodEnd", Description: "Billing period end in ISO-8601 UTC", Required: true, Type: "string"},
				},
			},
			{
				Name:        "approve",
				Description: "Approve the exact reviewed sources and draft totals",
				ToolName:    "UteamupServiceBillingRunApprove",
				HTTPMethod:  "POST",
				RESTPath:    "{runGuid}/approve",
				Args: []ArgDef{
					{Name: "runGuid", Description: "Service billing run external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceBillingTransitionFlags(false),
			},
			{
				Name:        "cancel",
				Description: "Cancel a draft run while preserving its evidence",
				ToolName:    "UteamupServiceBillingRunCancel",
				HTTPMethod:  "POST",
				RESTPath:    "{runGuid}/cancel",
				Args: []ArgDef{
					{Name: "runGuid", Description: "Service billing run external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceBillingTransitionFlags(true),
			},
		},
	})

	Register(&Domain{
		Name:        "service-invoice",
		Aliases:     []string{"service-invoices", "operational-invoice"},
		Description: "Inspect, issue, and correct immutable operational service invoices",
		APIPath:     "/api/service-invoices",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List immutable service-invoice snapshots and states",
				ToolName:    "UteamupServiceInvoiceList",
				Flags:       serviceBillingListFlags(),
			},
			{
				Name:        "get",
				Description: "Get one invoice with exact source, rate, quantity, tax, and freshness evidence",
				ToolName:    "UteamupServiceInvoiceGet",
				RESTPath:    "{invoiceGuid}",
				Args: []ArgDef{
					{Name: "invoiceGuid", Description: "Service invoice external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "issue",
				Description: "Issue an approved invoice and lock its financial snapshot",
				ToolName:    "UteamupServiceInvoiceIssue",
				HTTPMethod:  "POST",
				RESTPath:    "{invoiceGuid}/issue",
				Args: []ArgDef{
					{Name: "invoiceGuid", Description: "Service invoice external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceInvoiceIssueFlags(),
			},
			{
				Name:        "send",
				Description: "Record that an issued invoice was sent externally",
				ToolName:    "UteamupServiceInvoiceSend",
				HTTPMethod:  "POST",
				RESTPath:    "{invoiceGuid}/send",
				Args: []ArgDef{
					{Name: "invoiceGuid", Description: "Service invoice external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceInvoiceTransitionFlags(),
			},
			{
				Name:        "paid",
				Description: "Record that payment was received and reconciled",
				ToolName:    "UteamupServiceInvoicePaid",
				HTTPMethod:  "POST",
				RESTPath:    "{invoiceGuid}/paid",
				Args: []ArgDef{
					{Name: "invoiceGuid", Description: "Service invoice external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceInvoiceTransitionFlags(),
			},
			{
				Name:        "void",
				Description: "Void an invoice while preserving its financial snapshot",
				ToolName:    "UteamupServiceInvoiceVoid",
				HTTPMethod:  "POST",
				RESTPath:    "{invoiceGuid}/void",
				Args: []ArgDef{
					{Name: "invoiceGuid", Description: "Service invoice external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceInvoiceCorrectionFlags(),
			},
			{
				Name:        "credit",
				Description: "Credit an invoice while preserving its financial snapshot",
				ToolName:    "UteamupServiceInvoiceCredit",
				HTTPMethod:  "POST",
				RESTPath:    "{invoiceGuid}/credit",
				Args: []ArgDef{
					{Name: "invoiceGuid", Description: "Service invoice external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceInvoiceCorrectionFlags(),
			},
		},
	})
}
