package registry

// pricingRuleFlags is shared by create and update — the backend uses one
// PricingRuleCreateModel for both verbs (closed condition vocabulary).
func pricingRuleFlags() []FlagDef {
	return []FlagDef{
		{Name: "name", Description: "Rule name (required)", Required: true, Type: "string"},
		{Name: "is-active", Description: "Whether the rule participates in matching", Default: true, Type: "bool"},
		{Name: "priority", Description: "Match priority (lower wins; first match applies)", Default: 100, Type: "int"},
		{Name: "country-code", Description: "Condition: tenant country code (e.g. IS)", Type: "string"},
		{Name: "billing-cycle", Description: "Condition: billing cycle (e.g. monthly, yearly)", Type: "string"},
		{Name: "plan-sku", Description: "Condition: plan SKU", Type: "string"},
		{Name: "min-tenant-age-days", Description: "Condition: minimum tenant age in days", Type: "int"},
		{Name: "discount-percent", Description: "Discount percent applied when the rule matches (required)", Required: true, Type: "float"},
		{Name: "effective-from", Description: "Rule validity start (ISO 8601 datetime)", Type: "string"},
		{Name: "effective-to", Description: "Rule validity end (ISO 8601 datetime)", Type: "string"},
	}
}

func init() {
	Register(&Domain{
		Name:        "pricing-rule",
		Aliases:     []string{"pricing-rules", "pricingrule", "pricingrules"},
		Description: "Manage the pricing rule engine (priority-ordered, first-match discounts)",
		APIPath:     "/api/pricingrule",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List all pricing rules",
				ToolName:    "UteamupPricingRuleList",
			},
			{
				Name:        "get",
				Description: "Get a pricing rule by its stable GUID",
				ToolName:    "UteamupPricingRuleGet",
				RESTPath:    "by-guid/{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "Pricing rule GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "create",
				Description: "Create a pricing rule",
				ToolName:    "UteamupPricingRuleCreate",
				Flags:       pricingRuleFlags(),
			},
			{
				Name:        "update",
				Description: "Update a pricing rule by its stable GUID",
				ToolName:    "UteamupPricingRuleUpdate",
				RESTPath:    "by-guid/{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "Pricing rule GUID", Required: true, Type: "string"}},
				Flags:       pricingRuleFlags(),
			},
			{
				Name:        "delete",
				Description: "Delete a pricing rule by its stable GUID",
				ToolName:    "UteamupPricingRuleDelete",
				RESTPath:    "by-guid/{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "Pricing rule GUID", Required: true, Type: "string"}},
			},
		},
	})
}
