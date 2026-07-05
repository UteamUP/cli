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
					{Name: "near-latitude", Description: "Latitude for a 'tenants near me' sort (pass with --near-longitude)", Type: "float"},
					{Name: "near-longitude", Description: "Longitude for a 'tenants near me' sort (pass with --near-latitude)", Type: "float"},
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
				Name:        "listing-report",
				Description: "Report a listing for moderation review",
				ToolName:    "UteamupMarketplaceListingReport",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Listing GUID", Required: true, Type: "string"},
					{Name: "reason", Short: "r", Description: "Illegal | Prohibited | Misleading | Offensive | WrongCategory | Other", Required: true, Type: "string"},
					{Name: "details", Description: "Optional details for the moderation team", Type: "string"},
				},
			},
			{
				Name:        "messages-list",
				Description: "List buyer↔seller message threads on a listing or a transaction",
				ToolName:    "UteamupMarketplaceMessagesList",
				Flags: []FlagDef{
					{Name: "listing-guid", Description: "Listing GUID (pass this OR --transaction-guid)", Type: "string"},
					{Name: "transaction-guid", Description: "Transaction GUID (pass this OR --listing-guid)", Type: "string"},
					{Name: "page", Description: "Page number", Type: "float", Default: 1.0},
					{Name: "page-size", Description: "Page size", Type: "float", Default: 20.0},
				},
			},
			{
				Name:        "message-send",
				Description: "Send a buyer↔seller message on a listing or a transaction",
				ToolName:    "UteamupMarketplaceMessageSend",
				Flags: []FlagDef{
					{Name: "listing-guid", Description: "Listing GUID (for a listing thread)", Type: "string"},
					{Name: "transaction-guid", Description: "Transaction GUID (for a transaction thread)", Type: "string"},
					{Name: "parent-guid", Description: "Parent message GUID to reply to", Type: "string"},
					{Name: "body", Short: "b", Description: "Message body", Required: true, Type: "string"},
				},
			},
			{
				Name:        "message-thread",
				Description: "Get one message thread (root + replies) by root message GUID",
				ToolName:    "UteamupMarketplaceMessageThreadGet",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Root message GUID", Required: true, Type: "string"},
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
			{
				Name:        "saved-searches",
				Description: "List your saved marketplace searches",
				ToolName:    "UteamupMarketplaceSavedSearchesList",
			},
			{
				Name:        "save-search",
				Description: "Save a browse filter as a search, optionally notifying you on new matches",
				ToolName:    "UteamupMarketplaceSaveSearch",
				Flags: []FlagDef{
					{Name: "name", Short: "n", Description: "A short name for the saved search", Required: true, Type: "string"},
					{Name: "filters-json", Description: `Browse filter as JSON, e.g. {"itemType":"Tool","maxPrice":200}`, Type: "string", Default: "{}"},
					{Name: "notify-on-new-match", Description: "Notify me when a new listing matches", Type: "bool", Default: true},
				},
			},
			{
				Name:        "delete-saved-search",
				Description: "Delete one of your saved searches",
				ToolName:    "UteamupMarketplaceDeleteSavedSearch",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Saved search GUID", Required: true, Type: "string"},
				},
			},
		},
	})
}
