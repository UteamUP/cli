package registry

func init() {
	// Routes mirror AssetDossierController (/api/assetdossier/*). Generation is
	// asynchronous and spends AI credits: request returns a queued job and the client
	// polls it, so --wait is the terminal-friendly wrapper around that poll.
	jobGuid := ArgDef{Name: "jobGuid", Description: "Dossier job GUID", Required: true, Type: "non-empty-uuid"}

	Register(&Domain{
		Name:        "asset-dossier",
		Aliases:     []string{"dossier", "asset-dossiers"},
		Description: "Generate and follow the UPMate PDF dossier that documents one asset",
		APIPath:     "/api/assetdossier",
		Actions: []Action{
			{
				Name:        "request",
				Description: "Queue a dossier and charge its AI credits; --wait blocks until it finishes",
				ToolName:    "UteamupAssetDossierRequest",
				HTTPMethod:  "POST",
				RESTPath:    "jobs",
				Flags: []FlagDef{
					{Name: "asset-guid", BodyName: "assetGuid", Description: "Asset to document", Required: true, Type: "non-empty-uuid"},
					{Name: "request-guid", BodyName: "requestGuid", Description: "Caller-generated GUID; re-sending it returns the existing job instead of charging again", Required: true, Type: "non-empty-uuid"},
					{Name: "options", Description: "JSON object with sections[], includeWebResearch, includeDiagram and language", Type: "string", JSONFile: true},
					{Name: "wait", Description: "Block until the job reaches completed, failed or cancelled", Type: "bool", LocalOnly: true},
				},
				PollWaitFlag:         "wait",
				PollPath:             "jobs/{jobGuid}",
				PollIdentifierField:  "jobGuid",
				PollStatusField:      "status",
				PollTerminalStatuses: []string{"completed", "failed", "cancelled"},
			},
			{
				Name:        "get",
				Description: "Read one dossier job with its live step checklist and finished document",
				ToolName:    "UteamupAssetDossierGet",
				RESTPath:    "jobs/{jobGuid}",
				Args:        []ArgDef{jobGuid},
			},
			{
				Name:        "list",
				Description: "One asset's recent dossier jobs, newest first",
				ToolName:    "UteamupAssetDossierListByAsset",
				RESTPath:    "by-asset/{assetGuid}",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset GUID", Required: true, Type: "non-empty-uuid"}},
				Flags: []FlagDef{
					{Name: "take", QueryName: "take", Description: "How many recent jobs to return", Default: 10, Type: "int"},
				},
			},
			{
				Name:        "cancel",
				Description: "Cancel a job that has not started yet and refund its credits",
				ToolName:    "UteamupAssetDossierCancel",
				HTTPMethod:  "POST",
				RESTPath:    "jobs/{jobGuid}/cancel",
				Args:        []ArgDef{jobGuid},
			},
		},
	})
}
