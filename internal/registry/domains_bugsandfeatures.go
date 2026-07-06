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
					{Name: "status", Description: "Filter by status (New, Validated, Fixed, Confirmed, Rejected, WaitList)", Type: "string"},
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
					// Routed via HeaderName because BugsAndFeaturesController.Create reads
					// it from `[FromHeader(Name = "Idempotency-Key")]`. Sending it in the
					// JSON body returns `400 "Missing or invalid Idempotency-Key header."`.
					{Name: "idempotency-key", HeaderName: "Idempotency-Key", Description: "Client-generated idempotency key (GUID), sent as the Idempotency-Key HTTP header", Required: true, Type: "string"},
					{Name: "route-path", Description: "Route path the user was on when reporting", Type: "string"},
				},
			},
			{
				Name:        "update-status",
				Description: "Transition a submission's status (global-admin only)",
				ToolName:    "UteamupBugsAndFeaturesUpdateStatus",
				Args: []ArgDef{
					{Name: "externalGuid", Description: "ExternalGuid", Required: true, Type: "string"},
					{Name: "toStatus", Description: "Target status (Validated, Fixed, Confirmed, Rejected, New, WaitList)", Required: true, Type: "string"},
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
				Name:        "increment-hit",
				Description: "Manually record a hit (occurrence) for an existing submission (global-admin only). Atomically increments OccurrenceCount and updates LastSeenAtUtc; appends a [manual-hit] audit row carrying the optional route/environment/evidence/note metadata. Returns before/after counters so uteamup-validate reports can prove the hit was recorded without spoofing frontend auto-capture.",
				ToolName:    "UteamupBugsAndFeaturesIncrementHit",
				HTTPMethod:  "POST",
				RESTPath:    "{externalGuid}/increment-hit",
				Args:        []ArgDef{{Name: "externalGuid", Description: "ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "route-path", BodyName: "routePath", Description: "Route the validator was on when re-observing the issue. Max 512 chars.", Type: "string"},
					{Name: "environment", BodyName: "environment", Description: "Environment the hit was observed on (localhost / dev / staging / prod). Max 64 chars.", Type: "string"},
					{Name: "evidence", BodyName: "evidence", Description: "Free-text validation evidence (URL, trace id, brief description). Max 2 KB.", Type: "string"},
					{Name: "note", BodyName: "note", Description: "Optional human-readable note recorded on the audit row. Max 1 KB.", Type: "string"},
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
				Description: "List the comment thread on a bug (global-admin only). Top-level comments newest-first (latest at the top); replies under each comment stay oldest-first and are eagerly included.",
				ToolName:    "UteamupBugsAndFeaturesCommentsList",
				HTTPMethod:  "GET",
				RESTPath:    "{bugExternalGuid}/comments",
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
				HTTPMethod:  "POST",
				RESTPath:    "{bugExternalGuid}/comments",
				Args: []ArgDef{
					{Name: "bugExternalGuid", Description: "Bug ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					// CLI flag names stay short; BodyName aligns each one with the
					// backend BugCommentCreateRequest DTO (see UteamUP_API/Models/
					// BugCommentModels.cs). Without these mappings the body would
					// carry `text` / `parent` / `mention` and the backend would 400.
					{Name: "text", Short: "t", BodyName: "bodyHtml", Description: "Comment body (HTML accepted; plain text is wrapped). Max 8000 chars after sanitization.", Required: true, Type: "string"},
					{Name: "parent", BodyName: "parentCommentExternalGuid", Description: "ExternalGuid of the parent comment to reply to. Replies of replies are rejected (single-level threading).", Type: "string"},
					{Name: "mention", BodyName: "mentionedGlobalAdminGuids", Description: "Global-admin user GUID to @-mention. Repeatable; max 10 per comment. Server rejects non-admin GUIDs with 400.", Type: "stringSlice"},
					{Name: "share-with-reporter", BodyName: "isVisibleToReporter", Description: "Share this comment with the bug's original reporter (they see it in their notification center and can reply). Defaults to internal.", Type: "bool"},
				},
			},
			{
				Name:        "ping-reporter",
				Description: "Manually re-notify the bug's reporter that an admin is waiting for their reply (global-admin only). Sends the standard reporter notification + email. 409 when the report has no reporter account.",
				ToolName:    "UteamupBugsAndFeaturesPingReporter",
				HTTPMethod:  "POST",
				RESTPath:    "{bugExternalGuid}/ping-reporter",
				Args: []ArgDef{
					{Name: "bugExternalGuid", Description: "Bug ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "conversation",
				Description: "Read the reporter-facing conversation on a bug: the original report (title, description, status), the reporter-facing status timeline, and the comments shared with the reporter. Authorized to the report's submitter or any global admin.",
				ToolName:    "UteamupBugsAndFeaturesConversation",
				HTTPMethod:  "GET",
				RESTPath:    "{bugExternalGuid}/conversation",
				Args: []ArgDef{
					{Name: "bugExternalGuid", Description: "Bug ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "mine",
				Description: "List the bugs/features submitted by the current user (the caller's own reports), newest activity first. Slim reporter-safe projection — powers the 'My reports' page.",
				ToolName:    "UteamupBugsAndFeaturesMine",
				HTTPMethod:  "GET",
				RESTPath:    "mine",
				Args:        []ArgDef{},
			},
			{
				Name:        "attachments-list",
				Description: "List attachments on a bug (global-admin only) — images, documents, and videos. Output is oldest-first. Each row exposes Kind (Image|Document|Video), ContentType, OriginalFileName, and Extension so the caller can pick the right preview affordance.",
				ToolName:    "UteamupBugsAndFeaturesAttachmentsList",
				HTTPMethod:  "GET",
				RESTPath:    "{bugExternalGuid}/attachments",
				Args: []ArgDef{
					{Name: "bugExternalGuid", Description: "Bug ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "attachments-upload",
				Description: "Upload an image, document, or video attachment to a bug. Per-family caps: image 2 MB, document 25 MB, video 100 MB. Server stores in the global-admin-owned namespace inside the shared bugattachments blob container — never to a tenant's SharePoint or dedicated tenant storage. Bytes are byte-verbatim for documents/videos; images are re-encoded via SkiaSharp (drops EXIF, bounds longest edge to 1600 px). Multipart over REST.",
				ToolName:    "UteamupBugsAndFeaturesAttachmentsUpload",
				HTTPMethod:  "POST",
				RESTPath:    "{bugExternalGuid}/attachments",
				Args: []ArgDef{
					{Name: "bugExternalGuid", Description: "Bug ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "file", Short: "f", Description: "Local path to the file to upload. Supported: PNG, JPEG, GIF, WebP, SVG; PDF, DOC/DOCX, XLS/XLSX, PPT/PPTX, TXT, MD, CSV, RTF, JSON, XML; MP4, MOV, WEBM, MKV. Required.", Required: true, Type: "string", UploadFile: true},
				},
			},
			{
				Name:        "attachments-download",
				Description: "Download a single attachment via the SAS URL endpoint. Works for images, documents, and videos alike. Writes to ./<attachmentGuid>.<ext> by default; --out overrides. The server reuses the original extension for the output filename when known so the downloaded file opens in the right app.",
				ToolName:    "UteamupBugsAndFeaturesAttachmentsDownload",
				Args: []ArgDef{
					{Name: "bugExternalGuid", Description: "Bug ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
					{Name: "attachmentExternalGuid", Description: "Attachment ExternalGuid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "out", Short: "o", Description: "Output file path. If omitted, writes to ./<attachmentGuid>.<ext> in the current directory.", Type: "string"},
				},
			},
			{
				Name:        "attachments-delete",
				Description: "Hard-delete a single attachment row + best-effort delete its blob (global-admin only). Audit-logged server-side with the original filename and content-type. Works for images, documents, and videos.",
				ToolName:    "UteamupBugsAndFeaturesAttachmentsDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "{bugExternalGuid}/attachments/{attachmentExternalGuid}",
				Args: []ArgDef{
					{Name: "bugExternalGuid", Description: "Bug ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
					{Name: "attachmentExternalGuid", Description: "Attachment ExternalGuid", Required: true, Type: "string"},
				},
			},
		},
	})
}
