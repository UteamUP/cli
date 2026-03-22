package registry

func init() {
	Register(&Domain{
		Name:        "asset-calibration",
		Description: "Manage asset calibration records",
		Actions: []Action{
			{Name: "list", Description: "List calibrations for an asset", ToolName: "UteamupAssetCalibrationList", Flags: append(paginationFlags(), FlagDef{Name: "asset-id", Description: "Asset ID", Type: "int"})},
			{Name: "get", Description: "Get calibration by ID", ToolName: "UteamupAssetCalibrationGet", Args: idArg()},
			{Name: "create", Description: "Create a calibration record", ToolName: "UteamupAssetCalibrationCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a calibration record", ToolName: "UteamupAssetCalibrationUpdate", Args: idArg(), Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete a calibration record", ToolName: "UteamupAssetCalibrationDelete", Args: idArg()},
		},
	})
}
