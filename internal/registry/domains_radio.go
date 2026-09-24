package registry

// Radio administration and retained history use the same active-tenant REST boundary as mobile.
func init() {
	recording := ArgDef{Name: "recordingGuid", Description: "Public recording GUID", Type: "non-empty-uuid", Required: true}
	Register(&Domain{
		Name: "radio", Description: "Manage tenant radio policy and read authorized recording history", APIPath: "/api/radio",
		Actions: []Action{
			{Name: "status", Description: "Read subscription, policy and provisioning health", ToolName: "UteamupRadioEnvironmentGet", HTTPMethod: "GET", RESTPath: "environment"},
			{Name: "policy", Description: "Apply a complete reviewed policy from a JSON file; never starts microphone capture", ToolName: "UteamupRadioPolicyUpdate", HTTPMethod: "PUT", RESTPath: "policy",
				Flags: []FlagDef{{Name: "file", Description: "JSON file containing the complete RadioPolicyModel", Type: "string", Required: true, RootJSONObjectFile: true}}},
			{Name: "channels", Description: "List currently authorized shared history channels, including after cancellation", ToolName: "UteamupRadioHistoryChannelsList", HTTPMethod: "GET", RESTPath: "history-channels"},
			{Name: "recordings", Description: "List own private recordings, or shared history with --channel-guid", ToolName: "UteamupRadioRecordingsList", HTTPMethod: "GET", RESTPath: "recordings",
				Flags: []FlagDef{
					{Name: "channel-guid", BodyName: "channelGuid", QueryName: "channelGuid", Description: "Public shared channel GUID; omit for own private recordings", Type: "non-empty-uuid"},
					{Name: "skip", Description: "Number of recordings to skip (0–10000)", Type: "int", Default: 0},
					{Name: "take", Description: "Page size (1–100)", Type: "int", Default: 50},
				}},
			{Name: "get", Description: "Read authorized recording metadata and transcript", ToolName: "UteamupRadioRecordingGet", HTTPMethod: "GET", RESTPath: "recordings/{recordingGuid}", Args: []ArgDef{recording}},
			{Name: "playback", Description: "Create a short-lived private playback URL", ToolName: "UteamupRadioPlaybackGet", HTTPMethod: "GET", RESTPath: "recordings/{recordingGuid}/playback", Args: []ArgDef{recording}, DisableResponseExport: true},
		},
	})
}
