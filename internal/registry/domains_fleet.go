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
		{Name: "propose-maintenance", Description: "Prepare a governed maintenance proposal from fleet evidence", ToolName: "UteamupFleetMaintenancePropose", RESTBasePath: "/api/upmateassistant/fleet", RESTPath: "maintenance-proposals", HTTPMethod: "POST", Flags: []FlagDef{
			{Name: "source-type", BodyName: "sourceType", Description: "vehicle-inspection, telematics-event, or asset-maintenance-package", Required: true, Type: "string"},
			{Name: "source-guid", BodyName: "sourceGuid", Description: "Public GUID of the inspection, DTC event, or asset package", Required: true, Type: "string"},
			{Name: "idempotency-key", Description: "Stable retry key", Required: true, Type: "string", HeaderName: "Idempotency-Key"},
		}},
		{Name: "costs", Description: "Get fleet costs for an optional date range", ToolName: "UteamupFleetDashboardGetCosts", Flags: []FlagDef{
			{Name: "date-from", Description: "Optional UTC period start", Type: "string"},
			{Name: "date-to", Description: "Optional UTC period end", Type: "string"},
		}},
	}})
	Register(&Domain{
		Name:        "fleet-contact",
		Description: "Manage GUID-first fleet asset contact associations",
		Actions: []Action{
			{
				Name: "list", Description: "List contacts linked to an asset", ToolName: "UteamupFleetAssetContactList",
				Args: []ArgDef{{Name: "assetExternalGuid", Description: "Public asset GUID", Required: true, Type: "string"}},
			},
			{
				Name: "add", Description: "Link a contact to an asset", ToolName: "UteamupFleetAssetContactAdd",
				Args:  []ArgDef{{Name: "assetExternalGuid", Description: "Public asset GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{jsonFlag()},
			},
			{
				Name: "delete", Description: "Remove a contact association from an asset", ToolName: "UteamupFleetAssetContactDelete",
				Args: []ArgDef{
					{Name: "assetExternalGuid", Description: "Public asset GUID", Required: true, Type: "string"},
					{Name: "associationExternalGuid", Description: "Public asset-contact association GUID", Required: true, Type: "string"},
				},
			},
		},
	})
}
