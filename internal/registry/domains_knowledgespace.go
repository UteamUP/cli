package registry

// Knowledge spaces — the Confluence-style containers that hold knowledge pages.
// GUID-first per CLIGuidelines.md: every identifier at the boundary is a Guid
// (the `spaceGuid` positional / `userGuid` flag), never an int id. Tool names
// mirror the backend MCP tools in MCP/Tools/KnowledgeSpaceTools.cs exactly.
func init() {
	Register(&Domain{
		Name:        "knowledgespace",
		Aliases:     []string{"kbspace", "knowledge-space", "space"},
		Description: "Manage knowledge spaces (containers for knowledge pages)",
		Actions: []Action{
			{Name: "list", Description: "List the spaces visible to you (global system + your tenant)", ToolName: "UteamupKnowledgeSpaceList"},
			{Name: "get", Description: "Get a space by GUID", ToolName: "UteamupKnowledgeSpaceGet",
				Args: []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}}},
			{Name: "create", Description: "Create a space", ToolName: "UteamupKnowledgeSpaceCreate",
				Flags: []FlagDef{
					{Name: "name", BodyName: "name", Description: "Space name", Required: true, Type: "string"},
					{Name: "description", BodyName: "description", Description: "Space description", Type: "string"},
					{Name: "icon", BodyName: "iconName", Description: "Heroicon name for the space", Type: "string"},
					{Name: "required-permission", BodyName: "requiredPermission", Description: "Platform permission required to see the space (e.g. Tenant.Read)", Type: "string"},
				}},
			{Name: "update", Description: "Update a space by GUID", ToolName: "UteamupKnowledgeSpaceUpdate",
				Args: []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "name", BodyName: "name", Description: "Space name", Type: "string"},
					{Name: "description", BodyName: "description", Description: "Space description", Type: "string"},
					{Name: "icon", BodyName: "iconName", Description: "Heroicon name for the space", Type: "string"},
					{Name: "display-order", BodyName: "displayOrder", Description: "Sort order in the switcher", Type: "int"},
					{Name: "active", BodyName: "isActive", Description: "Whether the space is active", Type: "bool"},
					{Name: "required-permission", BodyName: "requiredPermission", Description: "Platform permission required to see the space", Type: "string"},
				}},
			{Name: "delete", Description: "Delete a space by GUID", ToolName: "UteamupKnowledgeSpaceDelete",
				Args: []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}}},
			{Name: "usage", Description: "Report usage (space/page/sub-page counts vs plan caps) for a space by GUID", ToolName: "UteamupKnowledgeSpaceUsage",
				Args: []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}}},
			{Name: "list-members", Description: "List a space's members by GUID", ToolName: "UteamupKnowledgeSpaceListMembers",
				Args: []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}}},
			{Name: "add-member", Description: "Add a member to a space by GUID", ToolName: "UteamupKnowledgeSpaceAddMember",
				Args: []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "user-guid", BodyName: "userGuid", Description: "User GUID to add", Required: true, Type: "string"},
					{Name: "role", BodyName: "role", Description: "Space role: 0=Viewer, 1=Editor, 2=Admin", Default: 0, Type: "int"},
				}},
		},
	})
}
