package registry

func init() {
	Register(&Domain{
		Name:        "driver",
		Aliases:     []string{"drivers"},
		Description: "Manage GUID-first drivers and licenses",
		Actions: []Action{
			{Name: "list", Description: "List drivers", ToolName: "UteamupDriverList", Flags: append(paginationFlags(),
				FlagDef{Name: "name-filter", Description: "Filter by driver name", Type: "string"},
				FlagDef{Name: "sort-by", Description: "Sort field", Type: "string"},
				FlagDef{Name: "sort-order", Description: "Sort order", Type: "string"},
				FlagDef{Name: "include-archived", Description: "Include archived drivers", Type: "bool"},
			)},
			{Name: "get", Description: "Get a driver by GUID", ToolName: "UteamupDriverGet", Args: []ArgDef{{Name: "driverGuid", Description: "Public driver GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{driverGuid}"},
			{Name: "create", Description: "Create a driver", ToolName: "UteamupDriverCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a driver by GUID", ToolName: "UteamupDriverUpdate", Args: []ArgDef{{Name: "driverGuid", Description: "Public driver GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{driverGuid}", Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete a driver by GUID", ToolName: "UteamupDriverDelete", Args: []ArgDef{{Name: "driverGuid", Description: "Public driver GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{driverGuid}"},
			{Name: "archive", Description: "Archive a driver by GUID", ToolName: "UteamupDriverArchive", HTTPMethod: "POST", Args: []ArgDef{{Name: "driverGuid", Description: "Public driver GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{driverGuid}/archive"},
			{Name: "licenses", Description: "List driver licenses", ToolName: "UteamupDriverGetLicenses", Args: []ArgDef{{Name: "driverGuid", Description: "Public driver GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{driverGuid}/licenses"},
			{Name: "license-add", Description: "Add a driver license", ToolName: "UteamupDriverAddLicense", Args: []ArgDef{{Name: "driverGuid", Description: "Public driver GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{driverGuid}/licenses", Flags: []FlagDef{jsonFlag()}},
			{Name: "license-update", Description: "Update a driver license", ToolName: "UteamupDriverUpdateLicense", Args: []ArgDef{{Name: "driverGuid", Description: "Public driver GUID", Required: true, Type: "string"}, {Name: "licenseGuid", Description: "Public license GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{driverGuid}/licenses/{licenseGuid}", Flags: []FlagDef{jsonFlag()}},
			{Name: "license-delete", Description: "Delete a driver license", ToolName: "UteamupDriverDeleteLicense", Args: []ArgDef{{Name: "driverGuid", Description: "Public driver GUID", Required: true, Type: "string"}, {Name: "licenseGuid", Description: "Public license GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{driverGuid}/licenses/{licenseGuid}"},
			{Name: "licenses-expiring", Description: "List expiring driver licenses", ToolName: "UteamupDriverGetExpiringLicenses", RESTPath: "licenses/expiring", Flags: []FlagDef{{Name: "days-from-now", Description: "Days from now", Type: "int"}}},
		},
	})

	assignmentGUIDArg := []ArgDef{{Name: "assignmentGuid", Description: "Public driver-assignment GUID", Required: true, Type: "string"}}
	Register(&Domain{
		Name: "driver-assignment", Aliases: []string{"da"}, Description: "Manage GUID-first driver assignments",
		Actions: []Action{
			{Name: "list", Description: "List driver assignments", ToolName: "UteamupDriverAssignmentList", Flags: paginationFlags()},
			{Name: "get", Description: "Get a driver assignment by GUID", ToolName: "UteamupDriverAssignmentGet", Args: assignmentGUIDArg, RESTPath: "by-guid/{assignmentGuid}"},
			{Name: "create", Description: "Create a driver assignment", ToolName: "UteamupDriverAssignmentCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a driver assignment by GUID", ToolName: "UteamupDriverAssignmentUpdate", Args: assignmentGUIDArg, RESTPath: "by-guid/{assignmentGuid}", Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete a driver assignment by GUID", ToolName: "UteamupDriverAssignmentDelete", Args: assignmentGUIDArg, RESTPath: "by-guid/{assignmentGuid}"},
			{Name: "end", Description: "End a driver assignment by GUID", ToolName: "UteamupDriverAssignmentEnd", HTTPMethod: "PUT", Args: assignmentGUIDArg, RESTPath: "by-guid/{assignmentGuid}/end"},
			{Name: "current", Description: "Get the current driver assignment for an asset", ToolName: "UteamupDriverAssignmentGetCurrentDriver", Args: []ArgDef{{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "string"}}, RESTPath: "asset/by-guid/{assetGuid}/current"},
		},
	})
	inspectionGUIDArg := []ArgDef{{Name: "inspectionGuid", Description: "Public vehicle inspection GUID", Required: true, Type: "string"}}
	vehicleInspectionActions := []Action{
		{Name: "list", Description: "List vehicle inspections", ToolName: "UteamupVehicleInspectionList", Flags: append(paginationFlags(), FlagDef{Name: "asset-guid", Description: "Filter by public asset GUID", Type: "string"})},
		{Name: "get", Description: "Get a vehicle inspection by GUID", ToolName: "UteamupVehicleInspectionGet", Args: inspectionGUIDArg, RESTPath: "by-guid/{inspectionGuid}"},
		{Name: "create", Description: "Create a vehicle inspection", ToolName: "UteamupVehicleInspectionCreate", Flags: []FlagDef{jsonFlag()}},
		{Name: "update", Description: "Update a vehicle inspection by GUID", ToolName: "UteamupVehicleInspectionUpdate", Args: inspectionGUIDArg, RESTPath: "by-guid/{inspectionGuid}", Flags: []FlagDef{jsonFlag()}},
		{Name: "delete", Description: "Delete a vehicle inspection by GUID", ToolName: "UteamupVehicleInspectionDelete", Args: inspectionGUIDArg, RESTPath: "by-guid/{inspectionGuid}"},
		{Name: "submit-items", Description: "Submit vehicle inspection results by GUID", ToolName: "UteamupVehicleInspectionSubmitItems", HTTPMethod: "POST", Args: inspectionGUIDArg, RESTPath: "by-guid/{inspectionGuid}/items", Flags: []FlagDef{jsonFlag()}},
		{
			Name:        "complete",
			Description: "Complete a vehicle inspection by GUID",
			ToolName:    "UteamupVehicleInspectionComplete",
			HTTPMethod:  "POST",
			Args:        inspectionGUIDArg,
			RESTPath:    "by-guid/{inspectionGuid}/complete",
			Flags: []FlagDef{
				jsonFlag(),
				{Name: "create-corrective-workorder", BodyName: "createCorrectiveWorkorder", Description: "Explicitly create one corrective work order for failed items", Type: "bool"},
			},
		},
		{
			Name:        "overdue",
			Description: "List tenant vehicles with overdue daily inspections",
			ToolName:    "UteamupVehicleInspectionGetOverdue",
			HTTPMethod:  "GET",
			RESTPath:    "overdue",
		},
	}
	Register(&Domain{
		Name:        "vehicle-inspection",
		Aliases:     []string{"vi"},
		Description: "Manage vehicle inspections",
		Actions:     vehicleInspectionActions,
	})
	Register(&Domain{
		Name:        "fuel-transaction",
		Aliases:     []string{"fuel"},
		Description: "Manage GUID-first fuel transactions",
		Actions: []Action{
			{
				Name: "list", Description: "List fuel transactions", ToolName: "UteamupFuelTransactionList",
				Flags: append(paginationFlags(),
					FlagDef{Name: "asset-guid", Description: "Filter by public asset GUID", Type: "string"},
					FlagDef{Name: "date-from", Description: "Optional UTC period start", Type: "string"},
					FlagDef{Name: "date-to", Description: "Optional UTC period end", Type: "string"},
				),
			},
			{
				Name: "get", Description: "Get a fuel transaction by GUID", ToolName: "UteamupFuelTransactionGet",
				Args: []ArgDef{{Name: "transactionGuid", Description: "Public fuel transaction GUID", Required: true, Type: "string"}},
			},
			{Name: "create", Description: "Create a fuel transaction", ToolName: "UteamupFuelTransactionCreate", Flags: []FlagDef{jsonFlag()}},
			{
				Name: "update", Description: "Update a fuel transaction by GUID", ToolName: "UteamupFuelTransactionUpdate",
				Args:  []ArgDef{{Name: "transactionGuid", Description: "Public fuel transaction GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{jsonFlag()},
			},
			{
				Name: "delete", Description: "Delete a fuel transaction by GUID", ToolName: "UteamupFuelTransactionDelete",
				Args: []ArgDef{{Name: "transactionGuid", Description: "Public fuel transaction GUID", Required: true, Type: "string"}},
			},
			{
				Name: "summary", Description: "Summarize fuel usage for an asset", ToolName: "UteamupFuelTransactionGetSummary",
				Args: []ArgDef{{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "date-from", Description: "Optional UTC period start", Type: "string"},
					{Name: "date-to", Description: "Optional UTC period end", Type: "string"},
				},
			},
			{
				Name: "efficiency", Description: "Calculate fuel efficiency for an asset", ToolName: "UteamupFuelTransactionGetEfficiency",
				Args: []ArgDef{{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "string"}},
			},
		},
	})
	Register(&Domain{Name: "fleet-dashboard", Description: "View fleet dashboard data", Actions: []Action{
		{Name: "get", Description: "Get the fleet dashboard summary", ToolName: "UteamupFleetDashboardGet"},
		{Name: "utilization", Description: "Get GUID-first vehicle utilization", ToolName: "UteamupFleetDashboardGetUtilization"},
		{Name: "compliance", Description: "Get GUID-first fleet compliance", ToolName: "UteamupFleetDashboardGetCompliance"},
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
