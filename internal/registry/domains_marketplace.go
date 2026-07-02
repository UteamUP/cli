package registry

// Marketplace CLI surface — mirrors the Marketplace MCP tools (read-oriented plus
// offer submission via the portal tools; identity comes from the Bearer token).

func init() {
	Register(&Domain{
		Name:        "marketplace",
		Aliases:     []string{"mp"},
		Description: "Browse the cross-tenant marketplace: listings, requirements, offers, transactions",
		Actions: []Action{
			{
				Name:        "browse",
				Description: "Browse cross-tenant listings (tenant listings + wholesaler catalogs)",
				ToolName:    "UteamupMarketplaceBrowse",
				Flags: []FlagDef{
					{Name: "search", Description: "Search term", Type: "string"},
					{Name: "item-type", Description: "Part | Tool | Chemical | Asset | Labor", Type: "string"},
					{Name: "condition", Description: "New | LikeNew | Good | Fair | ForParts", Type: "string"},
					{Name: "source-kind", Description: "TenantListing | WholesalerCatalog", Type: "string"},
					{Name: "page", Description: "Page number", Type: "float", Default: 1.0},
					{Name: "page-size", Description: "Page size", Type: "float", Default: 20.0},
				},
			},
			{
				Name:        "listing-get",
				Description: "Get one cross-tenant listing by GUID",
				ToolName:    "UteamupMarketplaceListingGet",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Listing GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "requirements",
				Description: "List open anonymous stock requirements visible to your tenant",
				ToolName:    "UteamupMarketplaceRequirementsList",
			},
			{
				Name:        "my-offers",
				Description: "List your tenant's offers on requirements",
				ToolName:    "UteamupMarketplaceMyOffersList",
			},
			{
				Name:        "transactions",
				Description: "List your tenant's marketplace transactions",
				ToolName:    "UteamupMarketplaceTransactionsList",
				Flags: []FlagDef{
					{Name: "page", Description: "Page number", Type: "float", Default: 1.0},
					{Name: "page-size", Description: "Page size", Type: "float", Default: 20.0},
				},
			},
			{
				Name:        "settings",
				Description: "Get your tenant's marketplace settings",
				ToolName:    "UteamupMarketplaceSettingsGet",
			},
		},
	})
}
