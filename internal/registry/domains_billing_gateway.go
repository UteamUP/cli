package registry

// Per-tenant billing-gateway switcher — CLI parity for the MCP tools
// registered in UteamUP_Backend/UteamUP_API/MCP/Tools/BillingGatewaySwitchTools.cs.
//
// Every action is globaladmin-only: the backend gate checks GlobalAdminEmails
// AND ApplicationUser.EmailConfirmed = true. A CLI caller whose email is not
// in the config list will get an UnauthorizedAccessException from the MCP proxy
// that surfaces as a 4xx / non-zero exit code.
//
// Command shape:
//   ut admin-billing-gateway change --tenant <guid> --to <stripe|ibt> --reason "..." [--kennitala <10-digits>] [--effective <cycle|immediate>] [--idempotency-key <key>]
//   ut admin-billing-gateway history --tenant <guid> [--page <n>] [--page-size <n>]
//   ut admin-billing-gateway get --tenant <guid> --audit <guid>
//   ut admin-billing-gateway cancel --tenant <guid> --audit <guid> --reason "..."
//
// The --to flag accepts the human-readable aliases "stripe" / "ibt" / "kling"
// and is translated to the backend's BillingMethod enum (0 / 1) at call time.
// --effective similarly accepts "cycle" (EndOfCurrentCycle = 0) / "immediate"
// (StartImmediately = 1).

func init() {
	Register(&Domain{
		Name:        "admin-billing-gateway",
		Aliases:     []string{"abg", "billing-gateway", "gateway"},
		Description: "Globaladmin-only per-tenant billing-gateway switcher (Stripe <-> IcelandicBankTransfer)",
		Actions: []Action{
			{
				Name:        "change",
				Description: "Request a billing-gateway change for a tenant. Returns a Kling checkout URL that the tenant owner must complete.",
				ToolName:    "AdminChangeTenantBillingMethod",
				Flags: []FlagDef{
					{Name: "tenant", Short: "t", Description: "Public GUID of the target tenant", Required: true, Type: "string"},
					{Name: "to", Description: "Target billing method: stripe, ibt (IcelandicBankTransfer), or kling (alias for ibt)", Required: true, Type: "string"},
					{Name: "reason", Short: "r", Description: "Globaladmin-authored reason, 10-1000 chars. Stored verbatim on the 7-year audit log.", Required: true, Type: "string"},
					{Name: "kennitala", Short: "k", Description: "10-digit Icelandic company kennitala. Required when --to = ibt / kling.", Type: "string"},
					{Name: "effective", Description: "When the new subscription starts: cycle (end of current billing cycle, default) or immediate", Type: "string"},
					{Name: "idempotency-key", Description: "Optional idempotency key. If omitted, the server generates one. Uniqueness is scoped to (key, tenant).", Type: "string"},
				},
			},
			{
				Name:        "history",
				Description: "Paginated audit log of billing-gateway changes for a tenant (reason + next-action URL excluded from list).",
				ToolName:    "AdminGetTenantBillingHistory",
				Flags: []FlagDef{
					{Name: "tenant", Short: "t", Description: "Public GUID of the target tenant", Required: true, Type: "string"},
					{Name: "page", Short: "p", Description: "1-based page number. Default 1.", Type: "int"},
					{Name: "page-size", Short: "s", Description: "Page size, clamped to [1, 100]. Default 20.", Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Fetch a single audit row with full detail (includes reason; next-action URL only when PendingPayment).",
				ToolName:    "AdminGetTenantBillingAudit",
				Flags: []FlagDef{
					{Name: "tenant", Short: "t", Description: "Public GUID of the target tenant", Required: true, Type: "string"},
					{Name: "audit", Short: "a", Description: "Public GUID of the audit row", Required: true, Type: "string"},
				},
			},
			{
				Name:        "cancel",
				Description: "Cancel a pending billing-gateway change. Records the cancelling admin's GUID; nulls the next-action URL. Does not change Tenant.BillingMethod.",
				ToolName:    "AdminCancelTenantBillingChange",
				Flags: []FlagDef{
					{Name: "tenant", Short: "t", Description: "Public GUID of the target tenant", Required: true, Type: "string"},
					{Name: "audit", Short: "a", Description: "Public GUID of the pending audit row to cancel", Required: true, Type: "string"},
					{Name: "reason", Short: "r", Description: "Globaladmin-authored cancel reason, 10-1000 chars.", Required: true, Type: "string"},
				},
			},
		},
	})
}
