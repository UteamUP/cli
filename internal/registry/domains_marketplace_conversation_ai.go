package registry

func init() {
	conversationGuid := []ArgDef{{
		Name:        "conversationGuid",
		Description: "Participant-scoped marketplace conversation GUID",
		Required:    true,
		Type:        "string",
	}}
	Register(&Domain{
		Name:        "marketplace-conversation-ai",
		Aliases:     []string{"conversation-summary-ai"},
		Description: "Summarize only authorized marketplace conversation history with explicit AI-credit costs",
		APIPath:     "/api/marketplace/conversations",
		Actions: []Action{
			{
				Name:        "cost",
				Description: "Preview the summary cost after participant authorization",
				ToolName:    "UteamupMarketplaceConversationAiSummaryCost",
				RESTPath:    "{conversationGuid}/ai-summary/cost",
				HTTPMethod:  "GET",
				Args:        conversationGuid,
			},
			{
				Name:        "summarize",
				Description: "Return an advisory visible-message summary without changing the conversation or commercial state",
				ToolName:    "UteamupMarketplaceConversationAiSummary",
				RESTPath:    "{conversationGuid}/ai-summary",
				HTTPMethod:  "POST",
				Args:        conversationGuid,
				Flags: []FlagDef{
					{Name: "language", Description: "Output language", Type: "string", Default: "English"},
				},
			},
		},
	})
}
