package registry

func init() {
	Register(&Domain{
		Name:        "user",
		Aliases:     []string{"users"},
		Description: "Manage users",
		Actions: []Action{
			// There is no /api/user list or single-user route; members are read through the
			// tenant's user routes, scoped to the X-Tenant-Guid the client sends.
			{
				Name:         "list",
				Description:  "List the members of your current tenant, with their user GUIDs",
				ToolName:     "UteamupUserList",
				RESTBasePath: "/api/tenant",
				RESTPath:     "users",
			},
			{
				Name:         "get",
				Description:  "Get one member of your current tenant by user GUID",
				ToolName:     "UteamupUserGet",
				RESTBasePath: "/api/tenant",
				RESTPath:     "users/{guid}",
				Args:         []ArgDef{{Name: "guid", Description: "The member's public user GUID (from ut user list)", Required: true, Type: "non-empty-uuid"}},
			},
			{
				Name:         "emergency-contacts",
				Description:  "Read a colleague's emergency contacts. Every read is logged and shown to the person it is about.",
				ToolName:     "UteamupEmergencycontactListForUser",
				RESTBasePath: "/api/emergencycontact",
				RESTPath:     "by-user/{guid}",
				HTTPMethod:   "GET",
				Args:         []ArgDef{{Name: "guid", Description: "The subject's public user GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "reason", Description: "Why you are reading them — shown to the subject", Type: "string", QueryName: "reason"},
				},
			},
		},
	})
}
