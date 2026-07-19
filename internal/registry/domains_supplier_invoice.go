package registry

func init() {
	Register(&Domain{
		Name:        "supplier-invoice",
		Aliases:     []string{"supplier-invoices", "invoice-match"},
		Description: "Capture and review supplier-invoice evidence against stock purchase orders",
		APIPath:     "/api/stock/supplier-invoices",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List tenant supplier invoices and their current match status",
				ToolName:    "UteamupStockListSupplierInvoices",
				Flags: []FlagDef{
					{Name: "page", BodyName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", BodyName: "pageSize", Description: "Results per page, maximum 200", Default: 50, Type: "int"},
					{Name: "match-status", BodyName: "matchStatus", Description: "Optional Unmatched, NeedsReview, or Matched filter", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get one supplier invoice and its captured source lines",
				ToolName:    "UteamupStockGetSupplierInvoice",
				RESTPath:    "{invoiceGuid}",
				Args: []ArgDef{
					{Name: "invoiceGuid", Description: "Supplier invoice GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Capture supplier-invoice source evidence without matching the purchase order",
				ToolName:    "UteamupStockCreateSupplierInvoice",
				HTTPMethod:  "POST",
				Flags:       []FlagDef{jsonFlag()},
			},
			{
				Name:        "match-preview",
				Description: "Review deterministic ordered, received, invoiced, price, tax, freight, currency, and duplicate evidence",
				ToolName:    "UteamupStockPreviewSupplierInvoiceMatch",
				RESTPath:    "{invoiceGuid}/match-preview",
				Args: []ArgDef{
					{Name: "invoiceGuid", Description: "Supplier invoice GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "match-prepare",
				Description: "Revalidate evidence and prepare an approval-bound UPMate durable run without applying the match",
				ToolName:    "UteamupStockPrepareSupplierInvoiceMatchRun",
				HTTPMethod:  "POST",
				RESTPath:    "{invoiceGuid}/match-runs",
				Args: []ArgDef{
					{Name: "invoiceGuid", Description: "Supplier invoice GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "idempotency-guid", BodyName: "idempotencyGuid", Description: "Caller-generated retry-safe GUID", Required: true, Type: "uuid"},
					{Name: "conversation-guid", BodyName: "conversationGuid", Description: "Optional UPMate conversation GUID", Type: "uuid"},
				},
			},
		},
	})
}
