package registry

func init() {
	Register(&Domain{
		Name:        "bank-transfer",
		Aliases:     []string{"bt", "billing-transfer"},
		Description: "Manage Icelandic bank transfer billing (invoices, subscriptions, reconciliation, refunds)",
		APIPath:     "/api/internalbilling",
		Actions: []Action{
			{
				Name:        "list-invoices",
				Description: "List pending (unsettled) bank transfer invoices",
				ToolName:    "UteamupBankTransferListInvoices",
				HTTPMethod:  "GET",
				RESTPath:    "admin/invoices/pending",
			},
			{
				Name:        "list-overdue",
				Description: "List overdue bank transfer invoices",
				ToolName:    "UteamupBankTransferListOverdueInvoices",
				HTTPMethod:  "GET",
				RESTPath:    "admin/invoices/overdue",
			},
			{
				Name:        "list-paid",
				Description: "List recently settled bank transfer invoices",
				ToolName:    "UteamupBankTransferListPaidInvoices",
				HTTPMethod:  "GET",
				RESTPath:    "admin/invoices/paid",
			},
			{
				Name:        "get-invoice",
				Description: "Get bank transfer invoice details by external GUID",
				ToolName:    "UteamupBankTransferGetInvoice",
				HTTPMethod:  "GET",
				RESTPath:    "invoices/{invoiceGuid}",
				Args:        []ArgDef{{Name: "invoiceGuid", Description: "Invoice external GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "mark-paid",
				Description: "Mark a bank transfer invoice as paid (reconciliation)",
				ToolName:    "UteamupBankTransferMarkPaid",
				HTTPMethod:  "POST",
				RESTPath:    "admin/invoices/{invoiceGuid}/mark-paid",
				Args:        []ArgDef{{Name: "invoiceGuid", Description: "Invoice external GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "amount", Short: "a", BodyName: "amount", Description: "Payment amount in ISK", Required: true, Type: "float"},
					{Name: "payment-date", BodyName: "paymentDate", Description: "ISO-8601 date the payment landed on the bank statement", Required: true, Type: "string"},
					{Name: "reference", Short: "r", BodyName: "bankReference", Description: "Bank statement reference", Type: "string"},
					{Name: "notes", BodyName: "adminNotes", Description: "Admin reconciliation notes", Type: "string"},
				},
			},
			{
				Name:        "status",
				Description: "Show the caller's own tenant billing overview (internal subscription status)",
				ToolName:    "UteamupBankTransferBillingOverview",
				HTTPMethod:  "GET",
				RESTPath:    "subscription-status",
			},
			{
				Name:        "list-subscriptions",
				Description: "List all bank transfer subscriptions",
				ToolName:    "UteamupBankTransferListSubscriptions",
				HTTPMethod:  "GET",
				RESTPath:    "admin/subscriptions",
			},
			{
				Name:        "activate",
				Description: "Activate a pending bank transfer subscription",
				ToolName:    "UteamupBankTransferActivateSubscription",
				HTTPMethod:  "POST",
				RESTPath:    "admin/subscriptions/{subscriptionGuid}/activate",
				Args:        []ArgDef{{Name: "subscriptionGuid", Description: "Subscription external GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "refund",
				Description: "Refund a settled provider payment (Kling card intents) and issue the paired credit note",
				ToolName:    "UteamupBankTransferRefundPayment",
				HTTPMethod:  "POST",
				RESTPath:    "admin/payments/{paymentGuid}/refund",
				Args:        []ArgDef{{Name: "paymentGuid", Description: "Payment external GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "reason", Short: "r", BodyName: "reason", Description: "Refund reason (max 500 characters)", Required: true, Type: "string"},
					{Name: "amount", Short: "a", BodyName: "amount", Description: "Partial refund amount; omit to refund the full remaining refundable amount", Type: "float"},
				},
			},
			{
				Name:        "dashboard",
				Description: "View bank transfer billing dashboard (outstanding, overdue, active stats)",
				ToolName:    "UteamupBankTransferDashboard",
				HTTPMethod:  "GET",
				RESTPath:    "admin/dashboard",
			},
		},
	})
}
