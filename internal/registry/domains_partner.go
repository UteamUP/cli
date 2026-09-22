package registry

// Partner CLI surface — mirrors the Partner MCP tools (read-oriented).
// Every action but list names its REST route: an action without a RESTPath falls back to
// GET /api/partner, so it silently printed the partner list instead of its own data.

func init() {
	Register(&Domain{
		Name:        "partner",
		Aliases:     []string{"partners", "resellers", "reseller", "rs"},
		Description: "Read partner program data: partners, applications, managed tenants, earnings",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List partners (global admins)",
				ToolName:    "UteamupPartnerList",
			},
			{
				Name:        "get",
				Description: "Get a partner by GUID (global admins)",
				ToolName:    "UteamupPartnerGet",
				RESTPath:    "{guid}",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Partner GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "applications",
				Description: "List partner applications (global admins)",
				ToolName:    "UteamupPartnerApplicationsList",
				RESTPath:    "application",
			},
			{
				Name:        "tenants",
				Description: "List a partner's managed tenants (global admins)",
				ToolName:    "UteamupPartnerTenantsList",
				RESTPath:    "{partnerGuid}/tenant",
				Flags: []FlagDef{
					{Name: "partner-guid", Short: "r", Description: "Partner GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "earnings",
				Description: "List a partner's earnings ledger (global admins)",
				ToolName:    "UteamupPartnerEarningsList",
				RESTPath:    "{partnerGuid}/earning",
				Flags: []FlagDef{
					{Name: "partner-guid", Short: "r", Description: "Partner GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "program-defaults",
				Description: "Get the global partner-program defaults new partners inherit (global admins)",
				ToolName:    "UteamupPartnerProgramDefaultsGet",
				RESTPath:    "program-defaults",
			},
			// New actions — 2026-06 partner program overhaul
			{
				Name:        "application-get",
				Description: "Get your own partner application thread (applicant self-serve)",
				ToolName:    "UteamupPartnerMyApplicationGet",
				RESTPath:    "me/application",
			},
			{
				Name:        "checklist",
				Description: "Get the reviewer validation checklist for a partner application (global admins)",
				ToolName:    "UteamupPartnerApplicationChecksGet",
				RESTPath:    "application/{applicationGuid}/checks",
				Flags: []FlagDef{
					{Name: "application-guid", Short: "a", Description: "Application GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "meetings",
				Description: "List the meetings (Teams + calendar invite) scheduled on a partner application (global admins)",
				ToolName:    "UteamupPartnerApplicationMeetingsGet",
				RESTPath:    "application/{applicationGuid}/meetings",
				Flags: []FlagDef{
					{Name: "application-guid", Short: "a", Description: "Application GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "referral-codes",
				Description: "List your partner referral codes (self-serve portal)",
				ToolName:    "UteamupPartnerMyReferralCodesGet",
				RESTPath:    "me/referral-codes",
			},
			{
				Name:        "tenant-manager",
				Description: "Get the partner managing your current tenant (visible to any tenant member)",
				ToolName:    "UteamupPartnerMyTenantManagerGet",
				RESTPath:    "my-tenant-manager",
			},
		},
	})
}
