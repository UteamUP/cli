package registry

func init() {
	projectArgs := []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "non-empty-uuid"}}
	edgeArgs := []ArgDef{
		{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "non-empty-uuid"},
		{Name: "dependencyGuid", Description: "Existing dependency GUID", Required: true, Type: "non-empty-uuid"},
	}
	reviewPayload := projectRequestFile()
	reviewPayload.Required = true
	reviewPayload.Description = "Complete reviewed request file, including stable requestGuid and current fingerprints when applying changes"
	Register(&Domain{Name: "project-schedule", Description: "Review project dates, working calendars and operational dependencies", APIPath: "/api/projects", Actions: []Action{
		{Name: "get", HTTPMethod: "GET", RESTPath: "{projectGuid}/timeline", Args: projectArgs,
			ToolName: "UteamupProjectScheduleGet", Description: "Read all four date bases, dependency graph and current review token"},
		{Name: "preview", HTTPMethod: "POST", RESTPath: "{projectGuid}/timeline/preview", Args: projectArgs, Flags: []FlagDef{reviewPayload},
			ToolName: "UteamupProjectSchedulePreview", Description: "Preview downstream dates and published booking evidence without saving"},
		{Name: "apply", HTTPMethod: "POST", RESTPath: "{projectGuid}/timeline/apply", Args: projectArgs, Flags: []FlagDef{reviewPayload},
			ToolName: "UteamupProjectScheduleApply", Description: "Apply the exact reviewed dates; preserve baseline and published bookings"},
		{Name: "calendar", HTTPMethod: "PUT", RESTPath: "{projectGuid}/timeline/calendar", Args: projectArgs, Flags: []FlagDef{reviewPayload},
			ToolName: "UteamupProjectScheduleCalendarSet", Description: "Adopt a reviewed working calendar and reload the schedule after its receipt"},
		{Name: "history", HTTPMethod: "GET", RESTPath: "{projectGuid}/timeline/history", Args: projectArgs,
			ToolName: "UteamupProjectScheduleHistory", Description: "Read retained reviews with current record permissions",
			Flags: []FlagDef{{Name: "page", Description: "One-based page", Type: "int", Default: 1},
				{Name: "page-size", Description: "Rows per page, 1 to 100", Type: "int", Default: 25}}},
		{Name: "dependency-create", HTTPMethod: "POST", RESTPath: "{projectGuid}/timeline/dependencies", Args: projectArgs, Flags: []FlagDef{reviewPayload},
			ToolName: "UteamupProjectScheduleDependencySave", Description: "Create a reviewed relationship between named project records"},
		{Name: "dependency-update", HTTPMethod: "PUT", RESTPath: "{projectGuid}/timeline/dependencies/{dependencyGuid}", Args: edgeArgs, Flags: []FlagDef{reviewPayload},
			ToolName: "UteamupProjectScheduleDependencySave", Description: "Change an existing relationship with its current graph token"},
		{Name: "dependency-remove", HTTPMethod: "POST", RESTPath: "{projectGuid}/timeline/dependencies/{dependencyGuid}/remove", Args: edgeArgs, Flags: []FlagDef{reviewPayload},
			ToolName: "UteamupProjectScheduleDependencyRemove", Description: "Remove a reviewed relationship while retaining exact history"},
	}})
}
