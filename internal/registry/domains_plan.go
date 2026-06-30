package registry

func init() {
	Register(&Domain{
		Name:        "plan",
		Aliases:     []string{"plans"},
		Description: "Manage subscription plans",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List all subscription plans",
				ToolName:    "UteamupPlanList",
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get plan details by its stable GUID",
				ToolName:    "UteamupPlanGet",
				RESTPath:    "by-guid/{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "Plan GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
		},
	})
}
