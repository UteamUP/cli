package registry

func init() {
	Register(&Domain{
		Name:        "tutorial",
		Aliases:     []string{"tutorials", "guide"},
		Description: "Browse published UPMate guided tutorials",
		APIPath:    "/api/tutorials",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List published tutorials for a module and platform",
				ToolName:    "UteamupTutorialList",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "module", Description: "Module key, for example stock", Type: "string", Default: "stock"},
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
					{Name: "tutorialId", Description: "Stable tutorial ID, for example stock.orientation", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "platform", Description: "Client platform: web or mobile", Type: "string", Default: "web"},
				},
			},
		},
	})
}
