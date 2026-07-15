package registry

// AI-credit grants CLI surface — issue claim-gated AI credits to a tenant and let the tenant list,
// claim, or revoke them from the terminal. Mirrors the backend AiCreditGrantController (GUID-first per
// Guidelines/ApiGuidelines.md):
//
//   issue    POST /api/aicreditgrant                 (body = flags, camelCased; global-admin only)
//   mine     GET  /api/aicreditgrant/mine            (tenant's pending claimable grants)
//   claim    POST /api/aicreditgrant/{guid}/claim
//   revoke   POST /api/aicreditgrant/{guid}/revoke
//
// The CLI calls these REST routes directly (CallREST); the ToolName is the MCP mirror declaration.

func init() {
	Register(&Domain{
		Name:        "aicreditgrant",
		Aliases:     []string{"aicreditgrants", "aicredit", "aicredits"},
		Description: "Issue, list, claim, and revoke AI-credit grants",
		APIPath:     "/api/aicreditgrant",
		Actions: []Action{
			{
				Name:        "issue",
				Description: "Issue AI credits to a tenant (auto-applied, or claim-gated when --requires-claim is set). Global-admin only.",
				ToolName:    "UteamupAiCreditGrantIssue",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "tenant-guid", Description: "Tenant GUID to grant the credits to (required)", Required: true, Type: "string"},
					{Name: "amount", Description: "Number of AI credits to grant (required)", Required: true, Type: "int"},
					{Name: "requires-claim", Description: "Require the tenant owner to claim the grant before it applies", Type: "bool"},
					{Name: "source", Description: "Optional source label for the grant", Type: "string"},
					{Name: "expires-after-days", Description: "Expiry (days from now) for an unclaimed grant", Type: "int"},
				},
			},
			{
				Name:        "mine",
				Description: "List the current tenant's pending, unexpired AI-credit grants",
				ToolName:    "UteamupAiCreditGrantMine",
				RESTPath:    "mine",
			},
			{
				Name:        "claim",
				Description: "Claim a pending grant for the current tenant by its stable GUID",
				ToolName:    "UteamupAiCreditGrantClaim",
				HTTPMethod:  "POST",
				RESTPath:    "{guid}/claim",
				Args:        []ArgDef{{Name: "guid", Description: "Grant GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "revoke",
				Description: "Revoke a grant by its stable GUID (clawback, reducing the balance if already claimed)",
				ToolName:    "UteamupAiCreditGrantRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "{guid}/revoke",
				Args:        []ArgDef{{Name: "guid", Description: "Grant GUID", Required: true, Type: "string"}},
			},
		},
	})
}
