package registry

func init() {
	Register(&Domain{
		Name:        "project",
		Aliases:     []string{"projects"},
		Description: "Manage projects",
		Actions: append(crudActions("Project"),
			Action{Name: "search", Description: "Search projects", ToolName: "UteamupProjectSearch", Args: queryArg(), Flags: paginationFlags()},
		),
	})
}
