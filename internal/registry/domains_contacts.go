package registry

func init() {
	Register(&Domain{
		Name:        "contact",
		Aliases:     []string{"contacts"},
		Description: "Manage contacts",
		Actions: append(contactActions(),
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

// contactActions keeps the original normal REST contracts but opts mutations into
// immutable actor-bound outcomes. Never generate a replacement key or refresh a
// reviewed revision automatically after an uncertain response.
func contactActions() []Action {
	intent := FlagDef{Name: "idempotency-key", Description: "One confirmed operation GUID; retain the same complete request after an uncertain result", Required: true, Type: "uuid"}
	version := FlagDef{Name: "mutation-outcome-version", Description: "Minimal Contact receipt contract version", Type: "int", Default: 1}
	revision := FlagDef{Name: "expected-updated-at", Description: "Exact original Contact updatedAt UTC wire value, including all seven fractional digits; never refresh an unresolved retry", Required: true, Type: "string"}
	confirmation := FlagDef{Name: "confirm", Description: "Confirm the reviewed Contact mutation", Required: true, Type: "bool", MustBeTrue: true, LocalOnly: true}
	fields := jsonFlag()
	fields.Required = true
	fields.Description = "Reviewed Contact operational fields JSON; keep the original file unchanged across uncertain retries and supply receipt metadata separately"
	return []Action{
		{Name: "list", Description: "List visible contacts", ToolName: "UteamupContactList", RESTPath: "paginated", Flags: paginationFlags()},
		{Name: "get", Description: "Get a visible contact by public GUID", ToolName: "UteamupContactGet", RESTPath: "{contactGuid}", Args: []ArgDef{{Name: "contactGuid", Required: true, Type: "uuid"}}},
		{Name: "create", Description: "Create a contact with an original operation key and minimal acknowledgment", ToolName: "UteamupContactCreate", Flags: []FlagDef{intent, version, confirmation, fields}},
		{Name: "update", Description: "Update the reviewed Contact GUID and exact original revision", ToolName: "UteamupContactUpdate", RESTPath: "{contactGuid}", Args: []ArgDef{{Name: "contactGuid", Required: true, Type: "uuid"}}, Flags: []FlagDef{intent, version, revision, confirmation, fields}},
		{Name: "delete", Description: "Delete the reviewed Contact GUID and retain the original key and revision for retries", ToolName: "UteamupContactDelete", RESTPath: "{contactGuid}", Args: []ArgDef{{Name: "contactGuid", Required: true, Type: "uuid"}}, Flags: []FlagDef{intent, version, revision, confirmation}},
	}
}
