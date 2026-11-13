package registry

func init() {
	Register(&Domain{
		Name:        "asset-calibration",
		Description: "Manage asset calibration records",
		APIPath:     "/api/assetcalibration",
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
			{
				Name:        "list",
				Description: "List calibrations for an asset by public GUID",
				ToolName:    "UteamupAssetcalibrationGetByAsset",
				HTTPMethod:  "GET",
				RESTPath:    "asset/by-guid/{assetGuid}",
				Args: []ArgDef{
					{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "list-for-tool",
				Description: "List calibrations for a tool by public GUID",
				ToolName:    "UteamupAssetcalibrationGetByToolGuid",
				HTTPMethod:  "GET",
				RESTPath:    "tool/by-guid/{toolGuid}",
				Args: []ArgDef{
					{Name: "toolGuid", Description: "Public tool GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "get",
				Description: "Get a calibration by public GUID",
				ToolName:    "UteamupAssetcalibrationGet",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{calibrationGuid}",
				Args: []ArgDef{
					{Name: "calibrationGuid", Description: "Public calibration GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "create",
				Description: "Create a calibration for an asset (--asset-guid) or a tool (--tool-guid); set exactly one",
				ToolName:    "UteamupAssetcalibrationCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					jsonFlag(),
					{Name: "asset-guid", Description: "Public asset GUID (exactly one of --asset-guid and --tool-guid)", Type: "uuid", BodyName: "assetGuid"},
					{Name: "tool-guid", Description: "Public tool GUID (exactly one of --asset-guid and --tool-guid)", Type: "uuid", BodyName: "toolGuid"},
				},
			},
			{
				Name:        "update",
				Description: "Update a calibration by public GUID",
				ToolName:    "UteamupAssetcalibrationUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{calibrationGuid}",
				Args: []ArgDef{
					{Name: "calibrationGuid", Description: "Public calibration GUID", Required: true, Type: "string"},
				},
				Flags: []FlagDef{jsonFlag()},
			},
			{
				Name:        "delete",
				Description: "Delete a calibration by public GUID",
				ToolName:    "UteamupAssetcalibrationDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{calibrationGuid}",
				Args: []ArgDef{
					{Name: "calibrationGuid", Description: "Public calibration GUID", Required: true, Type: "string"},
				},
			},
		},
	})
}
