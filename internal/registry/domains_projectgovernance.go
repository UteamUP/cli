package registry

var projectGuidArgument = []ArgDef{
	{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
}

func projectResourceArguments(resourceName, description string) []ArgDef {
	return []ArgDef{
		{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
		{Name: resourceName, Description: description, Required: true, Type: "string"},
	}
}

func init() {
	Register(&Domain{
		Name:        "project-member",
		Aliases:     []string{"project-members"},
		Description: "Manage project team membership, roles, and allocation",
		APIPath:     "/api/projects",
		Actions: []Action{
			{Name: "list", Description: "List project members", ToolName: "UteamupProjectMembersList", RESTPath: "{projectGuid}/members", Args: projectGuidArgument},
			{Name: "add", Description: "Add a project member", ToolName: "UteamupProjectMembersAdd", HTTPMethod: "POST", RESTPath: "{projectGuid}/members", Args: projectGuidArgument, Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a project member", ToolName: "UteamupProjectMembersUpdate", HTTPMethod: "PUT", RESTPath: "{projectGuid}/members/{memberGuid}", Args: projectResourceArguments("memberGuid", "Project member GUID"), Flags: []FlagDef{jsonFlag()}},
			{Name: "remove", Description: "Remove a project member", ToolName: "UteamupProjectMembersRemove", HTTPMethod: "DELETE", RESTPath: "{projectGuid}/members/{memberGuid}", Args: projectResourceArguments("memberGuid", "Project member GUID")},
		},
	})

	Register(&Domain{
		Name:        "project-dependency",
		Aliases:     []string{"project-dependencies"},
		Description: "Manage structured workorder dependencies within a project",
		APIPath:     "/api/projects",
		Actions: []Action{
			{Name: "list", Description: "List project dependencies", ToolName: "UteamupProjectDependenciesList", RESTPath: "{projectGuid}/dependencies", Args: projectGuidArgument},
			{Name: "add", Description: "Add a project dependency", ToolName: "UteamupProjectDependenciesAdd", HTTPMethod: "POST", RESTPath: "{projectGuid}/dependencies", Args: projectGuidArgument, Flags: []FlagDef{jsonFlag()}},
			{Name: "remove", Description: "Remove a project dependency", ToolName: "UteamupProjectDependenciesRemove", HTTPMethod: "DELETE", RESTPath: "{projectGuid}/dependencies/{dependencyGuid}", Args: projectResourceArguments("dependencyGuid", "Project dependency GUID")},
		},
	})

	Register(&Domain{
		Name:        "project-activity",
		Description: "Read the project activity trail",
		APIPath:     "/api/projects",
		Actions: []Action{
			{Name: "list", Description: "List project activity", ToolName: "UteamupProjectActivityList", RESTPath: "{projectGuid}/activity", Args: projectGuidArgument, Flags: []FlagDef{{Name: "limit", Description: "Maximum activity rows", Default: 100, Type: "int"}}},
		},
	})

	Register(&Domain{
		Name:        "project-comment",
		Aliases:     []string{"project-comments"},
		Description: "Manage project comments and recorded decisions",
		APIPath:     "/api/projects",
		Actions: []Action{
			{Name: "list", Description: "List project comments", ToolName: "UteamupProjectCommentsList", RESTPath: "{projectGuid}/comments", Args: projectGuidArgument},
			{Name: "add", Description: "Add a project comment or decision", ToolName: "UteamupProjectCommentsAdd", HTTPMethod: "POST", RESTPath: "{projectGuid}/comments", Args: projectGuidArgument, Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a project comment or decision", ToolName: "UteamupProjectCommentsUpdate", HTTPMethod: "PUT", RESTPath: "{projectGuid}/comments/{commentGuid}", Args: projectResourceArguments("commentGuid", "Project comment GUID"), Flags: []FlagDef{jsonFlag()}},
			{Name: "remove", Description: "Remove a project comment", ToolName: "UteamupProjectCommentsRemove", HTTPMethod: "DELETE", RESTPath: "{projectGuid}/comments/{commentGuid}", Args: projectResourceArguments("commentGuid", "Project comment GUID")},
		},
	})

	Register(&Domain{
		Name:        "project-baseline",
		Aliases:     []string{"project-baselines"},
		Description: "Capture project baselines and inspect variance",
		APIPath:     "/api/projects",
		Actions: []Action{
			{Name: "list", Description: "List project baselines", ToolName: "UteamupProjectBaselinesList", RESTPath: "{projectGuid}/baselines", Args: projectGuidArgument},
			{Name: "capture", Description: "Capture an immutable project baseline", ToolName: "UteamupProjectBaselinesCapture", HTTPMethod: "POST", RESTPath: "{projectGuid}/baselines", Args: projectGuidArgument, Flags: []FlagDef{jsonFlag()}},
			{Name: "variance", Description: "Get variance from the latest or selected baseline", ToolName: "UteamupProjectVarianceGet", HTTPMethod: "GET", RESTPath: "{projectGuid}/variance", Args: projectGuidArgument, Flags: []FlagDef{{Name: "baseline-guid", Description: "Optional baseline GUID", Type: "string"}}},
		},
	})

	Register(&Domain{
		Name:        "project-change-request",
		Aliases:     []string{"project-change-requests"},
		Description: "Manage reviewed project change requests",
		APIPath:     "/api/projects",
		Actions: []Action{
			{Name: "list", Description: "List project change requests", ToolName: "UteamupProjectChangeRequestsList", RESTPath: "{projectGuid}/change-requests", Args: projectGuidArgument},
			{Name: "create", Description: "Create a draft project change request", ToolName: "UteamupProjectChangeRequestsCreate", HTTPMethod: "POST", RESTPath: "{projectGuid}/change-requests", Args: projectGuidArgument, Flags: []FlagDef{jsonFlag()}},
			{Name: "submit", Description: "Submit a draft change request", ToolName: "UteamupProjectChangeRequestsSubmit", HTTPMethod: "POST", RESTPath: "{projectGuid}/change-requests/{requestGuid}/submit", Args: projectResourceArguments("requestGuid", "Change request GUID")},
			{Name: "approve", Description: "Approve a submitted change request", ToolName: "UteamupProjectChangeRequestsApprove", HTTPMethod: "POST", RESTPath: "{projectGuid}/change-requests/{requestGuid}/approve", Args: projectResourceArguments("requestGuid", "Change request GUID")},
			{Name: "reject", Description: "Reject a submitted change request", ToolName: "UteamupProjectChangeRequestsReject", HTTPMethod: "POST", RESTPath: "{projectGuid}/change-requests/{requestGuid}/reject", Args: projectResourceArguments("requestGuid", "Change request GUID")},
			{Name: "apply", Description: "Apply an approved change request", ToolName: "UteamupProjectChangeRequestsApply", HTTPMethod: "POST", RESTPath: "{projectGuid}/change-requests/{requestGuid}/apply", Args: projectResourceArguments("requestGuid", "Change request GUID")},
		},
	})
}
