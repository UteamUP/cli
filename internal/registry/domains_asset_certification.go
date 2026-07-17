package registry

func init() {
	Register(&Domain{
		Name:        "asset-certification",
		Description: "Manage asset certifications",
		Actions: []Action{
			{
				Name:        "expired",
				Description: "List expired tenant asset certifications",
				ToolName:    "UteamupAssetcertificationGetExpired",
				HTTPMethod:  "GET",
				RESTPath:    "expired",
			},
			{
				Name:        "expiring",
				Description: "List active certifications expiring within a planning window",
				ToolName:    "UteamupAssetcertificationGetExpiringSoon",
				HTTPMethod:  "GET",
				RESTPath:    "expiring-soon",
				Flags: []FlagDef{
					{Name: "days", Description: "Planning window in days (1-365)", Type: "int", Default: 30},
				},
			},
			{Name: "list", Description: "List certifications", ToolName: "UteamupAssetCertificationList", Flags: append(paginationFlags(), FlagDef{Name: "asset-id", Description: "Asset ID", Type: "int"})},
			{Name: "get", Description: "Get certification by ID", ToolName: "UteamupAssetCertificationGet", Args: idArg()},
			{Name: "create", Description: "Create a certification", ToolName: "UteamupAssetCertificationCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a certification", ToolName: "UteamupAssetCertificationUpdate", Args: idArg(), Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete a certification", ToolName: "UteamupAssetCertificationDelete", Args: idArg()},
		},
	})
}
