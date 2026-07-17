package registry

func init() {
	Register(&Domain{
		Name:        "root-cause",
		Aliases:     []string{"rca"},
		Description: "Manage GUID-first root cause analyses",
		APIPath:     "/api/rootcauseanalysis",
		Actions: []Action{
			{Name: "list", Description: "List root cause analyses", ToolName: "UteamupRcaList"},
			{
				Name: "get", Description: "Get a root cause analysis by public GUID",
				ToolName: "UteamupRcaGet", RESTPath: "{rcaGuid}", Args: []ArgDef{rcaGuidArg()},
			},
			{
				Name: "by-status", Description: "List root cause analyses by status",
				ToolName: "UteamupRcaList", HTTPMethod: "GET", RESTPath: "by-status/{status}",
				Args: []ArgDef{{Name: "status", Description: "RCA status", Required: true, Type: "string"}},
			},
			{
				Name: "by-entity", Description: "List root cause analyses linked to an entity GUID",
				ToolName: "UteamupRcaList", HTTPMethod: "GET",
				RESTPath: "by-entity/{entityType}/by-guid/{entityGuid}",
				Args: []ArgDef{
					{Name: "entityType", Description: "Linked entity type", Required: true, Type: "string"},
					{Name: "entityGuid", Description: "Linked entity public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name: "statistics", Description: "Get root cause analysis statistics",
				ToolName: "UteamupRcaList", HTTPMethod: "GET", RESTPath: "statistics",
			},
			{
				Name: "create", Description: "Create a root cause analysis",
				ToolName: "UteamupRcaCreate", Flags: []FlagDef{jsonFlag()},
			},
			{
				Name: "update", Description: "Update a root cause analysis by public GUID",
				ToolName: "UteamupRcaUpdate", RESTPath: "{rcaGuid}", Args: []ArgDef{rcaGuidArg()},
				Flags: []FlagDef{jsonFlag()},
			},
			{
				Name: "complete", Description: "Complete a root cause analysis by public GUID",
				ToolName: "UteamupRcaComplete", HTTPMethod: "POST", RESTPath: "{rcaGuid}/complete",
				Args: []ArgDef{rcaGuidArg()}, Flags: []FlagDef{jsonFlag()},
			},
			{
				Name: "delete", Description: "Delete a root cause analysis by public GUID",
				ToolName: "UteamupRcaDelete", RESTPath: "{rcaGuid}", Args: []ArgDef{rcaGuidArg()},
			},
			{
				Name: "step-add", Description: "Add an investigation step",
				ToolName: "UteamupRcaAddStep", HTTPMethod: "POST", RESTPath: "{rcaGuid}/steps",
				Args: []ArgDef{rcaGuidArg()}, Flags: []FlagDef{jsonFlag()},
			},
			{
				Name: "step-update", Description: "Update an investigation step by public GUID",
				ToolName: "UteamupRcaUpdateStep", HTTPMethod: "PUT",
				RESTPath: "{rcaGuid}/steps/{stepGuid}",
				Args: []ArgDef{
					rcaGuidArg(),
					{Name: "stepGuid", Description: "RCA step public GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{jsonFlag()},
			},
			{
				Name: "step-delete", Description: "Delete an investigation step by public GUID",
				ToolName: "UteamupRcaDeleteStep", HTTPMethod: "DELETE",
				RESTPath: "{rcaGuid}/steps/{stepGuid}",
				Args: []ArgDef{
					rcaGuidArg(),
					{Name: "stepGuid", Description: "RCA step public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name: "steps-reorder", Description: "Reorder investigation steps by public GUID",
				ToolName: "UteamupRcaReorderSteps", HTTPMethod: "PUT",
				RESTPath: "{rcaGuid}/steps/reorder", Args: []ArgDef{rcaGuidArg()},
				Flags: []FlagDef{jsonFlag()},
			},
			{
				Name: "action-add", Description: "Add a corrective action",
				ToolName: "UteamupRcaAddAction", HTTPMethod: "POST", RESTPath: "{rcaGuid}/actions",
				Args: []ArgDef{rcaGuidArg()}, Flags: []FlagDef{jsonFlag()},
			},
			rcaActionMutation("action-update", "PUT", ""),
			rcaActionMutation("action-delete", "DELETE", ""),
			rcaActionMutation("action-create-workorder", "POST", "/create-workorder"),
			rcaLinkMutation("assets-link", "POST", "assets", "", true),
			rcaLinkMutation("asset-unlink", "DELETE", "assets", "assetGuid", false),
			rcaLinkMutation("parts-link", "POST", "parts", "", true),
			rcaLinkMutation("part-unlink", "DELETE", "parts", "partGuid", false),
			rcaLinkMutation("workorders-link", "POST", "workorders", "", true),
			rcaLinkMutation("workorder-unlink", "DELETE", "workorders", "workOrderGuid", false),
			rcaLinkMutation("knowledge-link", "POST", "knowledgearticles", "", true),
			rcaLinkMutation("knowledge-unlink", "DELETE", "knowledgearticles", "articleGuid", false),
		},
	})
}

func rcaGuidArg() ArgDef {
	return ArgDef{
		Name: "rcaGuid", Description: "Root cause analysis public GUID", Required: true, Type: "uuid",
	}
}

func rcaActionMutation(name, method, suffix string) Action {
	flags := []FlagDef(nil)
	if method == "PUT" {
		flags = []FlagDef{jsonFlag()}
	}
	return Action{
		Name: name, Description: "Mutate an RCA corrective action by public GUID",
		ToolName: "UteamupRca" + titleWord(name), HTTPMethod: method,
		RESTPath: "{rcaGuid}/actions/by-guid/{actionGuid}" + suffix,
		Args: []ArgDef{
			rcaGuidArg(),
			{Name: "actionGuid", Description: "RCA action public GUID", Required: true, Type: "uuid"},
		},
		Flags: flags,
	}
}

func rcaLinkMutation(name, method, resource, childGuid string, body bool) Action {
	path := "{rcaGuid}/" + resource
	args := []ArgDef{rcaGuidArg()}
	if childGuid != "" {
		path += "/{" + childGuid + "}"
		args = append(args, ArgDef{
			Name: childGuid, Description: "Linked entity public GUID", Required: true, Type: "uuid",
		})
	}
	flags := []FlagDef(nil)
	if body {
		flags = []FlagDef{jsonFlag()}
	}
	return Action{
		Name: name, Description: "Manage GUID-first RCA links", ToolName: "UteamupRca" + titleWord(name),
		HTTPMethod: method, RESTPath: path, Args: args, Flags: flags,
	}
}
