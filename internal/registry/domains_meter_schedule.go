package registry

// Mirrors the MCP UteamupMeterschedule* tools backed by
// MeterReadingScheduleController on the backend. The domain is Guid-first
// per the GUIDs-In/Integer-Ids-Out rule — every action takes the asset's
// or the schedule's external Guid rather than the internal int ids. The
// runtime in registry.go calls apiClient.CallREST(...) against the
// declared APIPath, so action Name + (optional) RESTPath build the URL.
//
// REST surface (preferred, Guid-keyed):
//   GET    /api/meter-reading-schedules/{guid}                            — fetch one
//   POST   /api/meter-reading-schedules                                   — create (body carries assetGuid)
//   PUT    /api/meter-reading-schedules/{guid}                            — update
//   DELETE /api/meter-reading-schedules/{guid}                            — deactivate
//   POST   /api/meter-reading-schedules/initialize/asset/{assetGuid}      — auto-create from asset type
//   GET    /api/meter-reading-schedules/compliance/asset/{assetGuid}      — per-asset compliance
//   GET    /api/meter-reading-schedules/compliance/summary                — tenant-wide compliance (paginated)
//
// Legacy int-keyed routes are still callable as `[Obsolete]` deprecation
// shims on the backend; the CLI surface intentionally omits them so new
// users don't reach for the deprecated identifiers.
func init() {
	Register(&Domain{
		Name:        "meter-schedule",
		Aliases:     []string{"meter-reading-schedule", "meter-schedules", "ms"},
		Description: "Manage meter reading schedules on assets (Guid-keyed)",
		APIPath:     "/api/meter-reading-schedules",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get a single meter reading schedule by its external Guid",
				ToolName:    "UteamupMeterscheduleGet",
				Args: []ArgDef{
					{Name: "guid", Description: "Schedule external Guid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "create",
				Description: "Create a meter reading schedule for an asset attribute",
				ToolName:    "UteamupMeterscheduleCreateByGuid",
				Flags: []FlagDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
					{Name: "attribute-definition-guid", Description: "Attribute definition external Guid (must have IsMeter=true)", Required: true, Type: "string"},
					{Name: "interval-seconds", Description: "Reading interval in seconds (minimum 300). Required when recurrence-type is none/omitted; ignored for calendar recurrence.", Type: "int"},
					{Name: "recurrence-type", Description: "Recurrence anchoring: none (default), weekly, monthly, or yearly", Type: "string"},
					{Name: "days-of-week", BodyName: "daysOfWeek", Description: "Weekly only: day numbers (0=Sunday … 6=Saturday). Repeatable, e.g. --days-of-week 1 --days-of-week 4", Type: "stringSlice"},
					{Name: "day-of-month-mode", Description: "Monthly/Yearly: specificDay, firstDay, lastDay, firstWorkday, or lastWorkday", Type: "string"},
					{Name: "day-of-month", Description: "Monthly/Yearly with mode=specificDay: the day-of-month (1–31)", Type: "int"},
					{Name: "month-of-year", Description: "Yearly only: the month (1–12)", Type: "int"},
					{Name: "label", Description: "Optional human-readable schedule label", Type: "string"},
					{Name: "preferred-time", Description: "Optional preferred time of day (HH:mm:ss)", Type: "string"},
					{Name: "timezone", Description: "Optional IANA time zone (e.g. UTC, Europe/London, Atlantic/Reykjavik)", Type: "string"},
					{Name: "auto-workorder", Description: "Auto-create a workorder when overdue", Type: "bool"},
					{Name: "workorder-template-guid", Description: "Optional workorder template external Guid to use when auto-creating", Type: "string"},
				},
			},
			{
				Name:        "update",
				Description: "Update a meter reading schedule by its external Guid",
				ToolName:    "UteamupMeterscheduleUpdateByGuid",
				Args: []ArgDef{
					{Name: "guid", Description: "Schedule external Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "interval-seconds", Description: "New reading interval in seconds (minimum 300, recurrence-type none)", Type: "int"},
					{Name: "recurrence-type", Description: "Recurrence anchoring: none, weekly, monthly, or yearly. Sending this overwrites the whole recurrence block — supply the matching day/month flags too.", Type: "string"},
					{Name: "days-of-week", BodyName: "daysOfWeek", Description: "Weekly only: day numbers (0=Sunday … 6=Saturday). Repeatable.", Type: "stringSlice"},
					{Name: "day-of-month-mode", Description: "Monthly/Yearly: specificDay, firstDay, lastDay, firstWorkday, or lastWorkday", Type: "string"},
					{Name: "day-of-month", Description: "Monthly/Yearly with mode=specificDay: the day-of-month (1–31)", Type: "int"},
					{Name: "month-of-year", Description: "Yearly only: the month (1–12)", Type: "int"},
					{Name: "label", Description: "New schedule label", Type: "string"},
					{Name: "preferred-time", Description: "New preferred time of day (HH:mm:ss)", Type: "string"},
					{Name: "timezone", Description: "New IANA time zone", Type: "string"},
					{Name: "auto-workorder", Description: "Toggle auto-create workorder on overdue", Type: "bool"},
					{Name: "workorder-template-guid", Description: "Optional workorder template external Guid", Type: "string"},
					{Name: "is-active", Description: "Activate or deactivate the schedule", Type: "bool"},
				},
			},
			{
				Name:        "delete",
				Description: "Deactivate (soft-delete) a meter reading schedule by external Guid",
				ToolName:    "UteamupMeterscheduleDeactivateByGuid",
				Args: []ArgDef{
					{Name: "guid", Description: "Schedule external Guid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "list",
				Description: "List meter reading schedules, optionally filtered by asset Guid",
				ToolName:    "UteamupMeterscheduleList",
				Flags: append(paginationFlags(),
					FlagDef{Name: "asset-guid", Description: "Filter by asset external Guid", Type: "string"},
					FlagDef{Name: "overdue-only", Description: "Return only overdue schedules", Type: "bool"},
				),
			},
			{
				Name:        "overdue",
				Description: "List overdue meter reading schedules",
				ToolName:    "UteamupMeterscheduleGetOverdue",
				Flags:       paginationFlags(),
			},
			{
				Name:        "compliance-summary",
				Description: "Tenant-wide meter reading compliance summary",
				ToolName:    "UteamupMeterscheduleComplianceSummary",
				HTTPMethod:  "GET",
				RESTPath:    "compliance/summary",
				Flags: append(paginationFlags(),
					FlagDef{Name: "overdue-only", Description: "Include only overdue schedules", Type: "bool"},
				),
			},
			{
				Name:        "compliance-asset",
				Description: "Compliance breakdown for one asset by external Guid",
				ToolName:    "UteamupMeterscheduleComplianceForAssetByGuid",
				HTTPMethod:  "GET",
				RESTPath:    "compliance/asset/{assetGuid}",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "initialize",
				Description: "Auto-create schedules from an asset type's meter attribute defaults",
				ToolName:    "UteamupMeterscheduleInitializeByGuid",
				HTTPMethod:  "POST",
				RESTPath:    "initialize/asset/{assetGuid}",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "create-workorder",
				Description: "Force-create a workorder for a schedule now (independent of auto-create), optionally from a workorder template",
				ToolName:    "UteamupMeterscheduleCreateWorkorder",
				HTTPMethod:  "POST",
				RESTPath:    "{guid}/create-workorder",
				Args: []ArgDef{
					{Name: "guid", Description: "Schedule external Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "use-schedule-template", Description: "Seed from a template (override below, else the schedule's saved template). Set false for a generic MET workorder.", Default: true, Type: "bool"},
					{Name: "workorder-template-guid", Description: "Optional workorder template external Guid to seed from (overrides the schedule's saved template)", Type: "string"},
				},
			},
		},
	})
}
