package registry

// Reseller CLI surface — mirrors the Reseller MCP tools (read-oriented).

func init() {
	Register(&Domain{
		Name:        "reseller",
		Aliases:     []string{"resellers", "rs"},
		Description: "Read reseller program data: resellers, applications, managed tenants, earnings",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List resellers",
				ToolName:    "UteamupResellerList",
			},
			{
				Name:        "get",
				Description: "Get a reseller by GUID",
				ToolName:    "UteamupResellerGet",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Reseller GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "applications",
				Description: "List reseller applications",
				ToolName:    "UteamupResellerApplicationsList",
			},
			{
				Name:        "tenants",
				Description: "List a reseller's managed tenants",
				ToolName:    "UteamupResellerTenantsList",
				Flags: []FlagDef{
					{Name: "reseller-guid", Short: "r", Description: "Reseller GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "earnings",
				Description: "List a reseller's earnings ledger",
				ToolName:    "UteamupResellerEarningsList",
				Flags: []FlagDef{
					{Name: "reseller-guid", Short: "r", Description: "Reseller GUID", Required: true, Type: "string"},
				},
			},
		},
	})
}
