package registry

func projectScopePageFlags(pageSize int) []FlagDef {
	return []FlagDef{
		{Name: "page", Description: "One-based page", Type: "int", Default: 1, QueryName: "page"},
		{Name: "page-size", Description: "Rows per page, 1 to 100", Type: "int", Default: pageSize, QueryName: "pageSize"},
	}
}

func init() {
	intakeArgs := projectResourceArguments("intakeGuid", "Source intake GUID")
	Register(&Domain{
		Name: "project-intake", Aliases: []string{"project-intakes"},
		Description: "Prepare and reconcile exact source scope before reviewed approval", APIPath: "/api/projects",
		Actions: []Action{
			{Name: "list", Description: "List bounded source summaries; open a draft for reconciliation",
				ToolName: "UteamupProjectIntakesList", HTTPMethod: "GET", RESTPath: "{projectGuid}/intake", Args: projectGUIDArgument,
				Flags: projectScopePageFlags(25)},
			{Name: "get", Description: "Read the current draft, exact source version and discrepancies",
				ToolName: "UteamupProjectIntakeGet", HTTPMethod: "GET", RESTPath: "{projectGuid}/intake/{intakeGuid}", Args: intakeArgs},
			{Name: "history", Description: "Read retained mappings and review decisions in bounded pages",
				ToolName: "UteamupProjectIntakeHistory", HTTPMethod: "GET", RESTPath: "{projectGuid}/intake/{intakeGuid}/history", Args: intakeArgs,
				Flags: projectScopePageFlags(10)},
			{Name: "source-preview", Description: "Extract a verified exact document version into an editable proposal without saving",
				ToolName: "UteamupProjectIntakeSourcePreview", HTTPMethod: "POST", RESTPath: "{projectGuid}/intake/source-preview", Args: projectGUIDArgument,
				Flags: []FlagDef{projectRequestFile()}},
			{Name: "create", Description: "Create or reopen an exact source draft; preserve requestGuid, row identities and source checks",
				ToolName: "UteamupProjectIntakeCreate", HTTPMethod: "POST", RESTPath: "{projectGuid}/intake", Args: projectGUIDArgument,
				Flags: []FlagDef{projectRequestFile()}},
			{Name: "update", Description: "Save a complete editable draft with expectedUpdatedAt and the original requestGuid",
				ToolName: "UteamupProjectIntakeUpdate", HTTPMethod: "PUT", RESTPath: "{projectGuid}/intake/{intakeGuid}", Args: intakeArgs,
				Flags: []FlagDef{projectRequestFile()}},
			{Name: "review", Description: "Record human reconciliation review; approval and scope application remain separate",
				ToolName: "UteamupProjectIntakeReview", HTTPMethod: "POST", RESTPath: "{projectGuid}/intake/{intakeGuid}/review", Args: intakeArgs,
				Flags: []FlagDef{projectRequestFile()}},
			{Name: "apply-preview", Description: "Review selected intakes, exact proposed scope, sales shares and baseline predecessor",
				ToolName: "UteamupProjectScopeApplyPreview", HTTPMethod: "POST", RESTPath: "{projectGuid}/intake/apply-preview", Args: projectGUIDArgument,
				Flags: []FlagDef{projectRequestFile()}},
			{Name: "apply", Description: "Apply explicitly approved scope; retain requestGuid, original tokens, review fingerprint and approval note",
				ToolName: "UteamupProjectScopeApply", HTTPMethod: "POST", RESTPath: "{projectGuid}/intake/apply", Args: projectGUIDArgument,
				Flags: []FlagDef{projectRequestFile()}},
		},
	})
}
