package registry

// Knowledge spaces — the Confluence-style containers that hold knowledge pages.
// GUID-first per CLIGuidelines.md: every identifier at the boundary is a Guid
// (the `spaceGuid` positional / `userGuid` flag), never an int id. Tool names
// mirror the backend MCP tools in MCP/Tools/KnowledgeSpaceTools.cs exactly.
//
// Access contract (UTE-960): `access-catalog` lists the platform permissions and
// tenant roles a space manager may pick from. Create and update carry the same
// choices as `requiredPermissions` (string[]) and `allowedRoleGuids` (Guid[]).
// Any one selected permission or tenant role grants viewing; the obsolete
// single `requiredPermission` stays supported. Omitting a list on update keeps
// the stored list, an empty list clears it.
//
// The single-space routes are spelled out because the positional argument is
// `spaceGuid`, not the `id`/`externalGuid` names buildRESTPath recognises by
// default — without RESTPath the get/update/delete verbs would fall through to
// the collection root and never reach the backend's by-guid routes.
func init() {
	Register(&Domain{
		Name:        "knowledgespace",
		Aliases:     []string{"kbspace", "knowledge-space", "space"},
		Description: "Manage knowledge spaces (containers for knowledge pages)",
		Actions: withRecordArchive([]Action{
			{Name: "list", Description: "List the spaces visible to you (global system + your tenant)", ToolName: "UteamupKnowledgeSpaceList"},
			{Name: "access-catalog", Description: "List the platform permissions and tenant roles that can grant access to a space (requires Knowledge.ManageSpaces; role groups are not supported)", ToolName: "UteamupKnowledgeSpaceAccessCatalog",
				HTTPMethod: "GET", RESTPath: "access-catalog"},
			{Name: "get", Description: "Get a space by GUID", ToolName: "UteamupKnowledgeSpaceGet",
				RESTPath: "by-guid/{spaceGuid}",
				Args:     []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}}},
			{Name: "create", Description: "Create a space", ToolName: "UteamupKnowledgeSpaceCreate",
				Flags: []FlagDef{
					{Name: "name", BodyName: "name", Description: "Space name", Required: true, Type: "string"},
					{Name: "description", BodyName: "description", Description: "Space description", Type: "string"},
					{Name: "icon", BodyName: "iconName", Description: "Heroicon name for the space", Type: "string"},
					{Name: "required-permission", BodyName: "requiredPermission", Description: "Obsolete single platform permission required to see the space (e.g. Tenant.Read); System.All stays a platform-configured Global Admin requirement", Type: "string"},
					{Name: "required-permissions", BodyName: "requiredPermissions", Description: "Platform permissions that grant access to the space; any one of them is enough. Repeatable or comma-separated", Type: "stringSlice"},
					{Name: "allowed-role-guids", BodyName: "allowedRoleGuids", Description: "Tenant role GUIDs that grant access to the space; any one of them is enough. Repeatable or comma-separated", Type: "stringSlice"},
				}},
			{Name: "update", Description: "Update a space by GUID", ToolName: "UteamupKnowledgeSpaceUpdate",
				RESTPath: "by-guid/{spaceGuid}",
				Args:     []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "name", BodyName: "name", Description: "Space name", Type: "string"},
					{Name: "description", BodyName: "description", Description: "Space description", Type: "string"},
					{Name: "icon", BodyName: "iconName", Description: "Heroicon name for the space", Type: "string"},
					{Name: "display-order", BodyName: "displayOrder", Description: "Sort order in the switcher", Type: "int"},
					{Name: "active", BodyName: "isActive", Description: "Whether the space is active", Type: "bool"},
					{Name: "required-permission", BodyName: "requiredPermission", Description: "Obsolete single platform permission required to see the space; System.All stays a platform-configured Global Admin requirement", Type: "string"},
					{Name: "required-permissions", BodyName: "requiredPermissions", Description: "Complete platform-permission set; any one of them is enough. Omit to keep the current set, pass an empty value to clear it", Type: "stringSlice"},
					{Name: "allowed-role-guids", BodyName: "allowedRoleGuids", Description: "Complete tenant-role set; any one of them is enough. Omit to keep the current set, pass an empty value to clear it", Type: "stringSlice"},
				}},
			{Name: "delete", Description: "Delete a space by GUID", ToolName: "UteamupKnowledgeSpaceDelete",
				RESTPath: "by-guid/{spaceGuid}",
				Args:     []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}}},
			{Name: "usage", Description: "Report usage (space/page/sub-page counts vs plan caps) for a space by GUID", ToolName: "UteamupKnowledgeSpaceUsage",
				HTTPMethod: "GET", RESTPath: "by-guid/{spaceGuid}/usage",
				Args: []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}}},
			{Name: "list-members", Description: "List a space's members by GUID", ToolName: "UteamupKnowledgeSpaceListMembers",
				HTTPMethod: "GET", RESTPath: "by-guid/{spaceGuid}/members",
				Args: []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}}},
			{Name: "add-member", Description: "Add a member to a space by GUID", ToolName: "UteamupKnowledgeSpaceAddMember",
				HTTPMethod: "POST", RESTPath: "by-guid/{spaceGuid}/members",
				Args: []ArgDef{{Name: "spaceGuid", Description: "Knowledge space GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "user-guid", BodyName: "userGuid", Description: "User GUID to add", Required: true, Type: "string"},
					{Name: "role", BodyName: "role", Description: "Space role: 0=Viewer, 1=Editor, 2=Admin", Default: 0, Type: "int"},
				}},
		}, "knowledge space", "knowledge spaces", "KnowledgeSpace.Delete"),
	})
}
