package registry

// Subscription lifecycle admin actions (E7) — suspend/resume, immediate cancel,
// and scheduled cancellation. These live on InternalBillingController under
// /api/internalbilling/admin/subscriptions/{guid}/..., so the domain declares
// that base path explicitly rather than auto-deriving from its name.

func init() {
	Register(&Domain{
		Name:        "subscription-lifecycle",
		Aliases:     []string{"subscriptionlifecycle", "sub-lifecycle"},
		Description: "Admin lifecycle actions on internal subscriptions (suspend, cancel, reactivate, scheduled cancel)",
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
				// The backend binds `[FromBody] string? reason` — a raw JSON string,
				// which the flag→JSON-object body mapping cannot express. The CLI
				// sends no body; the nullable binder yields a null reason.
				Name:        "cancel",
				Description: "Cancel a subscription immediately",
				ToolName:    "UteamupSubscriptionCancel",
				HTTPMethod:  "POST",
				RESTPath:    "admin/subscriptions/{guid}/cancel",
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
