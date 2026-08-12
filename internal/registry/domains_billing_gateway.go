package registry

// Read/cancel compatibility surface for the retired per-tenant billing-gateway switch audit.
// New provider changes use the provider-neutral BillingAgreement lifecycle and are deliberately
// not exposed here.
//
// Every action is globaladmin-only: the backend gate checks GlobalAdminEmails AND
// ApplicationUser.EmailConfirmed = true, so a caller outside that list gets 403.
//
// Command shape:
//   ut admin-billing-gateway history --tenant <guid> [--page <n>] [--page-size <n>]
//   ut admin-billing-gateway get --tenant <guid> --audit <guid>
//   ut admin-billing-gateway cancel --tenant <guid> --audit <guid> --reason "..."

func init() {
	Register(&Domain{
		Name:        "admin-billing-gateway",
		Aliases:     []string{"abg", "billing-gateway", "gateway"},
		Description: "Globaladmin-only history/cancellation for the retired billing-gateway switcher",
		APIPath:     "/api/globaladmin",
		Actions: []Action{
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
