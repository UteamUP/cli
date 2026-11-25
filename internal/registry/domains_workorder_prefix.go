package registry

// Work order number prefixes (ticket ID layouts) are read through the backend MCP
// tools, which accept the same permissions as GET /api/workorder-prefixes.
func init() {
	Register(&Domain{
		Name:        "workorder-prefix",
		Aliases:     []string{"workorder-prefixes"},
		Description: "List and read work order number prefixes",
		Actions: []Action{
			{Name: "list", Description: "List work order prefixes", ToolName: "UteamupWorkorderPrefixList", MCPOnly: true, RejectExtraArgs: true,
				Flags: []FlagDef{{Name: "include-inactive", Description: "Also list deactivated prefixes", Type: "bool"}}},
			{Name: "get", Description: "Get a work order prefix", ToolName: "UteamupWorkorderPrefixGet", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{{Name: "prefix-guid", BodyName: "prefixGuid", Description: "Work order prefix public GUID", Required: true, Type: "non-empty-uuid"}}},
		},
	})
}
