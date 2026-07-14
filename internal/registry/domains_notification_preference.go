package registry

// Domain: notification-preference
//
// Exposes the signed-in user's per-tenant workorder notification preferences
// (reminder window + per-type toggles). Identity and tenant come from the MCP
// bearer token / API key context — the tools take NO userId / tenantId args.
//
// REST: GET|PUT /api/notificationpreference (current-user scoped, no path id).
//   - get → GET base path (action-name default).
//   - set → PUT base path (HTTPMethod override; "set" is not in the name->verb
//     map). Only flags the user actually passes are sent, so set is a true
//     partial update — flags must therefore carry NO Default. BodyName pins
//     each flag onto the backend WorkorderNotificationPreferenceUpdateModel
//     field, which the CLI's kebab->camel auto-conversion would otherwise
//     miss (e.g. --due-window-start -> dueWindowStart, not the field
//     dueDateWindowStartHours).
//
// Paired MCP tools:
//
//	UteamupNotificationPreferenceGet
//	UteamupNotificationPreferenceSet
func init() {
	Register(&Domain{
		Name:        "notification-preference",
		Aliases:     []string{"notification-preferences", "notif-pref", "notification-prefs"},
		Description: "Manage the signed-in user's workorder notification preferences (reminder window + per-type toggles)",
		APIPath:     "/api/notificationpreference",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get the signed-in user's workorder notification preferences for the active tenant",
				ToolName:    "UteamupNotificationPreferenceGet",
			},
			{
				Name:        "set",
				Description: "Partially update the signed-in user's workorder notification preferences (omit flags to leave them unchanged)",
				ToolName:    "UteamupNotificationPreferenceSet",
				HTTPMethod:  "PUT",
				Flags: []FlagDef{
					{Name: "due-window-start", Description: "Hours before the due date when reminders start (upper thumb, 1..24)", Type: "int", BodyName: "dueDateWindowStartHours"},
					{Name: "due-window-end", Description: "Hours before the due date when reminders stop (lower thumb, 1..24, must be < start)", Type: "int", BodyName: "dueDateWindowEndHours"},
					{Name: "start-window-start", Description: "Hours before the start date when reminders start (upper thumb, 1..24)", Type: "int", BodyName: "startDateWindowStartHours"},
					{Name: "start-window-end", Description: "Hours before the start date when reminders stop (lower thumb, 1..24, must be < start)", Type: "int", BodyName: "startDateWindowEndHours"},
					{Name: "notify-on-due-date", Description: "Enable/disable due-date reminders", Type: "bool", BodyName: "notifyOnDueDate"},
					{Name: "notify-on-start-date", Description: "Enable/disable start-date reminders", Type: "bool", BodyName: "notifyOnStartDate"},
					{Name: "notify-on-change", Description: "Enable/disable workorder change (status/content) notifications", Type: "bool", BodyName: "notifyOnChange"},
					{Name: "notify-on-comment", Description: "Enable/disable new-comment notifications", Type: "bool", BodyName: "notifyOnComment"},
				},
			},
		},
	})
}
