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
					{Name: "seller-tenant-guid", Description: "Show only this tenant seller's listings", Type: "string"},
					{Name: "wholesaler-guid", Description: "Show only this wholesaler's catalogue rows", Type: "string"},
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
				Name:        "requirement-draft-create",
				Description: "Create a private tenant-owned marketplace requirement draft",
				ToolName:    "UteamupMarketplaceRequirementCreateDraft",
				Flags: []FlagDef{
					{Name: "stock-item-guid", Description: "Optional tenant stock item GUID", Type: "uuid"},
					{Name: "item-name", Description: "Item name when no stock item GUID is supplied", Type: "string"},
					{Name: "item-type", Description: "Part | Tool | Chemical | Asset", Type: "string"},
					{Name: "requested-quantity", Description: "Exact requested quantity", Required: true, Type: "int"},
					{Name: "audience", Description: "Wholesalers | Tenants | Both", Required: true, Type: "string"},
					{Name: "target-unit-price", Description: "Optional target unit price", Type: "float"},
					{Name: "currency", Description: "Three-letter currency code", Default: "USD", Type: "string"},
					{Name: "needed-by-date", Description: "Optional needed-by timestamp", Type: "string"},
					{Name: "expires-at", Description: "Optional draft expiry timestamp", Type: "string"},
					{Name: "notes", Description: "Optional private owner notes", Type: "string"},
				},
			},
			{
				Name:        "requirement-publish",
				Description: "Publish one tenant-owned marketplace requirement",
				ToolName:    "UteamupMarketplaceRequirementPublish",
				Args: []ArgDef{
					{Name: "requirementGuid", Description: "Marketplace requirement GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "requirement-offers-compare",
				Description: "Compare current offers for one tenant-owned requirement",
				ToolName:    "UteamupMarketplaceRequirementOffersCompare",
				Args: []ArgDef{
					{Name: "requirementGuid", Description: "Marketplace requirement GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "requirement-offer-accept",
				Description: "Accept one explicitly selected current offer",
				ToolName:    "UteamupMarketplaceRequirementOfferAccept",
				Args: []ArgDef{
					{Name: "requirementGuid", Description: "Marketplace requirement GUID", Required: true, Type: "uuid"},
					{Name: "offerGuid", Description: "Selected offer GUID", Required: true, Type: "uuid"},
				},
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
			{
				Name:        "seller-scorecard",
				Description: "Get a seller's trust scorecard (rating, fulfillment, response time)",
				ToolName:    "UteamupMarketplaceSellerScorecard",
				Flags: []FlagDef{
					{Name: "seller-guid", Short: "s", Description: "Seller tenant GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "facets",
				Description: "Get item-type / condition / price-band facet counts for a search",
				ToolName:    "UteamupMarketplaceFacets",
				Flags: []FlagDef{
					{Name: "search", Description: "Optional free-text search to facet within", Type: "string"},
				},
			},
			{
				Name:        "buyer-reputation",
				Description: "Get a buyer's reputation (completed purchases, tenure, verified badge)",
				ToolName:    "UteamupMarketplaceBuyerReputation",
				Flags: []FlagDef{
					{Name: "buyer-guid", Short: "b", Description: "Buyer tenant GUID", Required: true, Type: "string"},
				},
			},
		},
	})
}
