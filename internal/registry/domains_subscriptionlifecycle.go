package registry

// Subscription lifecycle admin actions (E7) — suspend/resume and scheduled
// cancellation. These live on InternalBillingController under
// /api/internalbilling/admin/subscriptions/{guid}/..., so the domain declares
// that base path explicitly rather than auto-deriving from its name.

func init() {
	Register(&Domain{
		Name:        "subscription-lifecycle",
		Aliases:     []string{"subscriptionlifecycle", "sub-lifecycle"},
		Description: "Admin lifecycle actions on internal subscriptions (suspend, reactivate, scheduled cancel, extend period)",
		APIPath:     "/api/internalbilling",
		Actions: []Action{
			{
				Name:        "suspend",
				Description: "Suspend a subscription (tenant drops to read-only mode)",
				ToolName:    "UteamupSubscriptionSuspend",
				HTTPMethod:  "POST",
				RESTPath:    "admin/subscriptions/{guid}/suspend",
				Args:        []ArgDef{{Name: "guid", Description: "Subscription GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "reactivate",
				Description: "Reactivate a suspended subscription (resume)",
				ToolName:    "UteamupSubscriptionReactivate",
				HTTPMethod:  "POST",
				RESTPath:    "admin/subscriptions/{guid}/reactivate",
				Args:        []ArgDef{{Name: "guid", Description: "Subscription GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "schedule-cancel",
				Description: "Schedule a future cancellation; the lifecycle sweep cancels once the date passes",
				ToolName:    "UteamupSubscriptionScheduleCancel",
				HTTPMethod:  "POST",
				RESTPath:    "admin/subscriptions/{guid}/schedule-cancel",
				Args:        []ArgDef{{Name: "guid", Description: "Subscription GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "cancel-at", Description: "When to cancel (ISO 8601 datetime, must be in the future; required)", Required: true, Type: "string"},
				},
			},
			{
				// BIL-002. Waives only the settled-invoice guard; the tenant-state and
				// resolvable-owner guards still apply, so this cannot revive a cancelled
				// tenant. The reason is mandatory — the backend rejects blank and
				// whitespace — and the resulting subscription is excluded from MRR/ARR
				// until a real invoice settles.
				Name:        "activate-without-payment",
				Description: "Activate a subscription with no settled invoice (audited override; requires a reason)",
				ToolName:    "UteamupSubscriptionActivateWithoutPayment",
				HTTPMethod:  "POST",
				RESTPath:    "admin/subscriptions/{guid}/activate-without-payment",
				Args:        []ArgDef{{Name: "guid", Description: "Subscription GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "reason", Description: "Why payment is being waived (required; recorded permanently in the admin audit log)", Required: true, Type: "string"},
				},
			},
			{
				// For a bank-transfer tenant with no subscription at all (a demo or test
				// tenant), where activate-without-payment has no row to act on. Creates one on
				// the tenant's current plan and runs the same audited override; returns the
				// existing active subscription when there already is one.
				Name:        "grant-without-payment",
				Description: "Give a bank-transfer tenant with no subscription an active one on its plan (audited override; requires a reason)",
				ToolName:    "UteamupSubscriptionGrantWithoutPayment",
				HTTPMethod:  "POST",
				RESTPath:    "admin/tenants/{tenantGuid}/subscriptions/grant-without-payment",
				Args:        []ArgDef{{Name: "tenantGuid", Description: "Tenant GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "reason", Description: "Why payment is being waived (required; recorded permanently in the admin audit log)", Required: true, Type: "string"},
				},
			},
			{
				// Global Admin only. Moves an active subscription's current period end later
				// (at most 10 years ahead). The reason is written only to the admin audit log,
				// which only Global Admins can read.
				Name:        "extend-period",
				Description: "Move an active subscription's period end later (audited; requires a reason)",
				ToolName:    "UteamupSubscriptionExtendPeriod",
				HTTPMethod:  "POST",
				RESTPath:    "admin/subscriptions/{guid}/extend-period",
				Args:        []ArgDef{{Name: "guid", Description: "Subscription GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "new-period-end", Description: "New period end (ISO 8601 UTC, later than the current end and at most 10 years ahead; required)", Required: true, Type: "string"},
					{Name: "reason", Description: "Why the period is extended (required; recorded in the Global Admin audit log)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "clear-scheduled-cancel",
				Description: "Clear a scheduled cancellation before it fires; the subscription stays active",
				ToolName:    "UteamupSubscriptionClearScheduledCancel",
				HTTPMethod:  "POST",
				RESTPath:    "admin/subscriptions/{guid}/clear-scheduled-cancel",
				Args:        []ArgDef{{Name: "guid", Description: "Subscription GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
		},
	})
}
