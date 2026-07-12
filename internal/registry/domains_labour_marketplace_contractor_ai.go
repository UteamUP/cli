package registry

func init() {
	Register(&Domain{
		Name:        "labour-application-ai",
		Aliases:     []string{"contractor-application-ai"},
		Description: "Preview credits and create editable contractor application drafts without submitting",
		APIPath:     "/api/labour-marketplace/ai/contractor-application-draft",
		Actions: []Action{
			{
				Name:        "cost",
				Description: "Preview the live AI-credit cost and remaining balance",
				ToolName:    "UteamupLabourMarketplaceContractorApplicationDraftCost",
				RESTPath:    "cost",
				HTTPMethod:  "GET",
			},
			{
				Name:        "draft",
				Description: "Draft editable application content; never submits an application or proves qualifications",
				ToolName:    "UteamupLabourMarketplaceContractorApplicationDraftCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "job-guid", BodyName: "jobGuid", Description: "Open labour job GUID", Required: true, Type: "string"},
					{Name: "provider-party-guid", BodyName: "providerPartyGuid", Description: "Owned contractor party profile GUID", Required: true, Type: "string"},
					{Name: "language", Description: "Output language", Type: "string", Default: "English"},
				},
			},
		},
	})
}
