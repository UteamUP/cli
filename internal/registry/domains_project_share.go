package registry

func init() {
	Register(&Domain{
		Name:        "project-share",
		Aliases:     []string{"ps", "project-shares"},
		Description: "Share projects with a UteamUP user or by link, scoped to chosen sections",
		APIPath:     "/api/project/shares",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the shares (active and revoked) of a project, with each share's section scope",
				ToolName:    "UteamupProjectShareList",
				RESTPath:    "by-project/{projectGuid}",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "create",
				Description: "Share a project with an existing UteamUP user by email, or create an anonymous link",
				ToolName:    "UteamupProjectShareCreate",
				Flags: []FlagDef{
					{Name: "project-guid", BodyName: "projectGuid", Description: "Project GUID", Type: "uuid", Required: true},
					{Name: "email", Description: "Email of an existing UteamUP user (omit for an anonymous link)", Type: "string"},
					{Name: "access-level", BodyName: "accessLevel", Description: "Ceiling: read (default), update (edit existing) or write (also create and delete). No section may exceed it", Default: "read", Type: "string"},
					{Name: "workorder-access-level", BodyName: "workorderAccessLevel", Description: "What the recipient may do on the project's workorders: none (default), read, update or write. Capped by the workorders section", Default: "none", Type: "string"},
					{Name: "section", BodyName: "sections", Description: "Per-section scope as section=level, repeatable (e.g. --section stages=update --section budget=none). Omit for the safe default: everything at the ceiling except budget, bom and team", Type: "stringSlice"},
					{Name: "expires-in-days", BodyName: "expiresInDays", Description: "Days until the share expires, 1-365 (default 30)", Type: "int"},
					{Name: "note", Description: "Optional note shown to the recipient", Type: "string"},
				},
			},
			{
				Name:        "revoke",
				Description: "Revoke a share so its link stops working immediately",
				ToolName:    "UteamupProjectShareRevoke",
				HTTPMethod:  "DELETE",
				RESTPath:    "{shareGuid}",
				Args:        []ArgDef{{Name: "shareGuid", Description: "Share GUID", Required: true, Type: "uuid"}},
			},
		},
	})
}
