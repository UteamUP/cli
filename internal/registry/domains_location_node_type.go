package registry

func locationNodeTypeActions() []Action {
	fields := jsonFlag()
	fields.Required = true
	fields.Description = "Reviewed node-type JSON: name, optional iconName and displayOrder"
	confirm := FlagDef{Name: "confirm", Description: "Confirm the reviewed node-type change", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true}
	return []Action{
		{Name: "node-types", Description: "List this tenant's node-type GUIDs and labels", ToolName: "UteamupLocationNodeTypeList", HTTPMethod: "GET", RESTBasePath: "/api/locationnodetype"},
		{Name: "node-type-create", Description: "Create a tenant node type from reviewed JSON", ToolName: "UteamupLocationNodeTypeCreate", HTTPMethod: "POST", RESTBasePath: "/api/locationnodetype", Flags: []FlagDef{fields, confirm}},
		{Name: "node-type-update", Description: "Update a node type by public GUID", ToolName: "UteamupLocationNodeTypeUpdate", HTTPMethod: "PUT", RESTBasePath: "/api/locationnodetype", RESTPath: "{typeGuid}",
			Args: []ArgDef{{Name: "typeGuid", Description: "Public node-type GUID", Type: "non-empty-uuid", Required: true}}, Flags: []FlagDef{fields, confirm}},
	}
}
