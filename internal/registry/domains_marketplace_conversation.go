package registry

func init() {
	conversationGuid := ArgDef{
		Name:        "conversationGuid",
		Description: "Participant-scoped marketplace conversation GUID",
		Required:    true,
		Type:        "string",
	}
	Register(&Domain{
		Name:        "marketplace-conversation",
		Aliases:     []string{"conversation", "marketplace-chat"},
		Description: "Search and manage only the authenticated participant's private marketplace conversation",
		APIPath:     "/api/marketplace/conversations",
		Actions: []Action{
			{
				Name:        "search",
				Description: "Search only conversation messages the authenticated participant may read",
				ToolName:    "UteamupMarketplaceConversationMessagesSearch",
				RESTPath:    "{conversationGuid}/messages/search",
				HTTPMethod:  "GET",
				Args:        []ArgDef{conversationGuid},
				Flags: []FlagDef{
					{Name: "query", Description: "Search text, 2 to 200 characters", Type: "string", Required: true},
				},
			},
			{
				Name:        "mute",
				Description: "Mute or unmute notifications only for the authenticated participant",
				ToolName:    "UteamupMarketplaceConversationPreferencesUpdate",
				RESTPath:    "{conversationGuid}/preferences",
				HTTPMethod:  "PATCH",
				Args:        []ArgDef{conversationGuid},
				Flags: []FlagDef{
					{Name: "muted", BodyName: "isMuted", Description: "True to mute; false to unmute", Type: "bool", Required: true},
				},
			},
			{
				Name:        "pin",
				Description: "Pin or unpin one visible message; requires an active conversation owner",
				ToolName:    "UteamupMarketplaceConversationMessagePinUpdate",
				RESTPath:    "{conversationGuid}/messages/{messageGuid}/pin",
				HTTPMethod:  "PATCH",
				Args: []ArgDef{
					conversationGuid,
					{Name: "messageGuid", Description: "Visible marketplace conversation message GUID", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "pinned", BodyName: "isPinned", Description: "True to pin; false to unpin", Type: "bool", Required: true},
				},
			},
		},
	})
}
