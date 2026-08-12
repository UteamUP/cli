package registry

// Per-tenant billing-gateway switcher — CLI parity for the globaladmin REST surface on
// UteamUP_Backend/UteamUP_API/Controllers/GlobalAdminController.cs (the same operations the
// MCP tools in MCP/Tools/BillingGatewaySwitchTools.cs expose).
//
// Every action is globaladmin-only: the backend gate checks GlobalAdminEmails AND
// ApplicationUser.EmailConfirmed = true, so a caller outside that list gets 403.
//
// Command shape:
//   ut admin-billing-gateway change --tenant <guid> --to <stripe|icelandicBankTransfer> --reason "..." [--kennitala <10-digits>] [--effective <endOfCurrentCycle|startImmediately>] [--idempotency-key <key>]
//   ut admin-billing-gateway history --tenant <guid> [--page <n>] [--page-size <n>]
//   ut admin-billing-gateway get --tenant <guid> --audit <guid>
//   ut admin-billing-gateway cancel --tenant <guid> --audit <guid> --reason "..."
//
// --to and --effective take the backend ENUM NAMES verbatim. The API registers
// JsonStringEnumConverter(camelCase, allowIntegerValues: true), so the names bind directly and
// the CLI performs no translation. (A previous version of this comment promised stripe/ibt/kling
// aliases that never existed in code; "kling" was wrong regardless — Kling is the card provider
// that settles the IcelandicBankTransfer rail, not a billing method.)

func init() {
	Register(&Domain{
		Name:        "admin-billing-gateway",
		Aliases:     []string{"abg", "billing-gateway", "gateway"},
		Description: "Globaladmin-only per-tenant billing-gateway switcher (Stripe <-> IcelandicBankTransfer)",
		APIPath:     "/api/globaladmin",
		Actions: []Action{
			{
				Name:        "change",
				Description: "Request a billing-gateway change for a tenant. Returns a checkout URL that the tenant owner must complete.",
				ToolName:    "AdminChangeTenantBillingMethod",
				HTTPMethod:  "POST",
				RESTPath:    "tenants/{tenantGuid}/billing-method",
				Flags: []FlagDef{
					{Name: "tenant", Short: "t", BodyName: "tenantGuid", Description: "Public GUID of the target tenant", Required: true, Type: "string"},
					{Name: "to", BodyName: "newBillingMethod", Description: "Target billing method: stripe or icelandicBankTransfer", Required: true, Type: "string"},
					{Name: "reason", Short: "r", BodyName: "reason", Description: "Globaladmin-authored reason, 10-1000 chars. Stored verbatim on the 7-year audit log.", Required: true, Type: "string"},
					{Name: "kennitala", Short: "k", BodyName: "kennitala", Description: "10-digit Icelandic company kennitala. Required when --to = icelandicBankTransfer.", Type: "string"},
					{Name: "effective", BodyName: "effective", Description: "When the new subscription starts: endOfCurrentCycle (default) or startImmediately", Default: "endOfCurrentCycle", Type: "string"},
					{Name: "idempotency-key", HeaderName: "Idempotency-Key", Description: "Optional idempotency key sent as the Idempotency-Key header. If omitted, the server generates one. Uniqueness is scoped to (key, tenant).", Type: "string"},
				},
			},
			{
				Name:        "history",
				Description: "Paginated audit log of billing-gateway changes for a tenant (reason + next-action URL excluded from list).",
				ToolName:    "AdminGetTenantBillingHistory",
				HTTPMethod:  "GET",
				RESTPath:    "tenants/{tenantGuid}/billing-method/history",
				Flags: []FlagDef{
					{Name: "tenant", Short: "t", BodyName: "tenantGuid", Description: "Public GUID of the target tenant", Required: true, Type: "string"},
					{Name: "page", Short: "p", BodyName: "page", Description: "1-based page number. Default 1.", Type: "int"},
					{Name: "page-size", Short: "s", BodyName: "pageSize", Description: "Page size, clamped to [1, 100]. Default 20.", Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Fetch a single audit row with full detail (includes reason; next-action URL only when PendingPayment).",
				ToolName:    "AdminGetTenantBillingAudit",
				HTTPMethod:  "GET",
				RESTPath:    "tenants/{tenantGuid}/billing-method/{auditGuid}",
				Flags: []FlagDef{
					{Name: "tenant", Short: "t", BodyName: "tenantGuid", Description: "Public GUID of the target tenant", Required: true, Type: "string"},
					{Name: "audit", Short: "a", BodyName: "auditGuid", Description: "Public GUID of the audit row", Required: true, Type: "string"},
				},
			},
			{
				Name:        "cancel",
				Description: "Cancel a pending billing-gateway change. Records the cancelling admin's GUID; nulls the next-action URL. Does not change Tenant.BillingMethod.",
				ToolName:    "AdminCancelTenantBillingChange",
				HTTPMethod:  "POST",
				RESTPath:    "tenants/{tenantGuid}/billing-method/{auditGuid}/cancel",
				Flags: []FlagDef{
					{Name: "tenant", Short: "t", BodyName: "tenantGuid", Description: "Public GUID of the target tenant", Required: true, Type: "string"},
					{Name: "audit", Short: "a", BodyName: "auditGuid", Description: "Public GUID of the pending audit row to cancel", Required: true, Type: "string"},
					{Name: "reason", Short: "r", BodyName: "reason", Description: "Globaladmin-authored cancel reason, 10-1000 chars.", Required: true, Type: "string"},
				},
			},
		},
	})
}
