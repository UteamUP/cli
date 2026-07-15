package registry

func init() {
	Register(&Domain{
		Name:        "driver",
		Aliases:     []string{"drivers"},
		Description: "Manage drivers",
		Actions: append(crudActions("Driver"), Action{
			Name: "search", Description: "Search drivers", ToolName: "UteamupDriverSearch", Args: queryArg(), Flags: paginationFlags(),
		}),
	})

	Register(&Domain{Name: "driver-assignment", Aliases: []string{"da"}, Description: "Manage driver assignments", Actions: crudActions("DriverAssignment")})
	Register(&Domain{Name: "vehicle-inspection", Aliases: []string{"vi"}, Description: "Manage vehicle inspections", Actions: crudActions("VehicleInspection")})
	Register(&Domain{Name: "fuel-transaction", Aliases: []string{"fuel"}, Description: "Manage fuel transactions", Actions: crudActions("FuelTransaction")})
	Register(&Domain{Name: "fleet-dashboard", Description: "View fleet dashboard data", Actions: []Action{
		{Name: "get", Description: "Get the fleet dashboard summary", ToolName: "UteamupFleetDashboardGet"},
		{Name: "costs", Description: "Get fleet costs for an optional date range", ToolName: "UteamupFleetDashboardGetCosts", Flags: []FlagDef{
			{Name: "date-from", Description: "Optional UTC period start", Type: "string"},
			{Name: "date-to", Description: "Optional UTC period end", Type: "string"},
		}},
	}})
	Register(&Domain{Name: "fleet-contact", Description: "Manage fleet asset contacts", Actions: crudActions("FleetAssetContact")})
}
