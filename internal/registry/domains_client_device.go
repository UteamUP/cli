package registry

func init() {
	Register(&Domain{
		Name:        "client-device",
		Description: "Inspect GUID-enrolled devices or revoke an observed enrollment without wiping pending work; metadata grants no offline access",
		APIPath:     "/api/device/enrolled",
		Actions: []Action{
			{
				Name: "mine", Description: "List the current first-party tenant member's enrollments; API-key audit users have no own-device shortcut",
				ToolName: "UteamupClientDeviceMine", HTTPMethod: "GET", RESTPath: "mine",
				Flags: clientDevicePageFlags(),
			},
			{
				Name: "list", Description: "List active-tenant enrollments with current Sync.Manage authority",
				ToolName: "UteamupClientDeviceList", HTTPMethod: "GET", UseDomainBasePath: true,
				Flags: clientDevicePageFlags(),
			},
			{
				Name: "get", Description: "Read minimal enrollment metadata in current owner or Sync.Manage scope",
				ToolName: "UteamupClientDeviceGet", HTTPMethod: "GET", RESTPath: "{deviceGuid}",
				Args: []ArgDef{clientDeviceGuidArgument()},
			},
			{
				Name: "revoke", Description: "Revoke the latest observed enrollment revision; preserves pending work and does not promise disconnected-client lock delivery",
				ToolName: "UteamupClientDeviceRevoke", HTTPMethod: "POST", RESTPath: "{deviceGuid}/revoke",
				Args: []ArgDef{clientDeviceGuidArgument()},
				Flags: []FlagDef{
					{Name: "expected-enrollment-version", Description: "Current version from the latest metadata read; positive integer up to 9007199254740991", Type: "int", Required: true, BodyName: "expectedEnrollmentVersion"},
					{Name: "confirm", Description: "Explicitly confirm enrollment revocation; no data will be wiped", Type: "bool", Required: true, MustBeTrue: true, LocalOnly: true},
				},
			},
		},
	})
}

func clientDeviceGuidArgument() ArgDef {
	return ArgDef{Name: "device-guid", Description: "Non-empty enrolled device GUID", Type: "non-empty-uuid", Required: true, BodyName: "deviceGuid"}
}

func clientDevicePageFlags() []FlagDef {
	return []FlagDef{
		{Name: "page", Description: "Positive page number", Type: "int", Default: 1, QueryName: "page"},
		{Name: "page-size", Description: "Page size from 1 to 100", Type: "int", Default: 50, QueryName: "pageSize"},
	}
}
