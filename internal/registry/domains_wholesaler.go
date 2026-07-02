package registry

// Wholesaler CLI surface — mirrors the Wholesaler MCP tools (read-oriented; the
// goods-supplier program, distinct from the license reseller program).

func init() {
	Register(&Domain{
		Name:        "wholesaler",
		Aliases:     []string{"wholesalers", "whl"},
		Description: "Read wholesaler program data: wholesalers, applications, catalogs",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List wholesalers",
				ToolName:    "UteamupWholesalerList",
				Flags: []FlagDef{
					{Name: "status", Description: "PendingApproval | Active | Suspended | Deactivated", Type: "string"},
					{Name: "search", Description: "Search by name or contact email", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get a wholesaler by GUID",
				ToolName:    "UteamupWholesalerGet",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Wholesaler GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "applications",
				Description: "List wholesaler applications",
				ToolName:    "UteamupWholesalerApplicationsList",
				Flags: []FlagDef{
					{Name: "status", Description: "PendingApproval | InfoRequested | Approved | Rejected | Withdrawn", Type: "string"},
				},
			},
			{
				Name:        "catalog",
				Description: "Get a wholesaler's catalog",
				ToolName:    "UteamupWholesalerCatalogGet",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Wholesaler GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "me",
				Description: "Get your own wholesaler (portal self-view)",
				ToolName:    "UteamupWholesalerMyGet",
			},
		},
	})
}
