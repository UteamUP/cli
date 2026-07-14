package registry

func init() {
	Register(&Domain{
		Name:        "workforce-group",
		Aliases:     []string{"wg"},
		Description: "Manage workforce groups",
		APIPath:     "/api/workforcegroups",
		Actions: []Action{
			{Name: "list", Description: "List workforce groups", ToolName: "UteamupWorkforceGroupList"},
			{Name: "get", Description: "Get a workforce group by GUID", ToolName: "UteamupWorkforceGroupGet", Args: []ArgDef{{Name: "groupGuid", Description: "Workforce group GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{groupGuid}"},
			{Name: "create", Description: "Create a workforce group", ToolName: "UteamupWorkforceGroupCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a workforce group by GUID", ToolName: "UteamupWorkforceGroupUpdate", Args: []ArgDef{{Name: "groupGuid", Description: "Workforce group GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{groupGuid}", Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete a workforce group by GUID", ToolName: "UteamupWorkforceGroupDelete", Args: []ArgDef{{Name: "groupGuid", Description: "Workforce group GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{groupGuid}"},
			{Name: "members", Description: "List workforce group members by group GUID", ToolName: "UteamupWorkforceGroupGetMembers", HTTPMethod: "GET", Args: []ArgDef{{Name: "groupGuid", Description: "Workforce group GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{groupGuid}/members"},
			{Name: "member-add", Description: "Add a member to a workforce group by group GUID", ToolName: "UteamupWorkforceGroupAddMember", HTTPMethod: "POST", Args: []ArgDef{{Name: "groupGuid", Description: "Workforce group GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{groupGuid}/members", Flags: []FlagDef{jsonFlag()}},
			{Name: "member-remove", Description: "Remove a workforce group member by member GUID", ToolName: "UteamupWorkforceGroupRemoveMember", HTTPMethod: "DELETE", Args: []ArgDef{{Name: "groupGuid", Description: "Workforce group GUID", Required: true, Type: "string"}, {Name: "memberGuid", Description: "Workforce group member GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{groupGuid}/members/by-guid/{memberGuid}"},
		},
	})
	Register(&Domain{
		Name:        "workforce-training",
		Description: "Manage workforce group required training",
		APIPath:     "/api/workforcegrouprequiredtraining",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List workforce group required training",
				ToolName:    "UteamupWorkforceGroupRequiredTrainingList",
				Flags:       []FlagDef{{Name: "group-guid", Description: "Optional workforce group GUID filter", Type: "string"}},
			},
			{Name: "create", Description: "Create required training", ToolName: "UteamupWorkforceGroupRequiredTrainingCreate", Flags: []FlagDef{jsonFlag()}},
			{
				Name:        "update",
				Description: "Update required training by GUID",
				ToolName:    "UteamupWorkforceGroupRequiredTrainingUpdate",
				Args:        []ArgDef{{Name: "trainingGuid", Description: "Required training GUID", Required: true, Type: "string"}},
				RESTPath:    "by-guid/{trainingGuid}",
				Flags:       []FlagDef{jsonFlag()},
			},
			{
				Name:        "delete",
				Description: "Delete required training by GUID",
				ToolName:    "UteamupWorkforceGroupRequiredTrainingDelete",
				Args:        []ArgDef{{Name: "trainingGuid", Description: "Required training GUID", Required: true, Type: "string"}},
				RESTPath:    "by-guid/{trainingGuid}",
			},
		},
	})
	Register(&Domain{Name: "workforce-planning", Aliases: []string{"wp"}, Description: "Manage workforce planning", Actions: crudActions("WorkforcePlanning")})
	Register(&Domain{Name: "skill", Aliases: []string{"skills"}, Description: "Manage skills", Actions: crudActions("Skill")})
	Register(&Domain{Name: "team", Aliases: []string{"teams"}, Description: "Manage teams", Actions: crudActions("Team")})
}
