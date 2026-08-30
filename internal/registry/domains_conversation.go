package registry

// Platform messaging: direct messages, groups, channels and boards.
//
// Every action acts as the authenticated caller. None takes a user argument, and none should:
// identity comes from the token, so a user flag would let a caller read someone else's inbox
// or post as them.
//
// APIPath is mandatory. Omitting it makes the registry fabricate "/api/conversation" from the
// domain name, which happens to be right here but silently 404s the moment a name and a route
// diverge - so it is stated rather than inferred.
func init() {
	conversationGUID := ArgDef{
		Name:        "conversationGuid",
		Description: "Public GUID of a conversation the caller participates in",
		Required:    true,
		Type:        "uuid",
	}

	Register(&Domain{
		Name:        "conversation",
		Aliases:     []string{"message", "messages", "dm"},
		Description: "Read and send messages in conversations the authenticated user takes part in",
		APIPath:     "/api/conversation",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the caller's conversations, across tenants, with unread counts",
				ToolName:    "UteamupConversationList",
				HTTPMethod:  "GET",
			},
			{
				Name:        "get",
				Description: "Read one conversation and the messages the caller is entitled to see",
				ToolName:    "UteamupConversationGet",
				HTTPMethod:  "GET",
				RESTPath:    "{conversationGuid}",
				Args:        []ArgDef{conversationGUID},
			},
			{
				Name:        "send",
				Description: "Send a message as the caller. The sender cannot be specified",
				ToolName:    "UteamupConversationMessageSend",
				HTTPMethod:  "POST",
				RESTPath:    "{conversationGuid}/messages",
				Args:        []ArgDef{conversationGUID},
				Flags: []FlagDef{
					{Name: "body", Description: "Message body", Type: "string", Required: true},
					{Name: "parent-message-guid", BodyName: "parentMessageGuid", Description: "Optional parent message GUID to reply to", Type: "string"},
					{Name: "client-message-id", BodyName: "clientMessageId", Description: "Idempotency key; resending with the same value returns the original message instead of a duplicate", Type: "string"},
				},
			},
			{
				Name:        "react",
				Description: "Toggle the caller's emoji reaction on a message (same emoji again removes it)",
				ToolName:    "UteamupConversationMessageReactionToggle",
				HTTPMethod:  "POST",
				RESTPath:    "{conversationGuid}/messages/{messageGuid}/reactions",
				Args: []ArgDef{
					conversationGUID,
					{Name: "messageGuid", Description: "Public GUID of the message to react to", Required: true},
				},
				Flags: []FlagDef{
					{Name: "emoji", Description: "One of the six supported emojis: 👍 👎 ❤️ 😂 😮 ☹️", Type: "string", Required: true},
				},
			},
			{
				Name:        "search",
				Description: "Search messages the caller may read within one conversation",
				ToolName:    "UteamupConversationMessagesSearch",
				HTTPMethod:  "GET",
				RESTPath:    "{conversationGuid}/messages/search",
				Args:        []ArgDef{conversationGUID},
				Flags: []FlagDef{
					{Name: "query", Description: "Text to search for", Type: "string", Required: true},
				},
			},
			{
				Name:        "since",
				Description: "Fetch messages after a sequence number - the catch-up read used to reconcile after a disconnect",
				ToolName:    "UteamupConversationGet",
				HTTPMethod:  "GET",
				RESTPath:    "{conversationGuid}/messages",
				Args:        []ArgDef{conversationGUID},
				Flags: []FlagDef{
					{Name: "since", Description: "Highest sequence number already held; 0 returns the full readable history", Type: "int"},
				},
			},
			{
				Name:        "archive",
				Description: "Archive or unarchive a conversation for the caller only",
				ToolName:    "UteamupConversationGet",
				HTTPMethod:  "PATCH",
				RESTPath:    "{conversationGuid}/inbox-state",
				Args:        []ArgDef{conversationGUID},
				Flags: []FlagDef{
					{Name: "archived", BodyName: "isArchived", Description: "true to archive, false to restore", Type: "bool", Required: true},
				},
			},
		},
	})
}
