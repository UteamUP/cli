package registry

// Promotions / discounts CLI surface — manage discount offers and grant them to tenants from the
// terminal. Mirrors the backend PromotionController (GUID-first per Guidelines/ApiGuidelines.md):
//
//   list          GET    /api/promotion
//   get           GET    /api/promotion/by-guid/{guid}
//   create        POST   /api/promotion                              (body = flags, camelCased)
//   update        PUT    /api/promotion/by-guid/{guid}
//   archive       DELETE /api/promotion/by-guid/{guid}
//   grant         POST   /api/promotion/by-guid/{guid}/grant         (grant an existing offer to a tenant)
//   grant-adhoc   POST   /api/promotion/grant                        (create an AdminGrant offer + grant it)
//   redemptions   GET    /api/promotion/by-guid/{guid}/redemptions
//   revoke        DELETE /api/promotion/redemption/by-guid/{guid}
//
// The CLI calls these REST routes directly (CallREST); the ToolName is the MCP mirror declaration.

func init() {
	Register(&Domain{
		Name:        "promotion",
		Aliases:     []string{"promotions", "discount", "discounts"},
		Description: "Manage promotions / discounts and grant them to tenants",
		APIPath:     "/api/promotion",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List all promotions",
				ToolName:    "UteamupPromotionList",
			},
			{
				Name:        "get",
				Description: "Get a promotion by its stable GUID",
				ToolName:    "UteamupPromotionGet",
				RESTPath:    "by-guid/{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "Promotion GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "create",
				Description: "Create a promotion offer. Enums are integers: audience 0=AdminGrant 1=PublicCode 2=Referral; discount-type 0=Percentage 1=FixedAmount 2=FreeMonths; status 0=Draft 1=Active 2=Paused 4=Archived.",
				ToolName:    "UteamupPromotionCreate",
				Flags: []FlagDef{
					{Name: "name", Description: "Promotion name (required)", Required: true, Type: "string"},
					{Name: "description", Description: "Customer-facing description", Type: "string"},
					{Name: "audience", Description: "Audience: 0=AdminGrant, 1=PublicCode, 2=Referral", Default: 0, Type: "int"},
					{Name: "status", Description: "Status: 0=Draft, 1=Active, 2=Paused, 4=Archived", Default: 1, Type: "int"},
					{Name: "code", Description: "Redemption code (public/referral offers)", Type: "string"},
					{Name: "target-plan-guid", Description: "Plan GUID to upgrade the tenant to (the 'bump to' plan)", Type: "string"},
					{Name: "discount-type", Description: "Discount type: 0=Percentage, 1=FixedAmount, 2=FreeMonths", Default: 0, Type: "int"},
					{Name: "discount-value", Description: "Percent (0-100), fixed amount, or month count", Type: "float"},
					{Name: "duration-kind", Description: "Duration: 0=Once (one invoice), 1=Repeating, 2=Forever; FreeMonths uses Repeating", Type: "int"},
					{Name: "duration-months", Description: "How many months the discount runs", Type: "int"},
					{Name: "currency", Description: "Currency for fixed-amount discounts (USD/EUR/ISK)", Type: "string"},
				},
			},
			{
				Name:        "update",
				Description: "Update a promotion offer by its stable GUID",
				ToolName:    "UteamupPromotionUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "Promotion GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "name", Description: "Promotion name (required)", Required: true, Type: "string"},
					{Name: "description", Description: "Customer-facing description", Type: "string"},
					{Name: "audience", Description: "Audience: 0=AdminGrant, 1=PublicCode, 2=Referral", Default: 0, Type: "int"},
					{Name: "status", Description: "Status: 0=Draft, 1=Active, 2=Paused, 4=Archived", Default: 1, Type: "int"},
					{Name: "target-plan-guid", Description: "Plan GUID to upgrade the tenant to", Type: "string"},
					{Name: "discount-type", Description: "Discount type: 0=Percentage, 1=FixedAmount, 2=FreeMonths", Default: 0, Type: "int"},
					{Name: "discount-value", Description: "Percent (0-100), fixed amount, or month count", Type: "float"},
					{Name: "duration-kind", Description: "Duration: 0=Once (one invoice), 1=Repeating, 2=Forever; FreeMonths uses Repeating", Type: "int"},
					{Name: "duration-months", Description: "How many months the discount runs", Type: "int"},
					{Name: "currency", Description: "Currency for fixed-amount discounts (USD/EUR/ISK)", Type: "string"},
				},
			},
			{
				Name:        "archive",
				Description: "Archive a promotion offer by its stable GUID",
				ToolName:    "UteamupPromotionArchive",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "Promotion GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "grant",
				Description: "Grant an existing promotion to one or more tenants (optionally bumping the plan); partial success is reported per tenant",
				ToolName:    "UteamupPromotionGrant",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{guid}/grant",
				Args:        []ArgDef{{Name: "guid", Description: "Promotion GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "tenant-guid", Description: "Tenant GUID to grant the discount to", Type: "string"},
					{Name: "tenant-guids", BodyName: "tenantGuids", Description: "Additional tenant GUIDs for a bulk grant — repeatable or comma-separated", Type: "stringSlice"},
					{Name: "target-plan-guid", Description: "Override the promotion's target plan for this grant", Type: "string"},
					{Name: "billing-cycle", Description: "Billing cycle when bumping the plan: monthly | annual", Type: "string"},
				},
			},
			{
				Name:        "grant-adhoc",
				Description: "One-shot admin grant: create an AdminGrant offer and grant it to one or more tenants in one step",
				ToolName:    "UteamupPromotionGrantAdhoc",
				HTTPMethod:  "POST",
				RESTPath:    "grant",
				Flags: []FlagDef{
					{Name: "tenant-guid", Description: "Tenant GUID to grant the discount to", Type: "string"},
					{Name: "tenant-guids", BodyName: "tenantGuids", Description: "Additional tenant GUIDs for a bulk grant — repeatable or comma-separated", Type: "stringSlice"},
					{Name: "target-plan-guid", Description: "Plan GUID to upgrade the tenant to", Type: "string"},
					{Name: "billing-cycle", Description: "Billing cycle when bumping the plan: monthly | annual", Type: "string"},
					{Name: "discount-type", Description: "Discount type: 0=Percentage, 1=FixedAmount, 2=FreeMonths", Default: 0, Type: "int"},
					{Name: "discount-value", Description: "Percent (0-100), fixed amount, or month count", Type: "float"},
					{Name: "duration-kind", Description: "Duration: 0=Once (one invoice), 1=Repeating, 2=Forever; FreeMonths uses Repeating", Type: "int"},
					{Name: "duration-months", Description: "How many months the discount runs", Type: "int"},
					{Name: "currency", Description: "Currency for fixed-amount discounts (USD/EUR/ISK)", Type: "string"},
					{Name: "name", Description: "Optional label for the generated promotion", Type: "string"},
				},
			},
			{
				Name:        "redemptions",
				Description: "List a promotion's redemptions",
				ToolName:    "UteamupPromotionRedemptions",
				RESTPath:    "by-guid/{guid}/redemptions",
				Args:        []ArgDef{{Name: "guid", Description: "Promotion GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "revoke",
				Description: "Revoke a tenant's redemption by its stable GUID (removes the discount from the next invoice)",
				ToolName:    "UteamupPromotionRevokeRedemption",
				HTTPMethod:  "DELETE",
				RESTPath:    "redemption/by-guid/{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "Redemption GUID", Required: true, Type: "string"}},
			},
		},
	})
}
