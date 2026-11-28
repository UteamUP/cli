package registry

func init() {
	Register(&Domain{
		Name:        "iceland-public-data",
		Description: "Icelandic public source catalog and last-good snapshots (configured Global Admin)",
		APIPath:     "/api/globaladmin/iceland-data",
		Actions: []Action{{
			Name: "list", Description: "Read source metadata, reviewed reuse terms, configuration and snapshot freshness; never downloads the register",
			ToolName: "UteamupIcelandPublicDataList", HTTPMethod: "GET", UseDomainBasePath: true, RejectExtraArgs: true,
		}, {
			Name: "update", Description: "Update the supported source using its current revision; stale revisions and unknown configuration fields are rejected by the server",
			ToolName: "UteamupIcelandPublicDataUpdate", HTTPMethod: "PUT", RESTPath: "{sourceKey}", RejectExtraArgs: true,
			Args: []ArgDef{{Name: "sourceKey", Description: "Supported source key", Type: "string", Required: true, AllowedValues: []string{"is-hms-stadfangaskra"}}},
			Flags: []FlagDef{{Name: "from-json", Description: "Root JSON containing required expectedRevision (nonempty GUID), enabled (boolean) and refreshHours (integer 24..720); no other fields, URLs, credentials or code",
				Type: "string", Required: true, RootJSONObjectFile: true}},
		}, {
			Name: "refresh", Description: "Refresh enabled HMS data; a failed refresh preserves the last-good snapshot",
			ToolName: "UteamupIcelandPublicDataRefresh", HTTPMethod: "POST", RESTPath: "{sourceKey}/refresh", RejectExtraArgs: true,
			Args: []ArgDef{{Name: "sourceKey", Description: "Supported source key", Type: "string", Required: true, AllowedValues: []string{"is-hms-stadfangaskra"}}},
		}},
	})
}
