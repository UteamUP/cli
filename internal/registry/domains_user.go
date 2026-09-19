package registry

func init() {
	Register(&Domain{
		Name:        "user",
		Aliases:     []string{"users"},
		Description: "Manage users",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List all users with pagination",
				ToolName:    "UteamupUserList",
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
					{Name: "filter", Short: "f", Description: "Filter by name or email", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get user details by ID",
				ToolName:    "UteamupUserGet",
				Args:        []ArgDef{{Name: "id", Description: "User ID", Required: true, Type: "string"}},
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
