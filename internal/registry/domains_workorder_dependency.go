package registry

func init() {
	Register(&Domain{
		Name:        "workorder-dependency",
		Aliases:     []string{"wo-dependency", "wo-dep"},
		Description: "Link prerequisite workorders; a workorder cannot be completed until its prerequisites are",
		APIPath:     "/api/workorder",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List what one workorder waits on and what waits on it",
				ToolName:    "UteamupWorkorderDependencyList",
				HTTPMethod:  "GET",
				RESTPath:    "{workorderGuid}/dependencies",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Tenant-scoped workorder GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Link a prerequisite workorder (rejects self-links, duplicates and loops)",
				ToolName:    "UteamupWorkorderDependencyCreate",
				HTTPMethod:  "POST",
				RESTPath:    "{workorderGuid}/dependencies",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Tenant-scoped GUID of the workorder that waits", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "prerequisite-workorder-guid", BodyName: "prerequisiteWorkorderGuid", Description: "GUID of the workorder to wait on", Type: "non-empty-uuid", Required: true},
				},
			},
			{
				Name:        "delete",
				Description: "Remove a standalone prerequisite link (project timeline links are edited in the project)",
				ToolName:    "UteamupWorkorderDependencyDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "{workorderGuid}/dependencies/{dependencyGuid}",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Tenant-scoped workorder GUID", Required: true, Type: "uuid"},
					{Name: "dependencyGuid", Description: "Dependency link GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
