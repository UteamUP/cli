package registry

func init() {
	Register(&Domain{
		Name:        "asset-calibration",
		Description: "Manage asset calibration records",
		Actions: []Action{
			{
				Name:        "due-soon",
				Description: "List calibrations due within a planning window",
				ToolName:    "UteamupAssetcalibrationGetDueSoon",
				HTTPMethod:  "GET",
				RESTPath:    "due-soon",
				Flags: []FlagDef{
					{Name: "days", Description: "Planning window in days (1-365)", Type: "int", Default: 30},
				},
			},
			{
				Name:        "overdue",
				Description: "List overdue tenant asset calibrations",
				ToolName:    "UteamupAssetcalibrationGetOverdue",
				HTTPMethod:  "GET",
				RESTPath:    "overdue",
			},
			{Name: "list", Description: "List calibrations for an asset", ToolName: "UteamupAssetCalibrationList", Flags: append(paginationFlags(), FlagDef{Name: "asset-id", Description: "Asset ID", Type: "int"})},
			{Name: "get", Description: "Get calibration by ID", ToolName: "UteamupAssetCalibrationGet", Args: idArg()},
			{Name: "create", Description: "Create a calibration record", ToolName: "UteamupAssetCalibrationCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a calibration record", ToolName: "UteamupAssetCalibrationUpdate", Args: idArg(), Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete a calibration record", ToolName: "UteamupAssetCalibrationDelete", Args: idArg()},
		},
	})
}
