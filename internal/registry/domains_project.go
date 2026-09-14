package registry

func init() {
	Register(&Domain{
		Name:        "project",
		Aliases:     []string{"projects"},
		Description: "Manage projects",
		// GUID-first CRUD: reads and deletion use the existing REST routes.
		// Creation and update use the typed MCP model because the corresponding
		// REST actions require multipart forms. Public identities remain GUIDs.
		Actions: []Action{
			{Name: "list", Description: "List projects", ToolName: "UteamupProjectList", Flags: paginationFlags()},
			{Name: "get", Description: "Get a project by GUID", ToolName: "UteamupProjectGet", Args: externalGUIDArg()},
			{Name: "create", Description: "Create a project from its complete model", ToolName: "UteamupProjectCreate", MCPOnly: true, Flags: []FlagDef{projectMCPModelFile()}},
			{Name: "update", Description: "Update a project by GUID with its complete model", ToolName: "UteamupProjectUpdate", MCPOnly: true,
				Args: []ArgDef{{Name: "externalGuid", BodyName: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}}, Flags: []FlagDef{projectMCPModelFile()}},
			{Name: "delete", Description: "Delete a project by GUID", ToolName: "UteamupProjectDelete", Args: externalGUIDArg()},
			{Name: "search", Description: "Search projects", ToolName: "UteamupProjectSearch", Args: queryArg(), Flags: paginationFlags()},
			// my-projects mirrors GET /api/project/my-projects — lists projects
			// containing workorders assigned to the authenticated user.
			// Requires backend MediatR handler registered for UteamupProjectMyProjects.
			{Name: "my-projects", Description: "List projects where the current user has assigned workorders", ToolName: "UteamupProjectMyProjects"},
			// GUID-keyed field setters on ProjectController. Both identifiers ride
			// the URL (no body) — the int-keyed originals are [Obsolete] on the
			// backend, so the CLI only exposes the by-guid routes.
			{
				Name:        "set-status",
				Description: "Set a project's status by GUID (0=Planning, 1=Active, 2=OnHold, 3=Completed, 4=Cancelled)",
				ToolName:    "UteamupProjectSetStatus",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{projectGuid}/status/{statusId}",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "statusId", Description: "New status (0=Planning, 1=Active, 2=OnHold, 3=Completed, 4=Cancelled)", Required: true, Type: "int"},
				},
			},
			{
				Name:        "set-priority",
				Description: "Set a project's priority by GUID (1=Low … 5=Critical)",
				ToolName:    "UteamupProjectSetPriority",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{projectGuid}/priority/{priorityId}",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "priorityId", Description: "New priority (1=Low, 2=Medium, 3=High, 4=Urgent, 5=Critical)", Required: true, Type: "int"},
				},
			},
			{
				Name:        "set-owner",
				Description: "Set a project's owner using public project and user GUIDs",
				ToolName:    "UteamupProjectSetOwner",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{projectGuid}/owner-guid/{ownerGuid}",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "ownerGuid", Description: "New owner's user GUID", Required: true, Type: "string"},
				},
			},
			// Project templates. The typed MCP tools own this flow and expose no REST
			// adapter, so these commands call the tools directly: every positional arg
			// and flag maps to the tool's exact camelCase argument name, and identities
			// cross the boundary as public GUIDs only.
			{Name: "templates", Description: "List project templates", ToolName: "UteamupProjectTemplateList", MCPOnly: true},
			{
				Name:        "from-template",
				Description: "Create a project from a project template, including workorders from the template's workorder templates",
				ToolName:    "UteamupProjectCreateFromTemplate",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "templateGuid", Description: "Project template GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "name", Description: "Name of the new project", Required: true, Type: "string"},
					{Name: "start-date-utc", BodyName: "startDateUtc", Description: "Project start as an ISO-8601 UTC date-time, for example 2026-10-01T08:00:00Z", Required: true, Type: "string"},
					{Name: "owner-guid", BodyName: "ownerGuid", Description: "Project owner's user GUID", Required: true, Type: "uuid"},
					{Name: "description", Description: "Optional project description", Type: "string"},
					{Name: "project-code", BodyName: "projectCode", Description: "Optional project code", Type: "string"},
					{Name: "end-date-utc", BodyName: "endDateUtc", Description: "Optional project end as an ISO-8601 UTC date-time", Type: "string"},
					{Name: "priority", Description: "Project priority (1=Low, 2=Medium, 3=High, 4=Urgent, 5=Critical)", Default: 0, Type: "int"},
					{Name: "budget", Description: "Project budget", Default: 0.0, Type: "float"},
					{Name: "customer-guid", BodyName: "customerGuid", Description: "Optional customer GUID", Type: "uuid"},
					{Name: "location-guid", BodyName: "locationGuid", Description: "Optional location GUID", Type: "uuid"},
					{Name: "asset-group-guid", BodyName: "assetGroupGuid", Description: "Optional asset group GUID", Type: "uuid"},
				},
			},
			{
				Name:        "save-as-template",
				Description: "Save a copy of a project as a new project template, keeping the selected workorders as template workorders",
				ToolName:    "UteamupProjectSaveAsTemplate",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "name", Description: "Name of the new project template", Required: true, Type: "string"},
					{Name: "project-code", BodyName: "projectCode", Description: "Optional project template code", Type: "string"},
					{Name: "description", Description: "Optional project template description", Type: "string"},
					{Name: "workorder-guids", BodyName: "workorderGuids", Description: "Workorder GUIDs to keep as template workorders (repeatable or comma-separated)", Type: "stringSlice"},
				},
			},
			{
				Name:        "template-add-workorder-templates",
				Description: "Copy library workorder templates into a project template",
				ToolName:    "UteamupProjectTemplateAddWorkorderTemplates",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project template GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "File containing the model: items of workorderTemplateGuid with optional projectStartOffsetDays and projectStageOrder (without a model wrapper)",
						Type: "string", Required: true, JSONFile: true},
				},
			},
		},
	})
}
