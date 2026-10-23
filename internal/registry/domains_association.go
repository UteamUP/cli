package registry

func associationMeetingFlags() []FlagDef {
	return []FlagDef{
		{Name: "title", BodyName: "title", Description: "Meeting title (at most 200 characters)", Required: true, Type: "string"},
		{Name: "kind", BodyName: "kind", Description: "Meeting kind: general (annual), board or residents", Required: true, Type: "string", AllowedValues: []string{"general", "board", "residents"}},
		{Name: "held-on", BodyName: "heldOn", Description: "When the meeting is or was held (ISO 8601 UTC)", Required: true, Type: "string"},
		{Name: "location", BodyName: "location", Description: "Where the meeting is held", Type: "string"},
		{Name: "minutes-text", BodyName: "minutesText", Description: "The meeting minutes (at most 20000 characters)", Type: "string"},
		{Name: "minutes-document-guid", BodyName: "minutesDocumentGuid", Description: "GUID of an uploaded document in this tenant with a photo or PDF of the minutes", Type: "uuid"},
	}
}

func init() {
	meetingArg := []ArgDef{{Name: "meetingGuid", Description: "GUID of the meeting", Required: true, Type: "non-empty-uuid"}}
	voteArg := []ArgDef{{Name: "voteGuid", Description: "GUID of the vote", Required: true, Type: "non-empty-uuid"}}
	noticeArg := []ArgDef{{Name: "noticeGuid", Description: "GUID of the notice", Required: true, Type: "non-empty-uuid"}}

	Register(&Domain{
		Name:        "association",
		Aliases:     []string{"housing-association"},
		Description: "UteamUP Home housing association: meetings and minutes, votes and notices (reads and voting need Tenant.View, the rest Tenant.Update)",
		APIPath:     "/api/association",
		Actions: []Action{
			{
				Name: "meetings-list", Description: "List meetings with their minutes, newest first",
				ToolName: "uteamup_association_meetings_list", HTTPMethod: "GET", RESTPath: "meetings",
				RejectExtraArgs: true,
			},
			{
				Name: "meeting-get", Description: "Show one meeting and its minutes",
				ToolName: "uteamup_association_meeting_get", HTTPMethod: "GET", RESTPath: "meetings/{meetingGuid}",
				Args: meetingArg, RejectExtraArgs: true,
			},
			{
				Name: "meeting-create", Description: "Record a meeting (Tenant.Update)",
				ToolName: "uteamup_association_meeting_create", HTTPMethod: "POST", RESTPath: "meetings",
				Flags: associationMeetingFlags(), RejectExtraArgs: true,
			},
			{
				Name: "meeting-update", Description: "Replace a meeting's details and minutes; an omitted optional flag clears that value (Tenant.Update)",
				ToolName: "uteamup_association_meeting_update", HTTPMethod: "PUT", RESTPath: "meetings/{meetingGuid}",
				Args: meetingArg, Flags: associationMeetingFlags(), RejectExtraArgs: true,
			},
			{
				Name: "meeting-delete", Description: "Remove a meeting (Tenant.Update)",
				ToolName: "uteamup_association_meeting_delete", HTTPMethod: "DELETE", RESTPath: "meetings/{meetingGuid}",
				Args: meetingArg, RejectExtraArgs: true,
			},
			{
				Name: "votes-list", Description: "List votes with Yes/No/Abstain counts, whether each is open and your own answer",
				ToolName: "uteamup_association_votes_list", HTTPMethod: "GET", RESTPath: "votes",
				RejectExtraArgs: true,
			},
			{
				Name: "vote-create", Description: "Ask the members a Yes/No/Abstain question (Tenant.Update)",
				ToolName: "uteamup_association_vote_create", HTTPMethod: "POST", RESTPath: "votes",
				Flags: []FlagDef{
					{Name: "title", BodyName: "title", Description: "The question (at most 200 characters)", Required: true, Type: "string"},
					{Name: "description", BodyName: "description", Description: "Background for the members (at most 4000 characters)", Type: "string"},
					{Name: "closes-at", BodyName: "closesAt", Description: "When the vote closes (ISO 8601 UTC, in the future); omit to keep it open until closed", Type: "string"},
				},
				RejectExtraArgs: true,
			},
			{
				Name: "ballot-cast", Description: "Cast or change your answer while the vote is open",
				ToolName: "uteamup_association_ballot_cast", HTTPMethod: "POST", RESTPath: "votes/{voteGuid}/ballot",
				Args: voteArg,
				Flags: []FlagDef{
					{Name: "choice", BodyName: "choice", Description: "Your answer: yes, no or abstain", Required: true, Type: "string", AllowedValues: []string{"yes", "no", "abstain"}},
				},
				RejectExtraArgs: true,
			},
			{
				Name: "vote-close", Description: "Close a vote now (Tenant.Update)",
				ToolName: "uteamup_association_vote_close", HTTPMethod: "POST", RESTPath: "votes/{voteGuid}/close",
				Args: voteArg, RejectExtraArgs: true,
			},
			{
				Name: "notices-list", Description: "List the board's notices, newest first",
				ToolName: "uteamup_association_notices_list", HTTPMethod: "GET", RESTPath: "notices",
				RejectExtraArgs: true,
			},
			{
				Name: "notice-post", Description: "Post a notice; every other member gets a notification and a push (Tenant.Update)",
				ToolName: "uteamup_association_notice_post", HTTPMethod: "POST", RESTPath: "notices",
				Flags: []FlagDef{
					{Name: "title", BodyName: "title", Description: "Notice title (at most 200 characters)", Required: true, Type: "string"},
					{Name: "body", BodyName: "body", Description: "Notice text (at most 4000 characters)", Required: true, Type: "string"},
				},
				RejectExtraArgs: true,
			},
			{
				Name: "notice-delete", Description: "Remove a notice (Tenant.Update)",
				ToolName: "uteamup_association_notice_delete", HTTPMethod: "DELETE", RESTPath: "notices/{noticeGuid}",
				Args: noticeArg, RejectExtraArgs: true,
			},
		},
	})
}
