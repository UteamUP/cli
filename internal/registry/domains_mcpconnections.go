package registry

func init() {
	Register(&Domain{
		Name:        "mcpconnections",
		Aliases:     []string{"assistant-connections"},
		Description: "Review and revoke AI assistants (Grok, ChatGPT) linked to the tenant by sign-in (Tenant.Update)",
		APIPath:     "/api/mcp-oauth",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List every assistant link in the active tenant",
				ToolName:    "UteamupAssistantConnectionList",
				RESTPath:    "grants/tenant",
			},
			{
				Name:        "mine",
				Description: "List the assistants you linked yourself",
				ToolName:    "UteamupAssistantConnectionList",
				RESTPath:    "grants/mine",
				HTTPMethod:  "GET",
			},
			{
				Name:        "revoke",
				Description: "Revoke one assistant link; the assistant loses access on its next request",
				ToolName:    "UteamupAssistantConnectionRevoke",
				RESTPath:    "grants/{grantGuid}",
				HTTPMethod:  "DELETE",
				Args: []ArgDef{
					{Name: "grant-guid", BodyName: "grantGuid", Description: "Assistant link Guid (grantGuid)", Required: true, Type: "string"},
				},
			},
		},
	})
}
