package registry

var contactTypeFields = []FlagDef{
	{Name: "name", Description: "Reusable contact type name", Required: true, Type: "string"},
	{Name: "color", Description: "Six-digit hexadecimal color", Default: "#1A4697", Type: "string"},
	{
		Name:          "severity",
		Description:   "Operational severity",
		Default:       "informational",
		Type:          "string",
		AllowedValues: []string{"informational", "warning", "critical"},
	},
}

func init() {
	Register(&Domain{
		Name:        "contact-type",
		Aliases:     []string{"contacttypes", "contact-types"},
		Description: "Manage reusable tenant contact types",
		APIPath:     "/api/contacttype",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List contact types",
				ToolName:    "UteamupContacttypeList",
				HTTPMethod:  "GET",
			},
			{
				Name:        "get",
				Description: "Get a contact type by public GUID",
				ToolName:    "UteamupContacttypeGet",
				RESTPath:    "{externalGuid}",
				HTTPMethod:  "GET",
				Args:        externalGUIDArg(),
			},
			{
				Name:        "create",
				Description: "Create a reusable contact type",
				ToolName:    "UteamupContacttypeCreate",
				HTTPMethod:  "POST",
				Flags:       contactTypeFields,
			},
			{
				Name:        "update",
				Description: "Update a contact type by public GUID",
				ToolName:    "UteamupContacttypeUpdate",
				RESTPath:    "{externalGuid}",
				HTTPMethod:  "PUT",
				Args:        externalGUIDArg(),
				Flags:       contactTypeFields,
			},
			{
				Name:        "delete",
				Description: "Delete a contact type and unlink it from contacts",
				ToolName:    "UteamupContacttypeDelete",
				RESTPath:    "{externalGuid}",
				HTTPMethod:  "DELETE",
				Args:        externalGUIDArg(),
			},
		},
	})
}
