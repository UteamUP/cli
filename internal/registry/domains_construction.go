package registry

// Construction module domains. GUID-first throughout: get/sub-route actions take
// the record's public GUID positional arg (`externalGuid`) and every reference
// flag carries a GUID — no integer ids anywhere on this surface (per
// Guidelines/ApiGuidelines.md §GUIDs In, Integer IDs Out).
//
// Every domain pins APIPath to its REAL controller route so the REST adapter
// can never drift from the backend (no APIPath/RESTPath = silent 404):
//
//	construction-issue → /api/constructionissue
//	rfi                → /api/rfi
//	submittal          → /api/submittal
//	dailylog           → /api/dailylog
//	construction-sheet → /api/constructionsheet
func init() {
	Register(&Domain{
		Name:        "construction-issue",
		Aliases:     []string{"construction-issues"},
		Description: "Manage construction issues / punch items",
		APIPath:     "/api/constructionissue",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List construction issues / punch items",
				ToolName:    "UteamupConstructionIssueList",
				Flags: append([]FlagDef{
					{Name: "project-guid", Description: "Restrict to one project GUID", Type: "string"},
					{Name: "status", Description: "Filter by status (Draft, Open, ReadyForReview, Closed, Void)", Type: "string"},
					{Name: "mode", Description: "Filter by mode (Issue, PunchItem)", Type: "string"},
					{Name: "priority", Description: "Filter by priority (Low, Medium, High, Critical)", Type: "string"},
					{Name: "search", Description: "Free-text search over number/title/description", Type: "string"},
				}, paginationFlags()...),
			},
			{Name: "get", Description: "Get a construction issue by GUID", ToolName: "UteamupConstructionIssueGet", Args: externalGUIDArg()},
			{
				Name:        "create",
				Description: "Create a construction issue / punch item",
				ToolName:    "UteamupConstructionIssueCreate",
				Flags: []FlagDef{
					{Name: "project-guid", Description: "Owning project GUID", Type: "string", Required: true},
					{Name: "title", Description: "Issue title", Type: "string", Required: true},
					{Name: "description", Description: "Issue description", Type: "string"},
					{Name: "mode", Description: "Issue or PunchItem (default Issue)", Type: "string"},
					{Name: "type", Description: "Category (Quality, Safety, Design, Coordination, Warranty, Commissioning, Other)", Type: "string"},
					{Name: "priority", Description: "Priority (Low, Medium, High, Critical; default Medium)", Type: "string"},
					{Name: "assignee-user-guid", Description: "Assignee user GUID", Type: "string"},
					{Name: "responsible-contact-guid", Description: "Responsible company contact GUID", Type: "string"},
					{Name: "due-date", Description: "Due date (ISO 8601)", Type: "string"},
					{Name: "location-guid", Description: "Location node GUID", Type: "string"},
					{Name: "document-guids", Description: "Existing document GUIDs to attach", Type: "stringSlice"},
				},
			},
			// Validated status transition. The GUID rides the URL; the target
			// status and optional comment ride the body (ConstructionIssueStatusModel).
			{
				Name:        "set-status",
				Description: "Transition a construction issue's status by GUID (Draft, Open, ReadyForReview, Closed, Void)",
				ToolName:    "UteamupConstructionIssueUpdateStatus",
				HTTPMethod:  "POST",
				RESTPath:    "{externalGuid}/status",
				Args:        externalGUIDArg(),
				Flags: []FlagDef{
					{Name: "status", Description: "Target status (Draft, Open, ReadyForReview, Closed, Void)", Type: "string", Required: true},
					{Name: "comment", Description: "Optional transition comment", Type: "string"},
				},
			},
		},
	})

	Register(&Domain{
		Name:        "rfi",
		Aliases:     []string{"rfis"},
		Description: "Manage construction RFIs (requests for information)",
		APIPath:     "/api/rfi",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List construction RFIs",
				ToolName:    "UteamupRfiList",
				Flags: append([]FlagDef{
					{Name: "project-guid", Description: "Restrict to one project GUID", Type: "string"},
					{Name: "status", Description: "Filter by status (Draft, Submitted, Answered, Closed, Void)", Type: "string"},
					{Name: "discipline", Description: "Filter by discipline (Architectural, Structural, Mechanical, Electrical, ...)", Type: "string"},
					{Name: "search", Description: "Free-text search over number/subject/question", Type: "string"},
				}, paginationFlags()...),
			},
			{Name: "get", Description: "Get an RFI by GUID", ToolName: "UteamupRfiGet", Args: externalGUIDArg()},
			// Records the official response of record (Submitted → Answered).
			{
				Name:        "respond",
				Description: "Record the official response on a submitted RFI by GUID",
				ToolName:    "UteamupRfiRespond",
				HTTPMethod:  "POST",
				RESTPath:    "{externalGuid}/respond",
				Args:        externalGUIDArg(),
				Flags: []FlagDef{
					{Name: "official-response", Description: "The official response of record", Type: "string", Required: true},
					{Name: "document-guids", Description: "Document GUIDs to attach with the response", Type: "stringSlice"},
				},
			},
		},
	})

	Register(&Domain{
		Name:        "submittal",
		Aliases:     []string{"submittals"},
		Description: "Manage the construction submittal register",
		APIPath:     "/api/submittal",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List construction submittals",
				ToolName:    "UteamupSubmittalList",
				Flags: append([]FlagDef{
					{Name: "project-guid", Description: "Restrict to one project GUID", Type: "string"},
					{Name: "status", Description: "Filter by status (Draft, Open, InReview, Closed, Void)", Type: "string"},
					{Name: "review-status", Description: "Filter by review outcome (Pending, Approved, ApprovedAsNoted, ReviseAndResubmit, Rejected, ForRecordOnly)", Type: "string"},
					{Name: "search", Description: "Free-text search over number/title/spec section", Type: "string"},
				}, paginationFlags()...),
			},
			{Name: "get", Description: "Get a submittal by GUID", ToolName: "UteamupSubmittalGet", Args: externalGUIDArg()},
			// Records the review outcome (Approved/AsNoted/Rejected/ForRecordOnly
			// close; ReviseAndResubmit reopens).
			{
				Name:        "review",
				Description: "Record the review outcome on an in-review submittal by GUID",
				ToolName:    "UteamupSubmittalReview",
				HTTPMethod:  "POST",
				RESTPath:    "{externalGuid}/review",
				Args:        externalGUIDArg(),
				Flags: []FlagDef{
					{Name: "review-status", Description: "Outcome (Approved, ApprovedAsNoted, ReviseAndResubmit, Rejected, ForRecordOnly)", Type: "string", Required: true},
					{Name: "comment", Description: "Reviewer comment", Type: "string"},
					{Name: "markup-document-guids", Description: "Marked-up document GUIDs returned with the review", Type: "stringSlice"},
				},
			},
		},
	})

	Register(&Domain{
		Name:        "dailylog",
		Aliases:     []string{"dailylogs", "daily-log"},
		Description: "Manage construction daily logs",
		APIPath:     "/api/dailylog",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List construction daily logs (newest day first)",
				ToolName:    "UteamupDailylogList",
				Flags: append([]FlagDef{
					{Name: "project-guid", Description: "Restrict to one project GUID", Type: "string"},
					{Name: "from", Description: "From date, inclusive (ISO 8601)", Type: "string"},
					{Name: "to", Description: "To date, inclusive (ISO 8601)", Type: "string"},
					{Name: "status", Description: "Filter by status (Draft, Signed)", Type: "string"},
				}, paginationFlags()...),
			},
			{Name: "get", Description: "Get a daily log by GUID", ToolName: "UteamupDailylogGet", Args: externalGUIDArg()},
			// Idempotent get-or-create for one project + calendar day; the
			// backend route is POST /api/dailylog/get-or-create.
			{
				Name:        "create",
				Description: "Get or create the project's daily log for a calendar day (idempotent)",
				ToolName:    "UteamupDailylogCreate",
				RESTPath:    "get-or-create",
				Flags: []FlagDef{
					{Name: "project-guid", Description: "Owning project GUID", Type: "string", Required: true},
					{Name: "log-date", Description: "The calendar day (ISO 8601; date component used)", Type: "string", Required: true},
					{Name: "location-guid", Description: "Weather site location GUID (defaults from project context)", Type: "string"},
				},
			},
		},
	})

	Register(&Domain{
		Name:        "construction-sheet",
		Aliases:     []string{"construction-sheets"},
		Description: "Browse the construction sheet register",
		APIPath:     "/api/constructionsheet",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the construction sheet register",
				ToolName:    "UteamupConstructionSheetList",
				Flags: append([]FlagDef{
					{Name: "project-guid", Description: "Restrict to one project GUID", Type: "string"},
					{Name: "discipline", Description: "Filter by discipline (Architectural, Structural, Mechanical, Electrical, ...)", Type: "string"},
					{Name: "status", Description: "Filter by register status (Current, Superseded, Void)", Type: "string"},
					{Name: "search", Description: "Free-text search over sheet number/title", Type: "string"},
				}, paginationFlags()...),
			},
		},
	})
}
