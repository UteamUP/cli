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
	intent := FlagDef{Name: "idempotency-key", Description: "Original confirmed operation GUID; retain it and the complete request after an uncertain result", Required: true, Type: "uuid"}
	version := FlagDef{Name: "mutation-outcome-version", Description: "Minimal ContactType receipt contract version", Type: "int", Default: 1}
	revision := FlagDef{Name: "expected-updated-at", Description: "Exact original UTC ContactType updatedAt, including all seven fractional digits; never refresh an unresolved retry", Required: true, Type: "string"}
	confirmation := FlagDef{Name: "confirm", Description: "Confirm the reviewed ContactType mutation", Required: true, Type: "bool", MustBeTrue: true, LocalOnly: true}
	guidArg := []ArgDef{{Name: "contactTypeGuid", Required: true, Type: "uuid"}}
	createFlags := append(append([]FlagDef{}, contactTypeFields...), intent, version, confirmation)
	updateFlags := append(append([]FlagDef{}, createFlags...), revision)
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
				RESTPath:    "{contactTypeGuid}",
				HTTPMethod:  "GET",
				Args:        guidArg,
			},
			{
				Name:        "create",
				Description: "Create a reusable contact type",
				ToolName:    "UteamupContacttypeCreate",
				HTTPMethod:  "POST",
				Flags:       createFlags,
			},
			{
				Name:        "update",
				Description: "Update a contact type by public GUID",
				ToolName:    "UteamupContacttypeUpdate",
				RESTPath:    "{contactTypeGuid}",
				HTTPMethod:  "PUT",
				Args:        guidArg,
				Flags:       updateFlags,
			},
			{
				Name:        "delete",
				Description: "Delete a contact type and unlink it from contacts",
				ToolName:    "UteamupContacttypeDelete",
				RESTPath:    "{contactTypeGuid}",
				HTTPMethod:  "DELETE",
				Args:        guidArg,
				Flags:       []FlagDef{intent, version, revision, confirmation},
			},
		},
	})
}
