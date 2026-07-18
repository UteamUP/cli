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
	// Mirrors the MCP planning tools in UteamUP_Backend/UteamUP_API/MCP/Tools/
	// WorkforcePlanningTools.cs with named, argument-typed actions (the old
	// crudActions stub had no matching REST surface at all). grid is served by
	// WorkforcePlanningController; assign, available-technicians, and
	// group-members are served by the cross-domain controllers the MCP handlers
	// share repositories with (WorkOrderController, ScheduleController,
	// WorkforceGroupController) via RESTBasePath. The remaining MCP planning
	// tools (UteamupWorkforceGetCapacity, UteamupWorkforceSuggestAssignment,
	// UteamupWorkforceExplainAssignment) are intentionally NOT exposed here:
	// their logic lives inside MCP handlers with no REST adapter route, so a
	// CLI action would 404. Add them when WorkforcePlanningController grows
	// the matching routes.
	Register(&Domain{
		Name:        "workforce-planning",
		Aliases:     []string{"wp"},
		Description: "Workforce planning grid, capacity, and assignment intelligence",
		APIPath:     "/api/workforceplanning",
		Actions: []Action{
			{
				Name:        "grid",
				Description: "Get the workforce planning grid (shifts, members, workorder counts) for a date range",
				ToolName:    "UteamupWorkforcePlanningGrid",
				HTTPMethod:  "GET",
				RESTPath:    "grid",
				Flags: []FlagDef{
					{Name: "date-from", Description: "Start date of the grid range (YYYY-MM-DD)", Required: true, Type: "string"},
					{Name: "date-to", Description: "End date of the grid range (YYYY-MM-DD)", Required: true, Type: "string"},
					{Name: "category-guid", BodyName: "categoryGuids", Description: "Optional workforce group category GUID filter (single value)", Type: "string"},
					{Name: "location-guid", Description: "Optional location GUID to filter shifts by location", Type: "string"},
					{Name: "customer-guid", Description: "Optional customer GUID to filter shifts by customer", Type: "string"},
				},
			},
			{
				Name:         "assign",
				Description:  "Assign a workorder to a workforce member (updates the workorder's assignee)",
				ToolName:     "UteamupWorkforceAssignWorkOrder",
				HTTPMethod:   "PUT",
				RESTBasePath: "/api/workorder",
				RESTPath:     "by-guid/{workOrderGuid}/assignee",
				Args: []ArgDef{
					{Name: "workOrderGuid", Description: "Workorder GUID to assign", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "assignee-user-guid", BodyName: "assigneeGuid", Description: "User GUID of the member to assign the workorder to", Required: true, Type: "string"},
					{Name: "override-reason", BodyName: "qualificationOverrideReason", Description: "Optional reason justifying assignment despite unmet qualification requirements", Type: "string"},
				},
			},
			{
				Name:         "available-technicians",
				Description:  "Get available technicians for a date, optionally filtered by skill, certificate, or team",
				ToolName:     "UteamupWorkforceGetAvailableTechnicians",
				HTTPMethod:   "GET",
				RESTBasePath: "/api/schedule",
				RESTPath:     "available-technicians",
				Flags: []FlagDef{
					{Name: "date", Description: "Date to check availability for (YYYY-MM-DD or ISO8601)", Required: true, Type: "string"},
					{Name: "required-skill-guid", BodyName: "requiredSkillGuids", Description: "Optional skill GUID to filter technicians by capability (single value)", Type: "string"},
					{Name: "required-certificate-guid", BodyName: "requiredCertificateGuids", Description: "Optional certificate GUID to filter technicians by capability (single value)", Type: "string"},
					{Name: "team-guid", Description: "Optional team GUID to restrict availability checks", Type: "string"},
				},
			},
			{
				Name:         "group-members",
				Description:  "Get the members of a workforce group for planning grid display",
				ToolName:     "UteamupWorkforceGetGroupMembers",
				HTTPMethod:   "GET",
				RESTBasePath: "/api/workforcegroups",
				RESTPath:     "by-guid/{groupGuid}/members",
				Args: []ArgDef{
					{Name: "groupGuid", Description: "Workforce group GUID to get members for", Required: true, Type: "uuid"},
				},
			},
		},
	})
	Register(&Domain{Name: "skill", Aliases: []string{"skills"}, Description: "Manage skills", Actions: crudActions("Skill")})
	Register(&Domain{Name: "team", Aliases: []string{"teams"}, Description: "Manage teams", Actions: crudActions("Team")})
}
