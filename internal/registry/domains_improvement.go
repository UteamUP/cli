package registry

func init() {
	Register(&Domain{
		Name:        "improvement-project",
		Aliases:     []string{"improvement", "imp-project"},
		Description: "Manage GUID-first improvement projects, actions, PDCA, and work-order links",
		APIPath:     "/api/improvementproject",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List improvement projects",
				ToolName:    "UteamupImprovementProjectList",
				Flags:       paginationFlags(),
			},
			{
				Name:        "get",
				Description: "Get an improvement project by public GUID",
				ToolName:    "UteamupImprovementProjectGet",
				RESTPath:    "by-guid/{projectGuid}",
				Args:        []ArgDef{improvementProjectGUIDArg()},
			},
			{
				Name:        "create",
				Description: "Create an improvement project from a GUID-first JSON request",
				ToolName:    "UteamupImprovementProjectCreate",
				Flags:       []FlagDef{jsonFlag()},
			},
			{
				Name:        "update",
				Description: "Update an improvement project by public GUID",
				ToolName:    "UteamupImprovementProjectUpdate",
				RESTPath:    "by-guid/{projectGuid}",
				Args:        []ArgDef{improvementProjectGUIDArg()},
				Flags:       []FlagDef{jsonFlag()},
			},
			{
				Name:        "delete",
				Description: "Delete an improvement project by public GUID",
				ToolName:    "UteamupImprovementProjectDelete",
				RESTPath:    "by-guid/{projectGuid}",
				Args:        []ArgDef{improvementProjectGUIDArg()},
			},
			improvementProjectTransitionAction("approve", "Approve an improvement project"),
			improvementProjectTransitionAction("start", "Start an improvement project"),
			improvementProjectTransitionAction("complete", "Complete an improvement project"),
			{
				Name:        "cancel",
				Description: "Cancel an improvement project with an audited reason",
				ToolName:    "UteamupImprovementProjectCancel",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{projectGuid}/cancel",
				Args:        []ArgDef{improvementProjectGUIDArg()},
				Flags:       []FlagDef{jsonFlag()},
			},
			improvementProjectTransitionAction("hold", "Put an improvement project on hold"),
			{
				Name:        "create-workorder",
				Description: "Create a linked work order and return its public GUID",
				ToolName:    "UteamupImprovementProjectCreateWorkOrder",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{projectGuid}/create-workorder",
				Args:        []ArgDef{improvementProjectGUIDArg()},
			},
			{
				Name:        "linked-workorders",
				Description: "List work orders linked to an improvement project",
				ToolName:    "UteamupImprovementProjectLinkedWorkOrders",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{projectGuid}/linked-workorders",
				Args:        []ArgDef{improvementProjectGUIDArg()},
			},
			{
				Name:        "actions",
				Description: "List actions for an improvement project",
				ToolName:    "UteamupImprovementProjectActions",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{projectGuid}/actions",
				Args:        []ArgDef{improvementProjectGUIDArg()},
			},
			{
				Name:        "action-create",
				Description: "Create an action on an improvement project",
				ToolName:    "UteamupImprovementProjectActionCreate",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{projectGuid}/actions",
				Args:        []ArgDef{improvementProjectGUIDArg()},
				Flags:       []FlagDef{jsonFlag()},
			},
			improvementActionMutation("action-update", "PUT", "Update an improvement action"),
			improvementActionMutation("action-delete", "DELETE", "Delete an improvement action"),
			improvementActionMutation("action-complete", "POST", "Complete an improvement action"),
			{
				Name:        "pdca",
				Description: "List PDCA entries for an improvement project",
				ToolName:    "UteamupImprovementProjectPdca",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{projectGuid}/pdca",
				Args:        []ArgDef{improvementProjectGUIDArg()},
			},
			{
				Name:        "pdca-add",
				Description: "Add the current PDCA phase entry",
				ToolName:    "UteamupImprovementProjectPdcaAdd",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{projectGuid}/pdca",
				Args:        []ArgDef{improvementProjectGUIDArg()},
				Flags:       []FlagDef{jsonFlag()},
			},
			{
				Name:        "pdca-complete",
				Description: "Complete a PDCA entry by public GUID",
				ToolName:    "UteamupImprovementProjectPdcaComplete",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{projectGuid}/pdca/by-guid/{entryGuid}/complete",
				Args: []ArgDef{
					improvementProjectGUIDArg(),
					{Name: "entryGuid", Description: "PDCA entry public GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{jsonFlag()},
			},
		},
	})

	Register(&Domain{
		Name:        "kaizen-card",
		Aliases:     []string{"kaizen", "kc"},
		Description: "Manage GUID-first kaizen cards",
		APIPath:     "/api/kaizen",
		Actions: []Action{
			{Name: "board", Description: "Get the kaizen board", ToolName: "UteamupKaizenGetBoard", HTTPMethod: "GET", RESTPath: "board"},
			{Name: "get", Description: "Get a kaizen card by public GUID", ToolName: "UteamupKaizenGet", RESTPath: "by-guid/{cardGuid}", Args: []ArgDef{kaizenCardGUIDArg()}},
			{Name: "create", Description: "Create a kaizen card", ToolName: "UteamupKaizenCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a kaizen card by public GUID", ToolName: "UteamupKaizenUpdate", RESTPath: "by-guid/{cardGuid}", Args: []ArgDef{kaizenCardGUIDArg()}, Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete a kaizen card by public GUID", ToolName: "UteamupKaizenDelete", RESTPath: "by-guid/{cardGuid}", Args: []ArgDef{kaizenCardGUIDArg()}},
			{Name: "move", Description: "Move a kaizen card by public GUID", ToolName: "UteamupKaizenMove", HTTPMethod: "PUT", RESTPath: "by-guid/{cardGuid}/move", Args: []ArgDef{kaizenCardGUIDArg()}, Flags: []FlagDef{jsonFlag()}},
			{Name: "escalate", Description: "Escalate a kaizen card to a project", ToolName: "UteamupKaizenEscalate", HTTPMethod: "POST", RESTPath: "by-guid/{cardGuid}/escalate", Args: []ArgDef{kaizenCardGUIDArg()}},
			{Name: "create-suggestion", Description: "Create a suggestion from a kaizen card", ToolName: "UteamupKaizenCreateSuggestion", HTTPMethod: "POST", RESTPath: "by-guid/{cardGuid}/create-suggestion", Args: []ArgDef{kaizenCardGUIDArg()}},
		},
	})

	Register(&Domain{
		Name:        "improvement-suggestion",
		Aliases:     []string{"suggestion", "imp-suggestion"},
		Description: "Manage GUID-first improvement suggestions",
		APIPath:     "/api/improvementsuggestion",
		Actions: []Action{
			{Name: "list", Description: "List improvement suggestions", ToolName: "UteamupImprovementSuggestionList", Flags: paginationFlags()},
			{Name: "get", Description: "Get an improvement suggestion by public GUID", ToolName: "UteamupImprovementSuggestionGet", RESTPath: "by-guid/{suggestionGuid}", Args: []ArgDef{improvementSuggestionGUIDArg()}},
			{Name: "create", Description: "Create an improvement suggestion", ToolName: "UteamupImprovementSuggestionCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "review", Description: "Review an improvement suggestion", ToolName: "UteamupImprovementSuggestionReview", HTTPMethod: "POST", RESTPath: "by-guid/{suggestionGuid}/review", Args: []ArgDef{improvementSuggestionGUIDArg()}, Flags: []FlagDef{jsonFlag()}},
			{Name: "convert", Description: "Convert an improvement suggestion to a project", ToolName: "UteamupImprovementSuggestionConvert", HTTPMethod: "POST", RESTPath: "by-guid/{suggestionGuid}/convert", Args: []ArgDef{improvementSuggestionGUIDArg()}},
			{Name: "create-workorder", Description: "Create a linked work order", ToolName: "UteamupImprovementSuggestionCreateWorkOrder", HTTPMethod: "POST", RESTPath: "by-guid/{suggestionGuid}/create-workorder", Args: []ArgDef{improvementSuggestionGUIDArg()}},
			{Name: "linked-workorders", Description: "List linked work orders", ToolName: "UteamupImprovementSuggestionLinkedWorkOrders", HTTPMethod: "GET", RESTPath: "by-guid/{suggestionGuid}/linked-workorders", Args: []ArgDef{improvementSuggestionGUIDArg()}},
		},
	})
}

func kaizenCardGUIDArg() ArgDef {
	return ArgDef{
		Name:        "cardGuid",
		Description: "Kaizen card public GUID",
		Required:    true,
		Type:        "uuid",
	}
}

func improvementSuggestionGUIDArg() ArgDef {
	return ArgDef{
		Name:        "suggestionGuid",
		Description: "Improvement suggestion public GUID",
		Required:    true,
		Type:        "uuid",
	}
}

func improvementProjectGUIDArg() ArgDef {
	return ArgDef{
		Name:        "projectGuid",
		Description: "Improvement project public GUID",
		Required:    true,
		Type:        "uuid",
	}
}

func improvementProjectTransitionAction(name, description string) Action {
	return Action{
		Name:        name,
		Description: description,
		ToolName:    "UteamupImprovementProject" + titleWord(name),
		HTTPMethod:  "POST",
		RESTPath:    "by-guid/{projectGuid}/" + name,
		Args:        []ArgDef{improvementProjectGUIDArg()},
	}
}

func improvementActionMutation(name, method, description string) Action {
	flags := []FlagDef(nil)
	if method != "DELETE" {
		flags = []FlagDef{jsonFlag()}
	}
	return Action{
		Name:        name,
		Description: description,
		ToolName:    "UteamupImprovementProject" + titleWord(name),
		HTTPMethod:  method,
		RESTPath:    "by-guid/{projectGuid}/actions/by-guid/{actionGuid}" + actionCompleteSuffix(name),
		Args: []ArgDef{
			improvementProjectGUIDArg(),
			{Name: "actionGuid", Description: "Improvement action public GUID", Required: true, Type: "uuid"},
		},
		Flags: flags,
	}
}

func actionCompleteSuffix(name string) string {
	if name == "action-complete" {
		return "/complete"
	}
	return ""
}

func titleWord(value string) string {
	result := ""
	upperNext := true
	for _, character := range value {
		if character == '-' {
			upperNext = true
			continue
		}
		if upperNext && character >= 'a' && character <= 'z' {
			character -= 'a' - 'A'
		}
		result += string(character)
		upperNext = false
	}
	return result
}
