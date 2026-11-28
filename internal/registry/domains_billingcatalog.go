package registry

func init() {
	Register(&Domain{
		Name:        "billing-catalog",
		Aliases:     []string{"billing-catalog-governance", "price-catalogue"},
		Description: "Maker-checker queue for the governed billing price catalogue",
		APIPath:     "/api/billing/admin/governance",
		Actions: []Action{
			{
				Name:        "pending-approvals",
				Description: "List draft price books, price versions and rule packs awaiting a second administrator",
				ToolName:    "UteamupBillingCatalogApprovalsPending",
				RESTPath:    "pending-approvals",
			},
		},
	})
}
