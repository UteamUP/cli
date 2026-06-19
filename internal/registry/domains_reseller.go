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
			{
				Name:        "program-defaults",
				Description: "Get the global reseller-program defaults new resellers inherit",
				ToolName:    "UteamupResellerProgramDefaultsGet",
			},
			// New actions — 2026-06 reseller program overhaul
			{
				Name:        "application-get",
				Description: "Get your own reseller application thread (applicant self-serve)",
				ToolName:    "UteamupResellerMyApplicationGet",
			},
			{
				Name:        "checklist",
				Description: "Get the reviewer validation checklist for a reseller application",
				ToolName:    "UteamupResellerApplicationChecksGet",
				Flags: []FlagDef{
					{Name: "application-guid", Short: "a", Description: "Application GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "referral-codes",
				Description: "List your reseller referral codes (self-serve portal)",
				ToolName:    "UteamupResellerMyReferralCodesGet",
			},
			{
				Name:        "tenant-manager",
				Description: "Get the reseller managing your current tenant (visible to any tenant member)",
				ToolName:    "UteamupResellerMyTenantManagerGet",
			},
		},
	})
}
