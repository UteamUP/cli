package registry

func init() {
	Register(&Domain{
		Name:        "bugsandfeatures",
		Aliases:     []string{"bugs", "features", "baf"},
		Description: "Submit or triage user-reported bugs and feature requests (global-admin for list/get/update)",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List submissions with filters (global-admin only)",
				ToolName:    "UteamupBugsAndFeaturesList",
				Flags: []FlagDef{
					{Name: "type", Description: "Filter by type (Bug or Feature)", Type: "string"},
					{Name: "status", Description: "Filter by status (New, Validated, Fixed, Confirmed, Rejected)", Type: "string"},
					{Name: "severity", Description: "Filter by severity (Low, Medium, High, Critical)", Type: "string"},
					{Name: "source", Description: "Filter by source (Manual, FrontendAuto, PerformanceAuto)", Type: "string"},
					{Name: "tenant-guid", Description: "Filter by tenant ExternalGuid", Type: "string"},
					{Name: "submitter-user-id", Description: "Filter by submitter user id", Type: "string"},
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page (max 200)", Default: 50, Type: "int"},
					{Name: "hide-rejected-and-confirmed", Description: "Hide Rejected and Confirmed rows by default", Default: true, Type: "bool"},
					{Name: "search", Short: "q", Description: "Free-text search across title, description, tenant, submitter, location, interaction, user activity, component chain, additional notes, and audit trail (case-insensitive substring; max 200 chars)", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get a single submission by ExternalGuid (global-admin only)",
				ToolName:    "UteamupBugsAndFeaturesGet",
				Args:        []ArgDef{{Name: "externalGuid", Description: "ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "create",
				Description: "Submit a new bug or feature request",
				ToolName:    "UteamupBugsAndFeaturesCreate",
				Flags: []FlagDef{
					{Name: "type", Description: "Bug or Feature", Default: "Bug", Type: "string"},
					{Name: "severity", Description: "Low | Medium | High | Critical", Default: "Medium", Type: "string"},
					{Name: "title", Description: "Short title (max 200)", Required: true, Type: "string"},
					{Name: "description", Description: "Description (max 4000)", Required: true, Type: "string"},
					{Name: "idempotency-key", Description: "Client-generated idempotency key (GUID)", Required: true, Type: "string"},
					{Name: "route-path", Description: "Route path the user was on when reporting", Type: "string"},
				},
			},
			{
				Name:        "update-status",
				Description: "Transition a submission's status (global-admin only)",
				ToolName:    "UteamupBugsAndFeaturesUpdateStatus",
				Args: []ArgDef{
					{Name: "externalGuid", Description: "ExternalGuid", Required: true, Type: "string"},
					{Name: "toStatus", Description: "Target status (Validated, Fixed, Confirmed, Rejected, New)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "note", Description: "Required on Rejected and reopen transitions", Type: "string"},
					{Name: "resolution-reference", Description: "Required on Fixed transitions (URL or commit)", Type: "string"},
				},
			},
			{
				Name:        "update-notes",
				Description: "Set or clear freeform admin notes on a submission (global-admin only). Empty string clears.",
				ToolName:    "UteamupBugsAndFeaturesUpdateNotes",
				Args:        []ArgDef{{Name: "externalGuid", Description: "ExternalGuid", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "additional-notes", Description: "Notes body (pass empty string to clear). Max 8 KB.", Type: "string"},
				},
			},
			{
				Name:        "update-type",
				Description: "Convert a submission between Bug and Feature (global-admin only). Records the change in the audit history.",
				ToolName:    "UteamupBugsAndFeaturesUpdateType",
				Args: []ArgDef{
					{Name: "externalGuid", Description: "ExternalGuid", Required: true, Type: "string"},
					{Name: "type", Description: "Target type (Bug or Feature)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "note", Description: "Optional reason recorded on the audit-trail entry. Max 1 KB.", Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Permanently delete a submission (global-admin only; for junk entries — use Reject/Archive for normal lifecycle)",
				ToolName:    "UteamupBugsAndFeaturesDelete",
				Args:        []ArgDef{{Name: "externalGuid", Description: "ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "comments-list",
				Description: "List the comment thread on a bug (global-admin only). Top-level comments oldest-first; replies eagerly included.",
				ToolName:    "UteamupBugsAndFeaturesCommentsList",
				Args: []ArgDef{
					{Name: "bugExternalGuid", Description: "Bug ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number (top-level comments)", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Top-level comments per page (max 100)", Default: 50, Type: "int"},
				},
			},
			{
				Name:        "comments-add",
				Description: "Post a new comment (or a reply via --parent) on a bug. Optional --mention flags @-mention global admins (repeatable; max 10).",
				ToolName:    "UteamupBugsAndFeaturesCommentsAdd",
				Args: []ArgDef{
					{Name: "bugExternalGuid", Description: "Bug ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "text", Short: "t", Description: "Comment body (HTML accepted; plain text is wrapped). Max 8000 chars after sanitization.", Required: true, Type: "string"},
					{Name: "parent", Description: "ExternalGuid of the parent comment to reply to. Replies of replies are rejected (single-level threading).", Type: "string"},
					{Name: "mention", Description: "Global-admin user GUID to @-mention. Repeatable; max 10 per comment. Server rejects non-admin GUIDs with 400.", Type: "stringSlice"},
				},
			},
		},
	})
}
