package registry

func init() {
	Register(&Domain{
		Name:        "project",
		Aliases:     []string{"projects"},
		Description: "Manage projects",
		Actions: append(crudActions("Project"),
			Action{Name: "search", Description: "Search projects", ToolName: "UteamupProjectSearch", Args: queryArg(), Flags: paginationFlags()},
			// my-projects mirrors GET /api/project/my-projects — lists projects
			// containing workorders assigned to the authenticated user.
			// Requires backend MediatR handler registered for UteamupProjectMyProjects.
			Action{Name: "my-projects", Description: "List projects where the current user has assigned workorders", ToolName: "UteamupProjectMyProjects"},
			// GUID-keyed field setters on ProjectController. Both identifiers ride
			// the URL (no body) — the int-keyed originals are [Obsolete] on the
			// backend, so the CLI only exposes the by-guid routes.
			Action{
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
			Action{
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
			Action{
				Name:        "set-owner",
				Description: "Set a project's owner by GUID (ownerId is the new owner's Identity user id)",
				ToolName:    "UteamupProjectSetOwner",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{projectGuid}/owner/{ownerId}",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "ownerId", Description: "New owner's Identity user id", Required: true, Type: "string"},
				},
			},
		),
	})
}
