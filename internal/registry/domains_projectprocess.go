package registry

func init() {
	Register(&Domain{
		Name: "project-process", Description: "Review and explicitly adopt project lifecycle rules", APIPath: "/api/projects",
		Actions: []Action{
			{Name: "defaults", Description: "Read editable defaults without adopting them", ToolName: "UteamupProjectProcessDefaults",
				HTTPMethod: "GET", RESTPath: "{projectGuid}/process/defaults", Args: projectGUIDArgument,
				Flags: []FlagDef{{Name: "profile", Description: "general or equipmentDelivery", Type: "string", Default: "equipmentDelivery"}}},
			{Name: "get", Description: "Read current process or explicit legacy state", ToolName: "UteamupProjectProcessGet",
				HTTPMethod: "GET", RESTPath: "{projectGuid}/process", Args: projectGUIDArgument},
			{Name: "revision", Description: "Read an exact retained revision", ToolName: "UteamupProjectProcessRevision",
				HTTPMethod: "GET", RESTPath: "{projectGuid}/process/revisions/{revisionGuid}", Args: projectResourceArguments("revisionGuid", "Exact process revision GUID")},
			{Name: "history", Description: "List immutable adoption history", ToolName: "UteamupProjectProcessHistory",
				HTTPMethod: "GET", RESTPath: "{projectGuid}/process/history", Args: projectGUIDArgument,
				Flags: []FlagDef{{Name: "page", Description: "One-based page", Type: "int", Default: 1},
					{Name: "page-size", Description: "Rows per page, 1 to 100", Type: "int", Default: 25}}},
			{Name: "preview", Description: "Review complete rules and exact stage name/order changes without writing",
				ToolName: "UteamupProjectProcessPreview", HTTPMethod: "POST", RESTPath: "{projectGuid}/process/preview",
				Args: projectGUIDArgument, Flags: []FlagDef{projectRequestFile()}},
			{Name: "adopt", Description: "Adopt the human-reviewed proposal; retain requestGuid and expectedReviewFingerprint for retries",
				ToolName: "UteamupProjectProcessAdopt", HTTPMethod: "POST", RESTPath: "{projectGuid}/process/adopt",
				Args: projectGUIDArgument, Flags: []FlagDef{projectRequestFile()}},
		},
	})
}
