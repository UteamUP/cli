package registry

// The permit lifecycle reads permits and templates, authors Draft permits and
// submits them. Approve, reject, sign, complete and revoke stay in the app.
func init() {
	Register(&Domain{
		Name:        "work-permit-lifecycle",
		Aliases:     []string{"permit"},
		Description: "List and read work permits and templates, create and edit Draft permits, and submit them for approval",
		Actions: []Action{
			{Name: "list", Description: "List work permits, newest first", ToolName: "UteamupWorkPermitList", MCPOnly: true,
				Flags: []FlagDef{
					{Name: "status", Description: "Draft | Submitted | SupervisorApproved | SafetyApproved | Active | Expired | Revoked | Completed", Type: "string", AllowedValues: workPermitStatuses},
					{Name: "page", Description: "Page number, starting at 1", Type: "int", Default: 1},
					{Name: "page-size", Description: "Permits per page (1-100)", Type: "int", Default: 25},
				}},
			{Name: "get", Description: "Get one work permit with hazards, signatures, approvals and approvers", ToolName: "UteamupWorkPermitGet", MCPOnly: true,
				Args: []ArgDef{permitGuidArg()}},
			{Name: "templates", Description: "List work permit templates", ToolName: "UteamupWorkPermitTemplateList", MCPOnly: true,
				Flags: []FlagDef{{Name: "include-inactive", Description: "Include deactivated templates", Type: "bool"}}},
			{Name: "template", Description: "Get one work permit template", ToolName: "UteamupWorkPermitTemplateGet", MCPOnly: true,
				Args: []ArgDef{{Name: "template-guid", BodyName: "templateGuid", Description: "Work permit template GUID", Required: true, Type: "non-empty-uuid"}}},
			{Name: "create", Description: "Create a Draft work permit", ToolName: "UteamupWorkPermitCreate", MCPOnly: true,
				Flags: append([]FlagDef{
					{Name: "title", Description: "Permit title", Required: true, Type: "string"},
					{Name: "template-guid", Description: "Active template GUID", Type: "uuid"},
				}, permitEditableFlags()...)},
			{Name: "update", Description: "Edit a Draft work permit; omitted fields stay unchanged", ToolName: "UteamupWorkPermitUpdate", MCPOnly: true,
				Args: []ArgDef{permitGuidArg()},
				Flags: append([]FlagDef{
					{Name: "title", Description: "New permit title", Type: "string"},
				}, permitEditableFlags()...)},
			{Name: "submit", Description: "Submit a Draft work permit for approval (no signature)", ToolName: "UteamupWorkPermitSubmit", MCPOnly: true,
				Args: []ArgDef{permitGuidArg()}},
		},
	})
}

var workPermitStatuses = []string{"Draft", "Submitted", "SupervisorApproved", "SafetyApproved", "Active", "Expired", "Revoked", "Completed"}

func permitGuidArg() ArgDef {
	return ArgDef{Name: "work-permit-guid", BodyName: "workPermitGuid", Description: "Work permit GUID", Required: true, Type: "non-empty-uuid"}
}

func permitEditableFlags() []FlagDef {
	return []FlagDef{
		{Name: "description", Description: "Description", Type: "string"},
		{Name: "location-guid", Description: "Location GUID", Type: "uuid"},
		{Name: "location-floor-guid", Description: "Floor GUID on that location", Type: "uuid"},
		{Name: "location", Description: "Free-text location for an ad-hoc site", Type: "string"},
		{Name: "start-date", Description: "Start date and time (ISO 8601)", Type: "string"},
		{Name: "end-date", Description: "End date and time (ISO 8601)", Type: "string"},
		{Name: "rescue-plan-json", Description: "Rescue plan as a JSON string", Type: "string"},
		{Name: "location-street", Description: "Street of a custom address", Type: "string"},
		{Name: "location-postal-code", Description: "Postal code of a custom address", Type: "string"},
		{Name: "location-city", Description: "City of a custom address", Type: "string"},
		{Name: "location-region", Description: "Region of a custom address", Type: "string"},
		{Name: "location-country", Description: "Country of a custom address", Type: "string"},
		{Name: "location-latitude", Description: "Map pin latitude (-90 to 90)", Type: "float"},
		{Name: "location-longitude", Description: "Map pin longitude (-180 to 180)", Type: "float"},
	}
}
