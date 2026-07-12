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
				Name:        "send",
				Description: "Send private text and optional active-participant mentions without granting access or sharing history",
				ToolName:    "UteamupMarketplaceConversationMessageSend",
				RESTPath:    "{conversationGuid}/messages",
				HTTPMethod:  "POST",
				Args:        []ArgDef{conversationGuid},
				Flags: []FlagDef{
					{Name: "body", Description: "Message body", Type: "string", Required: true},
					{Name: "parent-message-guid", BodyName: "parentMessageGuid", Description: "Optional visible parent message GUID", Type: "string"},
					{Name: "mention", BodyName: "mentionedParticipantGuids", Description: "Active participant GUID to mention; repeatable, maximum 20", Type: "stringSlice"},
					{Name: "document-guid", BodyName: "documentGuids", Description: "Entitled document GUID to attach; repeatable, maximum 5", Type: "stringSlice"},
				},
			},
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
			{
				Name:        "meeting-create",
				Description: "Create a private server-owned meeting proposal; this does not approve a job, offer, or agreement",
				ToolName:    "UteamupMarketplaceConversationMeetingCreate",
				RESTPath:    "{conversationGuid}/messages/meeting-proposals",
				HTTPMethod:  "POST",
				Args:        []ArgDef{conversationGuid},
				Flags: []FlagDef{
					{Name: "subject", Description: "Meeting subject", Type: "string", Required: true},
					{Name: "start-utc", BodyName: "scheduledStartUtc", Description: "Future ISO-8601 meeting start in UTC", Type: "string", Required: true},
					{Name: "end-utc", BodyName: "scheduledEndUtc", Description: "ISO-8601 meeting end in UTC", Type: "string", Required: true},
					{Name: "time-zone", BodyName: "timeZone", Description: "IANA or supported system time-zone identifier", Type: "string", Required: true},
					{Name: "location", Description: "Optional location or meeting link", Type: "string"},
					{Name: "notes", Description: "Optional private meeting notes", Type: "string"},
				},
			},
			{
				Name:        "meeting-respond",
				Description: "Accept, decline, or cancel a meeting through participant-side authorization",
				ToolName:    "UteamupMarketplaceConversationMeetingRespond",
				RESTPath:    "{conversationGuid}/messages/{messageGuid}/meeting-response",
				HTTPMethod:  "PATCH",
				Args: []ArgDef{
					conversationGuid,
					{Name: "messageGuid", Description: "Meeting proposal message GUID", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "status", Description: "accepted, declined, or cancelled", Type: "string", Required: true},
				},
			},
			{
				Name:        "offer-share",
				Description: "Share a server-generated snapshot of the current authoritative offer revision",
				ToolName:    "UteamupMarketplaceConversationOfferCardCreate",
				RESTPath:    "{conversationGuid}/messages/offer-cards",
				HTTPMethod:  "POST",
				Args:        []ArgDef{conversationGuid},
				Flags: []FlagDef{
					{Name: "offer-revision-guid", BodyName: "offerRevisionGuid", Description: "Authoritative labour offer revision GUID", Type: "string", Required: true},
				},
			},
			{
				Name:        "contact-share",
				Description: "Share selected server-owned account contact fields after the configured disclosure stage",
				ToolName:    "UteamupMarketplaceConversationContactCardCreate",
				RESTPath:    "{conversationGuid}/messages/contact-cards",
				HTTPMethod:  "POST",
				Args:        []ArgDef{conversationGuid},
				Flags: []FlagDef{
					{Name: "email", BodyName: "includeEmail", Description: "Share the authenticated account email when available", Type: "bool"},
					{Name: "phone", BodyName: "includePhone", Description: "Share the authenticated work or mobile phone when available", Type: "bool"},
					{Name: "website", BodyName: "includeWebsite", Description: "Share the authenticated account website when available", Type: "bool"},
				},
			},
		},
	})
}
