package registry

// Microsoft Teams integration — read-only diagnostics.
//
//   config     GET  /api/teamsintegration/config
//   test       POST /api/teamsintegration/config/test
//   me         GET  /api/teamsintegration/me
//   binding    GET  /api/teamsintegration/conversations/{guid}/binding
//
// Deliberately no link, bind or unbind action. Both require a delegated Microsoft token that
// belongs to a specific person and is obtained through a browser sign-in; the CLI has no way to
// acquire one, so those actions could only ever fail. The CLI calls these REST routes directly
// (CallREST); ToolName is the MCP mirror declaration.

func init() {
	Register(&Domain{
		Name:        "teamsintegration",
		Aliases:     []string{"teams-integration", "msteams"},
		Description: "Read Microsoft Teams connection status, your own link, and a conversation's mirror state",
		APIPath:     "/api/teamsintegration",
		Actions: []Action{
			{
				Name:        "config",
				Description: "Read the tenant's Teams connection status and member coverage. Requires TeamsIntegration.Manage.",
				ToolName:    "UteamupTeamsConfigGet",
				HTTPMethod:  "GET",
				RESTPath:    "config",
			},
			{
				Name:        "test",
				Description: "Probe the stored Teams connection and report which capabilities work. Requires TeamsIntegration.Manage.",
				ToolName:    "UteamupTeamsConfigTest",
				HTTPMethod:  "POST",
				RESTPath:    "config/test",
			},
			{
				Name:        "me",
				Description: "Read your own Teams link status, including whether it can actually mirror messages.",
				ToolName:    "UteamupTeamsMyIdentityGet",
				HTTPMethod:  "GET",
				RESTPath:    "me",
			},
			{
				Name:        "binding",
				Description: "Read whether one conversation is mirrored into Teams, and how many of its members it carries.",
				ToolName:    "UteamupTeamsConversationBindingGet",
				HTTPMethod:  "GET",
				RESTPath:    "conversations/{guid}/binding",
				Args: []ArgDef{
					{Name: "guid", Description: "Conversation GUID", Required: true, Type: "string"},
				},
			},
		},
	})
}
