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
			Description: "Create an open work order from a template, identified by the template's public GUID. No asset or resolution note required.",
			ToolName:    "UteamupWorkorderTemplateCreateFromTemplateByGuid",
			Flags: []FlagDef{
				{Name: "template", Description: "Public GUID of the workorder template (required)", Required: true, Type: "string"},
				{Name: "name", Description: "Optional override for the new work order name", Type: "string"},
				{Name: "description", Description: "Optional override for the new work order description", Type: "string"},
				{Name: "priority", Description: "Optional priority override (1=Low … 5=Critical)", Type: "int"},
				{Name: "notes", Description: "Optional notes override", Type: "string"},
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
