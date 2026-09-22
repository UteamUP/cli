package registry

func init() {
	Register(&Domain{
		Name:        "property-damage",
		Aliases:     []string{"damage-report", "property-loss"},
		Description: "Report property loss from fires, leaks and other events, with or without injuries.",
		APIPath:     "/api/propertydamage",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List property damage reports",
				ToolName:    "UteamupPropertydamageList",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "skip", Description: "Skip offset", Default: 0, Type: "int", QueryName: "skip"},
					{Name: "take", Description: "Page size (server caps at 200)", Default: 50, Type: "int", QueryName: "take"},
				},
			},
			{
				Name:        "get",
				Description: "Get a property damage report by public GUID",
				ToolName:    "UteamupPropertydamageGet",
				RESTPath:    "by-guid/{guid}",
				HTTPMethod:  "GET",
				Args:        []ArgDef{{Name: "guid", Description: "Report GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "create",
				Description: "Create a property damage report with multiple items",
				ToolName:    "UteamupPropertydamageCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "JSON file containing PropertyDamageReportCreateModel", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "update",
				Description: "Update a property damage report and its item list",
				ToolName:    "UteamupPropertydamageUpdate",
				RESTPath:    "by-guid/{guid}",
				HTTPMethod:  "PUT",
				Args:        []ArgDef{{Name: "guid", Description: "Report GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "JSON file containing PropertyDamageReportUpdateModel", Required: true, Type: "string", JSONFile: true},
				},
			},
		},
	})
}
