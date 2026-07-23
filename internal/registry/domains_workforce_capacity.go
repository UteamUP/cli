package registry

// Workforce capacity planning (CAP-03 readiness snapshot + CAP-14 saved scenarios).
// Deterministic, GUID-first, tenant-scoped; charges no AI credits.
func init() {
	Register(&Domain{
		Name:        "capacity",
		Aliases:     []string{"workforce-capacity"},
		Description: "Workforce capacity data-readiness snapshot",
		APIPath:     "/api/workforcecapacity",
		Actions: []Action{
			{
				Name:        "readiness",
				Description: "Show workforce setup counts behind a forecast so 0 supply can be explained",
				ToolName:    "UteamupWorkforceCapacityReadiness",
				HTTPMethod:  "GET",
				RESTPath:    "readiness",
			},
		},
	})

	Register(&Domain{
		Name:        "capacity-scenario",
		Aliases:     []string{"capacity-scenarios"},
		Description: "Saved workforce capacity-planning scenarios (save, load, clone, compare)",
		APIPath:     "/api/workforcecapacityscenarios",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the tenant's saved capacity scenarios",
				ToolName:    "UteamupWorkforceCapacityScenarioList",
				HTTPMethod:  "GET",
			},
			{
				Name:        "get",
				Description: "Read one saved scenario (assumptions + snapshot) by public GUID",
				ToolName:    "UteamupWorkforceCapacityScenarioGet",
				HTTPMethod:  "GET",
				RESTPath:    "{scenarioGuid}",
				Args: []ArgDef{
					{Name: "scenarioGuid", Description: "Scenario GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Save the current scenario",
				ToolName:    "UteamupWorkforceCapacityScenarioCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "name", BodyName: "name", Description: "Scenario name", Required: true, Type: "string"},
					{Name: "description", BodyName: "description", Description: "Optional description", Type: "string"},
					{Name: "planning-horizon-from", BodyName: "planningHorizonFrom", Description: "Planning horizon start (ISO date)", Required: true, Type: "string"},
					{Name: "planning-horizon-to", BodyName: "planningHorizonTo", Description: "Planning horizon end (ISO date)", Required: true, Type: "string"},
					{Name: "assumptions-json", BodyName: "assumptionsJson", Description: "Assumptions JSON", Type: "string", Default: "{}"},
					{Name: "snapshot-json", BodyName: "snapshotJson", Description: "Demand/supply/forecast snapshot JSON", Type: "string", Default: "{}"},
				},
			},
			{
				Name:        "clone",
				Description: "Clone a scenario as a new Draft to compare alternatives",
				ToolName:    "UteamupWorkforceCapacityScenarioClone",
				HTTPMethod:  "POST",
				RESTPath:    "{scenarioGuid}/clone",
				Args: []ArgDef{
					{Name: "scenarioGuid", Description: "Scenario GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a saved scenario by public GUID",
				ToolName:    "UteamupWorkforceCapacityScenarioDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "{scenarioGuid}",
				Args: []ArgDef{
					{Name: "scenarioGuid", Description: "Scenario GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
