package registry

func init() {
	Register(&Domain{
		Name: "project-source-line", Aliases: []string{"project-source-lines"},
		Description: "Read current commercial source lines and retained allocation history", APIPath: "/api/projects",
		Actions: []Action{{Name: "list", Description: "List current approved lines; retired allocations remain history",
			ToolName: "UteamupProjectSourceLinesList", RESTPath: "{projectGuid}/intake/lines", Args: projectGUIDArgument,
			Flags: []FlagDef{
				{Name: "intake-guid", Description: "Optional source intake GUID", Type: "string"},
				{Name: "page", Description: "One-based page", Type: "int", Default: 1},
				{Name: "page-size", Description: "Rows per page, 1 to 100", Type: "int", Default: 25},
			}}},
	})
	Register(&Domain{
		Name: "project-requirement", Aliases: []string{"project-requirements"},
		Description: "Manage structured requirements with exact planned verification and retained history", APIPath: "/api/projects",
		Actions: []Action{
			{Name: "list", Description: "List visible requirements, optionally scoped to a deliverable and including retired records",
				ToolName: "UteamupProjectRequirementsList", HTTPMethod: "GET", RESTPath: "{projectGuid}/requirements", Args: projectGUIDArgument,
				Flags: append(projectScopePageFlags(50),
					FlagDef{Name: "deliverable-guid", Description: "Optional owning deliverable GUID", Type: "non-empty-uuid", QueryName: "deliverableGuid"},
					FlagDef{Name: "include-retired", Description: "Include retired requirements", Type: "bool", Default: false, QueryName: "includeRetired"})},
			{Name: "get", Description: "Read current values, approval binding and edit token",
				ToolName: "UteamupProjectRequirementGet", HTTPMethod: "GET", RESTPath: "{projectGuid}/requirements/{requirementGuid}",
				Args: projectResourceArguments("requirementGuid", "Requirement GUID")},
			{Name: "history", Description: "Read immutable revisions with current evidence permissions",
				ToolName: "UteamupProjectRequirementHistory", HTTPMethod: "GET", RESTPath: "{projectGuid}/requirements/{requirementGuid}/history",
				Args: projectResourceArguments("requirementGuid", "Requirement GUID"), Flags: projectScopePageFlags(20)},
			{Name: "create", Description: "Create unapproved scope with stable requestGuid; verification links do not mean tests passed",
				ToolName: "UteamupProjectRequirementCreate", HTTPMethod: "POST", RESTPath: "{projectGuid}/requirements", Args: projectGUIDArgument,
				Flags: []FlagDef{projectRequestFile()}},
			{Name: "update", Description: "Edit or retire unapproved scope; preserve expectedUpdatedAt and omitted versus empty evidence",
				ToolName: "UteamupProjectRequirementUpdate", HTTPMethod: "PUT", RESTPath: "{projectGuid}/requirements/{requirementGuid}",
				Args: projectResourceArguments("requirementGuid", "Requirement GUID"), Flags: []FlagDef{projectRequestFile()}},
			{Name: "verification", Description: "Read the exact inspection step's plan, revision and criterion",
				ToolName: "UteamupProjectVerificationReferenceGet", HTTPMethod: "GET", RESTPath: "{projectGuid}/requirements/verification-reference/{stepGuid}",
				Args: projectResourceArguments("stepGuid", "Exact inspection-plan step GUID")},
		},
	})
}
