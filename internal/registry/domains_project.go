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
		),
	})
}
