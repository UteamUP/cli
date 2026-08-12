package registry

func init() {
	Register(&Domain{
		Name:        "billing",
		Aliases:     []string{"commerce", "tenant-billing"},
		Description: "Read the provider-neutral tenant billing context, catalog, agreements, orders, and legal invoices",
		APIPath:     "/api/billing",
		Actions: []Action{
			{
				Name:        "context",
				Description: "Get billing availability, currency, payment methods, and allowed actions",
				ToolName:    "UteamupBillingContextGet",
				HTTPMethod:  "GET",
				RESTPath:    "context",
			},
			{
				Name:        "profile",
				Description: "Get the default legal billing profile with a masked tax identifier",
				ToolName:    "UteamupBillingProfileGet",
				HTTPMethod:  "GET",
				RESTPath:    "profile",
			},
			{
				Name:        "catalog",
				Description: "List eligible and explicitly blocked billing catalog items",
				ToolName:    "UteamupBillingCatalogList",
				HTTPMethod:  "GET",
				RESTPath:    "catalog",
			},
			{
				Name:        "order",
				Description: "Get one durable billing order by public GUID",
				ToolName:    "UteamupBillingOrderGet",
				HTTPMethod:  "GET",
				RESTPath:    "orders/{orderGuid}",
				Args: []ArgDef{
					{Name: "orderGuid", Description: "Billing order public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "agreements",
				Description: "List recurring billing agreements and scheduled lifecycle state",
				ToolName:    "UteamupBillingAgreementsList",
				HTTPMethod:  "GET",
				RESTPath:    "agreements",
			},
			{
				Name:        "invoices",
				Description: "List immutable legal invoices and credit notes",
				ToolName:    "UteamupBillingInvoicesList",
				HTTPMethod:  "GET",
				RESTPath:    "invoices",
			},
			{
				Name:        "invoice",
				Description: "Get one immutable invoice or credit note by public GUID",
				ToolName:    "UteamupBillingInvoiceGet",
				HTTPMethod:  "GET",
				RESTPath:    "invoices/{invoiceGuid}",
				Args: []ArgDef{
					{Name: "invoiceGuid", Description: "Invoice or credit-note public GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
