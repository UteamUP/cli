package registry

// Tenant API key CLI surface — mint + manage tenant-scoped API keys (e.g. an
// MCP-enabled key for ChatGPT) from the terminal without the web UI.
//
// Backend (GUID-first per Guidelines/ApiGuidelines.md):
//   create  POST /api/tenant-api-keys                          → key incl. guid, apiKey, secret (shown once)
//   list    GET  /api/tenant-api-keys                          → items include guid
//   get     GET  /api/tenant-api-keys/by-guid/{guid}
//   revoke  POST /api/tenant-api-keys/by-guid/{guid}/revoke
//
// APIPath is set explicitly: the registry strips hyphens from the domain Name
// when auto-deriving a base path, so "apikey" would become "/api/apikey" — the
// real route is "/api/tenant-api-keys".

func init() {
	Register(&Domain{
		Name:        "apikey",
		Aliases:     []string{"apikeys", "api-key"},
		Description: "Mint and manage tenant API keys (create an MCP-enabled key for ChatGPT, list, get, revoke)",
		APIPath:     "/api/tenant-api-keys",
		Actions: []Action{
			{
				Name:        "create",
				Description: "Create a tenant API key. The secret is returned ONCE in the response — copy it immediately. Use --mcp-enabled for a key usable with the MCP server (e.g. ChatGPT).",
				ToolName:    "UteamupTenantApiKeyCreate",
				Flags: []FlagDef{
					{Name: "name", Description: "API key name (required)", Required: true, Type: "string"},
					{Name: "description", Description: "API key description", Type: "string"},
					{Name: "role-id", Description: "Role assigned to the key (controls its permissions)", Type: "string"},
					{Name: "mcp-enabled", Description: "Enable this key for MCP server use (e.g. ChatGPT)", Default: false, Type: "bool"},
					{Name: "expires-at", Description: "Expiry timestamp (ISO 8601, e.g. 2027-01-01T00:00:00Z)", Type: "string"},
					{Name: "requests-per-minute", Description: "Rate limit: requests per minute", Type: "int"},
					{Name: "requests-per-hour", Description: "Rate limit: requests per hour", Type: "int"},
					{Name: "requests-per-day", Description: "Rate limit: requests per day", Type: "int"},
					{Name: "allowed-ip-addresses", Description: "Allowed IP addresses — repeatable or comma-separated (empty = any)", Type: "stringSlice"},
				},
			},
			{
				Name:        "list",
				Description: "List the current tenant's API keys (secrets are never returned here)",
				ToolName:    "UteamupTenantApiKeyList",
			},
			{
				Name:        "get",
				Description: "Get a tenant API key by its stable GUID",
				ToolName:    "UteamupTenantApiKeyGet",
				RESTPath:    "by-guid/{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "API key GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "revoke",
				Description: "Revoke a tenant API key by its stable GUID — the key stops working immediately",
				ToolName:    "UteamupTenantApiKeyRevoke",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{guid}/revoke",
				Args:        []ArgDef{{Name: "guid", Description: "API key GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
		},
	})
}
