package registry

// Mirrors the MCP UteamupUtilityMeterArchive / UteamupUtilityMeterRestore tools backed by
// UtilityMeterController. A meter with readings is marked inactive instead of deleted: it leaves meter
// lists and reading pickers and keeps every reading. Both actions need UtilityMeter.Delete.
//
//	POST /api/utility-meter/by-guid/{meterGuid}/archive  — mark inactive
//	POST /api/utility-meter/by-guid/{meterGuid}/restore  — reactivate
func init() {
	Register(&Domain{
		Name:        "utility-meter",
		Aliases:     []string{"utility-meters"},
		Description: "Mark utility meters inactive and reactivate them (Guid-keyed)",
		APIPath:     "/api/utility-meter",
		Actions: []Action{
			{
				Name: "archive",
				Description: "Mark a utility meter inactive: it leaves meter lists and reading pickers and keeps every reading. " +
					"Use this for a meter with readings; deleting it is refused",
				ToolName:        "UteamupUtilityMeterArchive",
				HTTPMethod:      "POST",
				RESTPath:        "by-guid/{meterGuid}/archive",
				RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "meterGuid", Description: "GUID of the meter to mark inactive", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name:            "restore",
				Description:     "Reactivate an inactive utility meter so it shows in meter lists and reading pickers again",
				ToolName:        "UteamupUtilityMeterRestore",
				HTTPMethod:      "POST",
				RESTPath:        "by-guid/{meterGuid}/restore",
				RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "meterGuid", Description: "GUID of the inactive meter", Required: true, Type: "non-empty-uuid"},
				},
			},
		},
	})
}
