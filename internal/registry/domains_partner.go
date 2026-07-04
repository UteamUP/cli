package registry

// Partner CLI surface — mirrors the Partner MCP tools (read-oriented).

func init() {
	Register(&Domain{
		Name:        "partner",
		Aliases:     []string{"partners", "resellers", "reseller", "rs"},
		Description: "Read partner program data: partners, applications, managed tenants, earnings",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List partners",
				ToolName:    "UteamupPartnerList",
			},
			{
				Name:        "get",
				Description: "Get a partner by GUID",
				ToolName:    "UteamupPartnerGet",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Partner GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "applications",
				Description: "List partner applications",
				ToolName:    "UteamupPartnerApplicationsList",
			},
			{
				Name:        "tenants",
				Description: "List a partner's managed tenants",
				ToolName:    "UteamupPartnerTenantsList",
				Flags: []FlagDef{
					{Name: "partner-guid", Short: "r", Description: "Partner GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "earnings",
				Description: "List a partner's earnings ledger",
				ToolName:    "UteamupPartnerEarningsList",
				Flags: []FlagDef{
					{Name: "partner-guid", Short: "r", Description: "Partner GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "program-defaults",
				Description: "Get the global partner-program defaults new partners inherit",
				ToolName:    "UteamupPartnerProgramDefaultsGet",
			},
			// New actions — 2026-06 partner program overhaul
			{
				Name:        "application-get",
				Description: "Get your own partner application thread (applicant self-serve)",
				ToolName:    "UteamupPartnerMyApplicationGet",
			},
			{
				Name:        "checklist",
				Description: "Get the reviewer validation checklist for a partner application",
				ToolName:    "UteamupPartnerApplicationChecksGet",
				Flags: []FlagDef{
					{Name: "application-guid", Short: "a", Description: "Application GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "meetings",
				Description: "List the meetings (Teams + calendar invite) scheduled on a partner application",
				ToolName:    "UteamupPartnerApplicationMeetingsGet",
				Flags: []FlagDef{
					{Name: "application-guid", Short: "a", Description: "Application GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "referral-codes",
				Description: "List your partner referral codes (self-serve portal)",
				ToolName:    "UteamupPartnerMyReferralCodesGet",
			},
			{
				Name:        "tenant-manager",
				Description: "Get the partner managing your current tenant (visible to any tenant member)",
				ToolName:    "UteamupPartnerMyTenantManagerGet",
			},
		},
	})
}
