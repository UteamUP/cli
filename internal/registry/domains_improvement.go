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
				Args:        []ArgDef{improvementProjectGuidArg()},
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
				Args:        []ArgDef{improvementProjectGuidArg()},
				Flags:       []FlagDef{jsonFlag()},
			},
			{
				Name:        "delete",
				Description: "Delete an improvement project by public GUID",
				ToolName:    "UteamupImprovementProjectDelete",
				RESTPath:    "by-guid/{projectGuid}",
				Args:        []ArgDef{improvementProjectGuidArg()},
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
				Args:        []ArgDef{improvementProjectGuidArg()},
				Flags:       []FlagDef{jsonFlag()},
			},
			improvementProjectTransitionAction("hold", "Put an improvement project on hold"),
			{
				Name:        "create-workorder",
				Description: "Create a linked work order and return its public GUID",
				ToolName:    "UteamupImprovementProjectCreateWorkOrder",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{projectGuid}/create-workorder",
				Args:        []ArgDef{improvementProjectGuidArg()},
			},
			{
				Name:        "linked-workorders",
				Description: "List work orders linked to an improvement project",
				ToolName:    "UteamupImprovementProjectLinkedWorkOrders",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{projectGuid}/linked-workorders",
				Args:        []ArgDef{improvementProjectGuidArg()},
			},
			{
				Name:        "actions",
				Description: "List actions for an improvement project",
				ToolName:    "UteamupImprovementProjectActions",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{projectGuid}/actions",
				Args:        []ArgDef{improvementProjectGuidArg()},
			},
			{
				Name:        "action-create",
				Description: "Create an action on an improvement project",
				ToolName:    "UteamupImprovementProjectActionCreate",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{projectGuid}/actions",
				Args:        []ArgDef{improvementProjectGuidArg()},
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
				Args:        []ArgDef{improvementProjectGuidArg()},
			},
			{
				Name:        "pdca-add",
				Description: "Add the current PDCA phase entry",
				ToolName:    "UteamupImprovementProjectPdcaAdd",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{projectGuid}/pdca",
				Args:        []ArgDef{improvementProjectGuidArg()},
				Flags:       []FlagDef{jsonFlag()},
			},
			{
				Name:        "pdca-complete",
				Description: "Complete a PDCA entry by public GUID",
				ToolName:    "UteamupImprovementProjectPdcaComplete",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{projectGuid}/pdca/by-guid/{entryGuid}/complete",
				Args: []ArgDef{
					improvementProjectGuidArg(),
					{Name: "entryGuid", Description: "PDCA entry public GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{jsonFlag()},
			},
		},
	})

	Register(&Domain{
		Name:        "kaizen-card",
		Aliases:     []string{"kaizen", "kc"},
		Description: "Manage kaizen cards",
		Actions:     crudActions("KaizenCard"),
	})

	Register(&Domain{
		Name:        "improvement-suggestion",
		Aliases:     []string{"suggestion", "imp-suggestion"},
		Description: "Manage improvement suggestions",
		Actions:     crudActions("ImprovementSuggestion"),
	})
}

func improvementProjectGuidArg() ArgDef {
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
		Args:        []ArgDef{improvementProjectGuidArg()},
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
			improvementProjectGuidArg(),
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
