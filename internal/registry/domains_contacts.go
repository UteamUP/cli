package registry

func init() {
	Register(&Domain{
		Name:        "contact",
		Aliases:     []string{"contacts"},
		Description: "Manage contacts",
		Actions: append(crudActions("Contact"),
			Action{Name: "search", Description: "Search contacts", ToolName: "UteamupContactSearch", Args: queryArg(), Flags: paginationFlags()},
			Action{
				Name:         "ice-assignments",
				Description:  "List users whose ICE record points at a contact. Reads are privacy-audited.",
				ToolName:     "UteamupEmergencycontactAssignmentsForContact",
				RESTBasePath: "/api/emergencycontact",
				RESTPath:     "by-contact/{guid}",
				HTTPMethod:   "GET",
				Args:         []ArgDef{{Name: "guid", Description: "The contact's public GUID", Required: true, Type: "uuid"}},
			},
			Action{
				Name:         "ice-assignments-sync",
				Description:  "Replace the employees whose ICE record points at a contact",
				ToolName:     "UteamupEmergencycontactAssignmentsSyncForContact",
				RESTBasePath: "/api/emergencycontact",
				RESTPath:     "by-contact/{guid}/assignments",
				HTTPMethod:   "PUT",
				Args:         []ArgDef{{Name: "guid", Description: "The contact's public GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{{
					Name:               "from-json",
					Description:        "JSON object with reviewed userGuids array; use [] to clear all assignments",
					Type:               "string",
					Required:           true,
					RootJSONObjectFile: true,
				}},
			},
		),
	})

	Register(&Domain{
		Name:        "customer",
		Aliases:     []string{"customers"},
		Description: "Manage customers",
		Actions: append(crudActions("Customer"),
			Action{Name: "search", Description: "Search customers", ToolName: "UteamupCustomerSearch", Args: queryArg(), Flags: paginationFlags()},
		),
	})

}
