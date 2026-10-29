package registry

func commentAssetActions() []Action {
	asset := []ArgDef{{Name: "assetGuid", Description: "Asset public GUID", Type: "non-empty-uuid", Required: true}}
	fields := jsonFlag()
	fields.Required = true
	return []Action{
		{Name: "list-by-asset", Description: "Read a bounded page of asset messages or activity replies", ToolName: "UteamupCommentListByAssetPaginated", HTTPMethod: "GET", RESTPath: "asset/{assetGuid}/paginated", Args: asset,
			Flags: []FlagDef{
				{Name: "page", Description: "Page number", Type: "int", Default: 1},
				{Name: "page-size", Description: "Page size (1 to 100)", QueryName: "pageSize", Type: "int", Default: 20},
				{Name: "activity-log-guid", Description: "Filter replies to one activity on this asset", QueryName: "activityLogGuid", Type: "non-empty-uuid"},
			}},
		{Name: "create-for-asset", Description: "Post an asset message with optional activityLogGuid in the JSON body", ToolName: "UteamupCommentCreate", HTTPMethod: "POST", RESTPath: "asset/{assetGuid}", Args: asset, Flags: []FlagDef{fields}},
	}
}
