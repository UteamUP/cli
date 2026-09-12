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
		Description: "Read current governed requirements and exact verification references", APIPath: "/api/projects",
		Actions: []Action{
			{Name: "get", Description: "Read current values, approval binding and edit token",
				ToolName: "UteamupProjectRequirementGet", RESTPath: "{projectGuid}/requirements/{requirementGuid}",
				Args: projectResourceArguments("requirementGuid", "Requirement GUID")},
			{Name: "verification", Description: "Read the exact inspection step's plan, revision and criterion",
				ToolName: "UteamupProjectVerificationReferenceGet", RESTPath: "{projectGuid}/requirements/verification-reference/{stepGuid}",
				Args: projectResourceArguments("stepGuid", "Exact inspection-plan step GUID")},
		},
	})
}
