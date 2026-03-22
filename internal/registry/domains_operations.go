package registry

func init() {
	Register(&Domain{
		Name:        "route",
		Aliases:     []string{"routes", "operational-route"},
		Description: "Manage operational routes",
		Actions: append(crudActions("OperationalRoute"),
			Action{Name: "search", Description: "Search routes", ToolName: "UteamupOperationalRouteSearch", Args: queryArg(), Flags: paginationFlags()},
		),
	})

	Register(&Domain{Name: "automation", Aliases: []string{"automations"}, Description: "Manage automations", Actions: crudActions("Automation")})
	Register(&Domain{Name: "notification", Aliases: []string{"notifications"}, Description: "Manage notifications", Actions: crudActions("Notification")})
	Register(&Domain{Name: "helpdesk", Description: "Manage helpdesk", Actions: crudActions("Helpdesk")})
	Register(&Domain{Name: "extension", Aliases: []string{"extensions"}, Description: "Manage extensions", Actions: listGetActions("Extension")})
	Register(&Domain{Name: "weather", Description: "Get weather data", Actions: listGetActions("Weather")})
}
