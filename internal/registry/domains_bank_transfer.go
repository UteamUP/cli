package registry

func init() {
	Register(&Domain{
		Name:        "bank-transfer",
		Aliases:     []string{"bt", "billing-transfer"},
		Description: "Manage Icelandic bank transfer billing (invoices, subscriptions, reconciliation)",
		Actions: []Action{
			{
				Name:        "list-invoices",
				Description: "List bank transfer invoices (optionally filter by status)",
				ToolName:    "UteamupBankTransferListInvoices",
				Flags: []FlagDef{
					{Name: "status", Short: "s", Description: "Filter by status: Draft, Issued, Sent, Paid, Overdue, Cancelled, CreditNote", Type: "string"},
				},
			},
			{
				Name:        "get-invoice",
				Description: "Get bank transfer invoice details by ID",
				ToolName:    "UteamupBankTransferGetInvoice",
				Args:        []ArgDef{{Name: "id", Description: "Invoice ID", Required: true, Type: "int"}},
			},
			{
				Name:        "mark-paid",
				Description: "Mark a bank transfer invoice as paid (reconciliation)",
				ToolName:    "UteamupBankTransferMarkPaid",
				Args:        []ArgDef{{Name: "id", Description: "Invoice ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "amount", Short: "a", Description: "Payment amount in ISK", Required: true, Type: "float"},
					{Name: "reference", Short: "r", Description: "Bank statement reference", Type: "string"},
				},
			},
			{
				Name:        "list-subscriptions",
				Description: "List all bank transfer subscriptions",
				ToolName:    "UteamupBankTransferListSubscriptions",
			},
			{
				Name:        "activate",
				Description: "Activate a pending bank transfer subscription",
				ToolName:    "UteamupBankTransferActivateSubscription",
				Args:        []ArgDef{{Name: "id", Description: "Subscription ID", Required: true, Type: "int"}},
			},
			{
				Name:        "dashboard",
				Description: "View bank transfer billing dashboard (outstanding, overdue, active stats)",
				ToolName:    "UteamupBankTransferDashboard",
			},
		},
	})
}
