package registry

func init() {
	Register(&Domain{
		Name:        "workorder-share",
		Aliases:     []string{"wos", "workorder-shares"},
		Description: "Share work orders with a UteamUP user or by read-only / edit link",
		APIPath:     "/api/workorder/shares",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the shares (active and revoked) of a work order",
				ToolName:    "UteamupWorkorderShareList",
				RESTPath:    "by-workorder/{workorderGuid}",
				Args:        []ArgDef{{Name: "workorderGuid", Description: "Work order GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "create",
				Description: "Share a work order with an existing UteamUP user by email, or create an anonymous link",
				ToolName:    "UteamupWorkorderShareCreate",
				Flags: []FlagDef{
					{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Work order GUID", Type: "uuid", Required: true},
					{Name: "email", Description: "Email of an existing UteamUP user (omit for an anonymous link)", Type: "string"},
					{Name: "access-level", BodyName: "accessLevel", Description: "readOnly (default) or edit; edit allows status changes and comments", Default: "readOnly", Type: "string"},
					{Name: "expires-in-days", BodyName: "expiresInDays", Description: "Days until the share expires, 1-365 (default 30)", Type: "int"},
					{Name: "note", Description: "Optional note shown to the recipient", Type: "string"},
				},
			},
			{
				Name:        "revoke",
				Description: "Revoke a share so its link stops working immediately",
				ToolName:    "UteamupWorkorderShareRevoke",
				HTTPMethod:  "DELETE",
				RESTPath:    "{shareGuid}",
				Args:        []ArgDef{{Name: "shareGuid", Description: "Share GUID", Required: true, Type: "uuid"}},
			},
		},
	})
}
