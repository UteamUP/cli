package registry

// Domain: user-ui-state
//
// Exposes the signed-in user's persisted UI preferences and session state
// (last-known page). Identity comes from the MCP bearer token — the tools
// take NO userId / tenantId args.
//
// Paired MCP tools:
//   UteamupUserPreferencesGet / Set
//   UteamupUserStateGetLastPage / SetLastPage / ClearLastPage
func init() {
	Register(&Domain{
		Name:        "user-ui-state",
		Aliases:     []string{"ui-state", "prefs", "user-prefs"},
		Description: "Manage the signed-in user's UI preferences and last-known page",
		Actions: []Action{
			{
				Name:        "get-preferences",
				Description: "Get the signed-in user's UI preferences (Bug & Feature widget, Feature Preview widget, show restore banner)",
				ToolName:    "UteamupUserPreferencesGet",
			},
			{
				Name:        "set-preferences",
				Description: "Partially update the signed-in user's UI preferences (omit flags to leave them unchanged)",
				ToolName:    "UteamupUserPreferencesSet",
				Flags: []FlagDef{
					{Name: "enable-bug-and-feature-widget", Description: "Enable/disable the Bug & Feature capture widget", Type: "bool"},
					{Name: "enable-feature-preview-widget", Description: "Enable/disable the Feature Preview widget (bounded by tenant plan cap)", Type: "bool"},
					{Name: "show-restore-banner", Description: "Show the 'restored from last session' banner after silent restore", Type: "bool"},
					{Name: "enable-session-restore", Description: "Master toggle for session restore. When off, routes are not recorded and login goes straight to the default landing", Type: "bool"},
				},
			},
			{
				Name:        "get-last-page",
				Description: "Get the signed-in user's last-known page (null when none recorded)",
				ToolName:    "UteamupUserStateGetLastPage",
			},
			{
				Name:        "set-last-page",
				Description: "Record the signed-in user's last-known page. Path must be a relative route starting with '/'",
				ToolName:    "UteamupUserStateSetLastPage",
				Flags: []FlagDef{
					{Name: "last-page-path", Description: "Relative route path (e.g. /workorder/123)", Required: true, Type: "string"},
					{Name: "last-page-query", Description: "Optional query string (e.g. ?tab=details)", Type: "string"},
				},
			},
			{
				Name:        "clear-last-page",
				Description: "Clear the signed-in user's last-known page",
				ToolName:    "UteamupUserStateClearLastPage",
			},
		},
	})
}
