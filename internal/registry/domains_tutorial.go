package registry

func init() {
	Register(&Domain{
		Name:        "tutorial",
		Aliases:     []string{"tutorials", "guide"},
		Description: "Discover published UPMate guided tutorials across modules",
		APIPath:    "/api/tutorials",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List published tutorials across modules or filter by module",
				ToolName:    "UteamupTutorialList",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "module", Description: "Optional module key; omit to discover all available tutorials", Type: "string"},
					{Name: "platform", Description: "Client platform: web or mobile", Type: "string", Default: "web"},
				},
			},
			{
				Name:        "get",
				Description: "Get one published tutorial by its stable semantic ID",
				ToolName:    "UteamupTutorialGet",
				HTTPMethod:  "GET",
				RESTPath:    "{tutorialId}",
				Args: []ArgDef{
					{Name: "tutorialId", Description: "Stable tutorial ID, for example workorders.create", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "platform", Description: "Client platform: web or mobile", Type: "string", Default: "web"},
				},
			},
		},
	})
}
