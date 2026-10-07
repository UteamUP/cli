package registry

// listTemplateActions shares the existing GUID routes for checklist and task-list
// templates. JSON files carry the complete typed model without flattening steps.
func listTemplateActions(toolPrefix string) []Action {
	modelFile := jsonFlag()
	modelFile.Required = true
	modelFile.Description = "JSON template model: name, items/tasks, and workorderMaintenanceTypes (empty = all types); checklist items retain externalGuid and may have a condition referencing an earlier step"
	return []Action{
		{Name: "list", Description: "List templates", ToolName: "Uteamup" + toolPrefix + "List", HTTPMethod: "GET", Flags: paginationFlags()},
		{Name: "get", Description: "Get a template by public GUID", ToolName: "Uteamup" + toolPrefix + "Get", HTTPMethod: "GET", RESTPath: "{externalGuid}", Args: externalGUIDArg()},
		{Name: "create", Description: "Create a template with optional work-order type policy", ToolName: "Uteamup" + toolPrefix + "Create", HTTPMethod: "POST", Flags: []FlagDef{modelFile}},
		{Name: "update", Description: "Update a template and preserve step GUIDs and recorded answers", ToolName: "Uteamup" + toolPrefix + "Update", HTTPMethod: "PUT", RESTPath: "{externalGuid}", Args: externalGUIDArg(), Flags: []FlagDef{modelFile}},
		{Name: "delete", Description: "Delete a template by public GUID", ToolName: "Uteamup" + toolPrefix + "Delete", HTTPMethod: "DELETE", RESTPath: "{externalGuid}", Args: externalGUIDArg()},
	}
}
