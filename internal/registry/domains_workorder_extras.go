package registry

func init() {
	Register(&Domain{Name: "workorder-template", Aliases: []string{"wot"}, Description: "Manage work order templates", Actions: append(crudActions("WorkorderTemplate"),
		Action{
			Name:        "active",
			Description: "List bounded active tenant workorder templates",
			ToolName:    "UteamupWorkorderTemplateGetActive",
			HTTPMethod:  "GET",
			Flags: []FlagDef{
				{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
				{Name: "page-size", Short: "s", Description: "Items per page (max 100)", Default: 20, Type: "int"},
				{Name: "is-active", Description: "Only return active templates", Default: true, Type: "bool"},
				{Name: "sort-by", Description: "Sort field", Default: "Name", Type: "string"},
				{Name: "sort-order", Description: "Sort order: asc or desc", Default: "asc", Type: "string"},
			},
		},
		Action{
			Name:        "create-workorder",
			Description: "Create an open work order from an active template, with optional scalar and public-GUID relationship overrides.",
			ToolName:    "UteamupWorkorderTemplateCreateFromTemplateByGuid",
			HTTPMethod:  "POST",
			RESTPath:    "{templateGuid}/create-workorder",
			Flags: []FlagDef{
				{Name: "template", BodyName: "templateGuid", Description: "Public GUID of the workorder template (required)", Required: true, Type: "string"},
				{Name: "name", Description: "Optional override for the new work order name", Type: "string"},
				{Name: "description", Description: "Optional override for the new work order description", Type: "string"},
				{Name: "maintenance-type", BodyName: "maintenanceType", Description: "Optional EN 13306 maintenance type number", Type: "int"},
				{Name: "status", Description: "Optional work order status override", Type: "int"},
				{Name: "priority", Description: "Optional priority override (1=Low … 5=Critical)", Type: "int"},
				{Name: "notes", Description: "Optional notes override", Type: "string"},
				{Name: "asset-guid", BodyName: "assetGuid", Description: "Optional single asset public GUID", Type: "string"},
				{Name: "asset-guids", BodyName: "assetGuids", Description: "Complete asset GUID set; empty clears and omission inherits", Type: "stringSlice"},
				{Name: "part-guids", BodyName: "partGuids", Description: "Complete part GUID set", Type: "stringSlice"},
				{Name: "tool-guids", BodyName: "toolGuids", Description: "Complete tool GUID set", Type: "stringSlice"},
				{Name: "chemical-guids", BodyName: "chemicalGuids", Description: "Complete chemical GUID set", Type: "stringSlice"},
				{Name: "task-list-guids", BodyName: "taskListGuids", Description: "Complete task-list GUID set", Type: "stringSlice"},
				{Name: "check-list-guids", BodyName: "checkListGuids", Description: "Complete checklist GUID set", Type: "stringSlice"},
				{Name: "location-guid", BodyName: "locationGuid", Description: "Optional location public GUID", Type: "string"},
				{Name: "location-floor-guid", BodyName: "locationFloorGuid", Description: "Optional floor or area public GUID", Type: "string"},
				{Name: "primary-assignee-guid", BodyName: "primaryAssigneeGuid", Description: "Optional tenant-user public GUID", Type: "string"},
				{Name: "leave-unassigned", BodyName: "leaveUnassigned", Description: "Suppress the template/current-user assignee fallback", Type: "bool"},
				{Name: "stop-required", BodyName: "stopRequired", Description: "Optional equipment-stop override", Type: "bool"},
				{Name: "estimated-duration", BodyName: "estimatedDuration", Description: "Optional TimeSpan duration, for example 02:30:00", Type: "string"},
				{Name: "estimated-cost", BodyName: "estimatedCost", Description: "Optional non-negative estimated cost", Type: "float"},
				{Name: "start-date-utc", BodyName: "startDateUtc", Description: "Optional ISO-8601 UTC start", Type: "string"},
				{Name: "due-date-utc", BodyName: "dueDateUtc", Description: "Optional ISO-8601 UTC due time", Type: "string"},
				{Name: "idempotency-key", HeaderName: "Idempotency-Key", Description: "Stable retry key for this create", Type: "string"},
			},
		},
		Action{
			Name:        "run-schedule-now",
			Description: "Generate one workorder right now from a configured template schedule (the manual \"Generate now\" QA action). Reuses the existing template-to-workorder path, updates LastCreatedDate / WorkordersCreated / LastGeneratedWorkorderGuid on the schedule, and does NOT advance NextCreateDate.",
			ToolName:    "UteamupWorkorderTemplateRunScheduleNow",
			Args: []ArgDef{
				{Name: "scheduleGuid", Description: "External GUID of the workorder-template schedule (required)", Required: true, Type: "string"},
			},
		},
		Action{
			Name:        "analysis-preview",
			Description: "Preview how many completed workorders are linked to a template and the AI-credit cost to analyze them (5 credits each). Does NOT charge credits.",
			ToolName:    "UteamupWorkorderTemplateAnalyzePreview",
			Flags: []FlagDef{
				{Name: "template", BodyName: "templateGuid", Description: "Public GUID of the workorder template (required)", Required: true, Type: "string"},
			},
		},
		Action{
			Name:        "analyze",
			Description: "Analyze a template's completed workorders with AI and return suggested enhancements (description, checklist/task lists, tools/chemicals, estimated duration/cost). Charges 5 AI credits per analyzed workorder; preview first with analysis-preview.",
			ToolName:    "UteamupWorkorderTemplateAnalyze",
			Flags: []FlagDef{
				{Name: "template", BodyName: "templateGuid", Description: "Public GUID of the workorder template (required)", Required: true, Type: "string"},
			},
		},
	)})
	Register(&Domain{Name: "workorder-signature", Description: "Manage work order signatures", Actions: crudActions("WorkorderSignature")})
	Register(&Domain{Name: "workorder-watchlist", Description: "Manage work order watchlists", Actions: crudActions("WorkorderWatchlist")})
	Register(&Domain{Name: "tasklist", Aliases: []string{"tasks"}, Description: "Manage task lists", Actions: crudActions("TaskList")})
	Register(&Domain{Name: "checklist", Aliases: []string{"checklists"}, Description: "Manage checklists", Actions: crudActions("CheckList")})
	Register(&Domain{Name: "language", Aliases: []string{"lang"}, Description: "Language utilities (AI translation)", Actions: []Action{
		{
			Name:        "translate",
			Description: "Translate authored content into other languages using AI. Charges 2 AI credits per target language (supported: en, is, pl, de, es).",
			ToolName:    "UteamupLanguageTranslate",
			Flags: []FlagDef{
				{Name: "source-text", BodyName: "sourceText", Description: "The text to translate (required)", Required: true, Type: "string"},
				{Name: "source-lang", BodyName: "sourceLanguage", Description: "Source language code: en|is|pl|de|es (required)", Required: true, Type: "string"},
				{Name: "target-langs", BodyName: "targetLanguages", Description: "Target language codes — repeatable or comma-separated (subset of en|is|pl|de|es)", Required: true, Type: "stringSlice"},
			},
		},
	}})
}
