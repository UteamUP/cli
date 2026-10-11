package registry

// Radio administration and retained history use the same active-tenant REST boundary as mobile.
func init() {
	Register(&Domain{
		Name: "radio-admin", Description: "Configured Global Admin controls for the selected tenant's Cloud Radio pilot", APIPath: "/api/admin/radio",
		Actions: []Action{
			{Name: "validate", Description: "Explicitly probe the selected tenant's Radio services and recording storage; never starts audio capture", ToolName: "UteamupRadioAdministrationValidate", HTTPMethod: "POST", RESTPath: "validate"},
			{Name: "status", Description: "Read the selected tenant's Radio environment without participant access", ToolName: "UteamupRadioAdministrationGet", HTTPMethod: "GET", RESTPath: "environment"},
			{Name: "grant", Description: "Grant a complimentary 25-listener pilot with 30-day retention and 50 GiB storage; queues provisioning without an invoice", ToolName: "UteamupRadioComplimentaryGrant", HTTPMethod: "POST", RESTPath: "complimentary",
				Flags: []FlagDef{{Name: "file", Description: "JSON containing requestGuid, expiresAt (null until manually disabled), and reason", Type: "string", Required: true, RootJSONObjectFile: true}}},
			{Name: "revoke", Description: "Revoke a named pilot grant and drain compute while preserving retained recordings", ToolName: "UteamupRadioComplimentaryRevoke", HTTPMethod: "POST", RESTPath: "complimentary/revoke",
				Flags: []FlagDef{{Name: "file", Description: "JSON containing grantGuid and reason", Type: "string", Required: true, RootJSONObjectFile: true}}},
		},
	})
	tenant := FlagDef{Name: "tenant", Short: "t", BodyName: "tenantGuid", Description: "Public GUID of the tenant", Required: true, Type: "non-empty-uuid"}
	member := FlagDef{Name: "user-guid", BodyName: "userGuid", Description: "Public GUID of the tenant member", Required: true, Type: "non-empty-uuid"}
	Register(&Domain{
		Name: "radio-license", Description: "Tenant.Update: Radio seats per member (seats = the tenant's Radio listener capacity)", APIPath: "/api/tenant",
		Actions: []Action{
			{Name: "list", Description: "List members holding a Radio seat; backfilled seats came from the rollout grant", ToolName: "UteamupRadioLicenseList", HTTPMethod: "GET", RESTPath: "{tenantGuid}/radio-license-users",
				Flags: []FlagDef{tenant}},
			{Name: "assign", Description: "Give a member a Radio seat; fails with RADIO_LICENSE_POOL_EXHAUSTED when all seats are used", ToolName: "UteamupRadioLicenseAssign", HTTPMethod: "POST", RESTPath: "{tenantGuid}/radio-licenses/assign",
				Flags: []FlagDef{tenant, member}},
			{Name: "remove", Description: "Remove a member's Radio seat; their live Radio stops within seconds, recordings are kept", ToolName: "UteamupRadioLicenseRemove", HTTPMethod: "POST", RESTPath: "{tenantGuid}/radio-licenses/remove",
				Flags: []FlagDef{tenant, member}},
		},
	})
	recording := ArgDef{Name: "recordingGuid", Description: "Public recording GUID", Type: "non-empty-uuid", Required: true}
	Register(&Domain{
		Name: "radio", Description: "Manage tenant radio policy and read authorized recording history", APIPath: "/api/radio",
		Actions: []Action{
			{Name: "team-eligibility", Description: "Team.View and Radio.Manage: read a saved Unit team's Radio eligibility without assigning seats", ToolName: "UteamupRadioTeamEligibilityGet", HTTPMethod: "GET", RESTPath: "teams/{teamGuid}/eligibility",
				Args: []ArgDef{{Name: "teamGuid", Description: "Public Unit team GUID", Type: "non-empty-uuid", Required: true}}},
			{Name: "backup-sharepoint", Description: "Radio.Manage: read the selected tenant's SharePoint backup availability", ToolName: "UteamupRadioBackupSharePointAvailability", HTTPMethod: "GET", RESTPath: "backup-destinations/sharepoint"},
			{Name: "validate", Description: "Radio.Manage: explicitly probe services and recording storage; reports unverified device checks separately", ToolName: "UteamupRadioValidate", HTTPMethod: "POST", RESTPath: "validate"},
			{Name: "stop-pilot", Description: "Stop the selected tenant's named complimentary pilot and safely remove audio servers; retained recordings are kept", ToolName: "UteamupRadioPilotStop", HTTPMethod: "POST", RESTPath: "complimentary/revoke",
				Flags: []FlagDef{{Name: "file", Description: "JSON containing grantGuid and reason; requires Radio.Manage and tenant membership", Type: "string", Required: true, RootJSONObjectFile: true}}},
			{Name: "transcript", Description: "Read an authorized speaking-turn transcript, including unarchived audio", ToolName: "UteamupRadioTransmissionTranscriptGet", HTTPMethod: "GET", RESTPath: "channels/{channelGuid}/transmissions/{transmissionGuid}/transcript",
				Args: []ArgDef{{Name: "channelGuid", Description: "Public channel GUID", Type: "non-empty-uuid", Required: true}, {Name: "transmissionGuid", Description: "Public transmission GUID", Type: "non-empty-uuid", Required: true}}},
			{Name: "transmissions", Description: "Read authorized speaking-turn history, including unrecorded turns", ToolName: "UteamupRadioTransmissionsList", HTTPMethod: "GET", RESTPath: "channels/{channelGuid}/transmissions",
				Args:  []ArgDef{{Name: "channelGuid", Description: "Public channel GUID", Type: "non-empty-uuid", Required: true}},
				Flags: []FlagDef{{Name: "skip", Description: "Skip count", Type: "int", Default: 0}, {Name: "take", Description: "Page size (1–100)", Type: "int", Default: 50}}},
			{Name: "status", Description: "Read subscription, policy and provisioning health", ToolName: "UteamupRadioEnvironmentGet", HTTPMethod: "GET", RESTPath: "environment"},
			{Name: "policy", Description: "Apply a complete reviewed policy from a JSON file; never starts microphone capture", ToolName: "UteamupRadioPolicyUpdate", HTTPMethod: "PUT", RESTPath: "policy",
				Flags: []FlagDef{{Name: "file", Description: "JSON file containing the complete RadioPolicyModel", Type: "string", Required: true, RootJSONObjectFile: true}}},
			{Name: "channels", Description: "List currently authorized shared history channels, including after cancellation", ToolName: "UteamupRadioHistoryChannelsList", HTTPMethod: "GET", RESTPath: "history-channels"},
			{Name: "recordings", Description: "List own private recordings (--scope Mine, the default), own plus every authorized shared channel's history with --scope All, or one shared channel with --channel-guid", ToolName: "UteamupRadioRecordingsList", HTTPMethod: "GET", RESTPath: "recordings",
				Flags: []FlagDef{
					{Name: "channel-guid", BodyName: "channelGuid", QueryName: "channelGuid", Description: "Public shared channel GUID; omit for own private recordings", Type: "non-empty-uuid"},
					{Name: "scope", QueryName: "scope", Description: "Mine (own private recordings, the default) or All (own plus authorized shared history)", Type: "string", AllowedValues: []string{"Mine", "All"}},
					{Name: "from", QueryName: "from", Description: "Only recordings started at or after this UTC timestamp (RFC3339)", Type: "string"},
					{Name: "to", QueryName: "to", Description: "Only recordings started before this UTC timestamp (RFC3339, exclusive)", Type: "string"},
					{Name: "team-guid", QueryName: "teamGuid", Description: "Public Unit team GUID; only shared recordings in that team's readable channels", Type: "non-empty-uuid"},
					{Name: "has-transcript", QueryName: "hasTranscript", Description: "true for recordings with a final transcript, false for those without; omit for both", Type: "bool"},
					{Name: "transcript-contains", QueryName: "transcriptContains", Description: "Only recordings whose transcript contains this text (2–200 characters)", Type: "string"},
					{Name: "skip", Description: "Number of recordings to skip (0–10000)", Type: "int", Default: 0},
					{Name: "take", Description: "Page size (1–100)", Type: "int", Default: 50},
				}},
			{Name: "always-transcribe", Description: "Radio.Listen: turn 'Always transcribe while speaking' on or off for yourself; enabling requires AI.Use and spends AI credits per minute of speech", ToolName: "UteamupRadioAlwaysTranscribeSet", HTTPMethod: "PUT", RESTPath: "me/always-transcribe",
				Flags: []FlagDef{{Name: "enabled", BodyName: "enabled", Description: "true to transcribe every speaking turn, false to stop", Type: "bool", Required: true}}},
			{Name: "transcribe", Description: "AI.Use: request or retry transcription of a recording that still has audio; retrying deletes the previous partial transcript", ToolName: "UteamupRadioTranscriptionRequest", HTTPMethod: "POST", RESTPath: "recordings/{recordingGuid}/transcription", Args: []ArgDef{recording}},
			{Name: "get", Description: "Read authorized recording metadata and transcript", ToolName: "UteamupRadioRecordingGet", HTTPMethod: "GET", RESTPath: "recordings/{recordingGuid}", Args: []ArgDef{recording}},
			{Name: "playback", Description: "Create a short-lived private playback URL", ToolName: "UteamupRadioPlaybackGet", HTTPMethod: "GET", RESTPath: "recordings/{recordingGuid}/playback", Args: []ArgDef{recording}, DisableResponseExport: true},
		},
	})
}
