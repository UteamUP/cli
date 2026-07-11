package registry

func init() {
	Register(&Domain{
		Name:        "labour-ai-offers",
		Aliases:     []string{"offer-compare-ai"},
		Description: "Compare current labour offers with explicit AI-credit costs and no automatic selection",
		APIPath:     "/api/labour-marketplace/ai/offer-comparison",
		Actions: []Action{
			{
				Name:        "cost",
				Description: "Preview the live AI-credit cost and your remaining balance",
				ToolName:    "UteamupLabourMarketplaceOfferComparisonCost",
				RESTPath:    "cost",
				HTTPMethod:  "GET",
			},
			{
				Name:        "compare",
				Description: "Explain current offer differences; never selects, ranks, accepts, or rejects a provider",
				ToolName:    "UteamupLabourMarketplaceOfferComparison",
				Flags: []FlagDef{
					{Name: "job-guid", Description: "Buyer-owned labour job GUID", Required: true, Type: "string"},
					{Name: "revision-guids", Description: "Current offer revision GUIDs (repeatable or comma-separated; omit for latest per applicant)", Type: "stringSlice"},
					{Name: "language", Description: "Output language", Type: "string", Default: "English"},
				},
			},
		},
	})
}
