package registry

func init() {
	Register(&Domain{
		Name: "asset-bom", Description: "Review attached PDF parts tables before applying asset BOM quantities",
		APIPath: "/api/asset",
		Actions: []Action{
			{
				Name: "sources", Description: "Read saved BOM quantities and exact PDF provenance separately from manual parts",
				ToolName: "UteamupAssetBomSources", HTTPMethod: "GET", RESTPath: "{assetGuid}/bom/sources", RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "assetGuid", Description: "assetGuid public GUID", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name: "review-create", Description: "Start or resume free review of a PDF already attached to this asset",
				ToolName: "UteamupAssetBomReviewCreate", HTTPMethod: "POST", RESTPath: "{assetGuid}/bom/review-sessions", RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "assetGuid", Description: "assetGuid public GUID", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{{Name: "document-guid", BodyName: "documentGuid", Description: "Attached active PDF document public GUID", Required: true, Type: "non-empty-uuid"}},
			},
			{
				Name: "review-get", Description: "Read original candidates, current revision, page evidence and quoted credit cost",
				ToolName: "UteamupAssetBomReviewGet", HTTPMethod: "GET", RESTPath: "{assetGuid}/bom/review-sessions/{sessionGuid}", RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "assetGuid", Description: "assetGuid public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "sessionGuid", Description: "sessionGuid public GUID", Required: true, Type: "non-empty-uuid"},
				},
			},
			{
				Name: "analyze", Description: "Queue selected pages using explicitly confirmed per-page and total AI-credit values",
				ToolName: "UteamupAssetBomReviewAnalyze", HTTPMethod: "POST", RESTPath: "{assetGuid}/bom/review-sessions/{sessionGuid}/analyze", RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "assetGuid", Description: "assetGuid public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "sessionGuid", Description: "sessionGuid public GUID", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{{Name: "from-json", Description: "Reviewed canonical request JSON; confirmation and credit values must be explicit", Required: true, Type: "string", RootJSONObjectFile: true}},
			},
			{
				Name: "patch-row", Description: "Save a reviewed row, observed revision and explicit part choice",
				ToolName: "UteamupAssetBomReviewPatchRow", HTTPMethod: "PATCH", RESTPath: "{assetGuid}/bom/review-sessions/{sessionGuid}/candidates/{candidateGuid}", RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "assetGuid", Description: "assetGuid public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "sessionGuid", Description: "sessionGuid public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "candidateGuid", Description: "candidateGuid public GUID", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{{Name: "from-json", Description: "Reviewed canonical request JSON; confirmation and credit values must be explicit", Required: true, Type: "string", RootJSONObjectFile: true}},
			},
			{
				Name: "patch-page", Description: "Change reviewed page selection; exclude with isSelected:false and an exclusionReason",
				ToolName: "UteamupAssetBomReviewPatchPage", HTTPMethod: "PATCH", RESTPath: "{assetGuid}/bom/review-sessions/{sessionGuid}/pages/{pageGuid}", RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "assetGuid", Description: "assetGuid public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "sessionGuid", Description: "sessionGuid public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "pageGuid", Description: "pageGuid public GUID", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{{Name: "from-json", Description: "Reviewed canonical request JSON; confirmation and credit values must be explicit", Required: true, Type: "string", RootJSONObjectFile: true}},
			},
			{
				Name: "apply", Description: "Apply only the reviewed source with observed revision and explicit acknowledged:true",
				ToolName: "UteamupAssetBomReviewApply", HTTPMethod: "POST", RESTPath: "{assetGuid}/bom/review-sessions/{sessionGuid}/commit", RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "assetGuid", Description: "assetGuid public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "sessionGuid", Description: "sessionGuid public GUID", Required: true, Type: "non-empty-uuid"},
				},
				Flags: []FlagDef{{Name: "from-json", Description: "Reviewed canonical request JSON; confirmation and credit values must be explicit", Required: true, Type: "string", RootJSONObjectFile: true}},
			},
			{
				Name: "delete", Description: "Delete staging only; applied quantities and immutable source receipts remain",
				ToolName: "UteamupAssetBomReviewDelete", HTTPMethod: "DELETE", RESTPath: "{assetGuid}/bom/review-sessions/{sessionGuid}", RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "assetGuid", Description: "assetGuid public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "sessionGuid", Description: "sessionGuid public GUID", Required: true, Type: "non-empty-uuid"},
				},
			},
		},
	})
}
