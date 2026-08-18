package registry

func init() {
	Register(&Domain{
		Name:        "asset-certification",
		Description: "Manage asset certifications",
		APIPath:     "/api/assetcertification",
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
			{
				Name:        "list",
				Description: "List certifications for an asset by public GUID",
				ToolName:    "UteamupAssetcertificationGetByAssetGuid",
				HTTPMethod:  "GET",
				RESTPath:    "asset/by-guid/{assetGuid}",
				Args: []ArgDef{
					{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get a certification by public GUID",
				ToolName:    "UteamupAssetcertificationGetByGuid",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{certificationGuid}",
				Args: []ArgDef{
					{Name: "certificationGuid", Description: "Public certification GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "create",
				Description: "Create a certification",
				ToolName:    "UteamupAssetcertificationCreate",
				HTTPMethod:  "POST",
				Flags:       []FlagDef{jsonFlag()},
			},
			{
				Name:        "update",
				Description: "Update a certification by public GUID",
				ToolName:    "UteamupAssetcertificationUpdateByGuid",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{certificationGuid}",
				Args: []ArgDef{
					{Name: "certificationGuid", Description: "Public certification GUID", Required: true, Type: "string"},
				},
				Flags: []FlagDef{jsonFlag()},
			},
			{
				Name:        "delete",
				Description: "Delete a certification by public GUID",
				ToolName:    "UteamupAssetcertificationDeleteByGuid",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{certificationGuid}",
				Args: []ArgDef{
					{Name: "certificationGuid", Description: "Public certification GUID", Required: true, Type: "string"},
				},
			},
		},
	})
}
