package registry

func init() {
	Register(&Domain{
		Name:        "oncall",
		Aliases:     []string{"on-call"},
		Description: "Read the on-call rota",
		APIPath:     "/api/oncall",
		Actions: []Action{
			{
				Name:        "who",
				Description: "Who is on call for a schedule at an instant (defaults to now)",
				ToolName:    "UteamupOnCallWho",
				HTTPMethod:  "GET",
				RESTPath:    "{schedule-guid}/who",
				Args: []ArgDef{
					{Name: "schedule-guid", Description: "On-call schedule external Guid", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "at", Description: "Instant to resolve (ISO-8601 UTC). Defaults to now.", Type: "string"},
				},
			},
			{
				Name:        "schedule-list",
				Description: "List the tenant's on-call schedules",
				ToolName:    "UteamupOnCallScheduleList",
				HTTPMethod:  "GET",
				RESTPath:    "schedules",
			},
			{
				Name:        "callout-summary",
				Description: "Summarize on-call callout operations, SLA breaches, repeat cycles, and manager reviews",
				ToolName:    "UteamupOnCallCalloutSummary",
				HTTPMethod:  "GET",
				RESTPath:    "callouts/summary",
			},
			{
				// ToolName casing follows the backend-registered tool name for
				// this action ("Oncall"), which predates the "OnCall" casing
				// used by the older tools in this file.
				Name:        "callouts",
				Description: "List tenant callouts needing on-call attention (defaults to open/escalated only)",
				ToolName:    "UteamupOncallActiveCallouts",
				HTTPMethod:  "GET",
				RESTPath:    "callouts",
				Flags: []FlagDef{
					{Name: "include-closed", Description: "Include closed callouts as well", Type: "bool"},
				},
			},
			{
				Name:        "calendar",
				Description: "Download an authenticated iCalendar feed for an on-call schedule",
				ToolName:    "UteamupOnCallCalendar",
				HTTPMethod:  "GET",
				RESTPath:    "{schedule-guid}/calendar.ics",
				Args: []ArgDef{
					{Name: "schedule-guid", Description: "On-call schedule external Guid", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "from", Description: "Calendar window start (ISO-8601 UTC). Defaults to now.", Type: "string"},
					{Name: "to", Description: "Calendar window end (ISO-8601 UTC). Defaults to 30 days after from.", Type: "string"},
				},
			},
			{
				Name:        "calendar-subscription-get",
				Description: "Inspect a schedule's revocable public iCalendar subscription status",
				ToolName:    "UteamupOnCallCalendarSubscriptionGet",
				HTTPMethod:  "GET",
				RESTPath:    "{schedule-guid}/calendar-subscription",
				Args: []ArgDef{
					{Name: "schedule-guid", Description: "On-call schedule external Guid", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "calendar-subscription-rotate",
				Description: "Create or rotate a schedule's public iCalendar subscription link",
				ToolName:    "UteamupOnCallCalendarSubscriptionRotate",
				HTTPMethod:  "POST",
				RESTPath:    "{schedule-guid}/calendar-subscription",
				Args: []ArgDef{
					{Name: "schedule-guid", Description: "On-call schedule external Guid", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "calendar-subscription-revoke",
				Description: "Revoke a schedule's public iCalendar subscription link",
				ToolName:    "UteamupOnCallCalendarSubscriptionRevoke",
				HTTPMethod:  "DELETE",
				RESTPath:    "{schedule-guid}/calendar-subscription",
				Args: []ArgDef{
					{Name: "schedule-guid", Description: "On-call schedule external Guid", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "schedule-create",
				Description: "Create an on-call schedule",
				ToolName:    "UteamupOnCallScheduleCreate",
				HTTPMethod:  "POST",
				RESTPath:    "schedules",
				Flags: []FlagDef{
					{Name: "name", Description: "Rota name", Type: "string", Required: true},
					{Name: "timeZone", Description: "IANA time zone for handoff times", Type: "string"},
				},
			},
			{
				Name:        "layer-add",
				Description: "Add a rotation layer (ordered people + cadence) to a schedule",
				ToolName:    "UteamupOnCallLayerAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{schedule-guid}/layers",
				Args: []ArgDef{
					{Name: "schedule-guid", Description: "On-call schedule external Guid", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "user", Description: "Ordered user GUID (repeat for each; rotation order preserved)", Type: "stringSlice", Required: true, BodyName: "orderedUserGuids"},
					{Name: "shift-minutes", Description: "Minutes each person holds the pager (10080 = weekly)", Type: "int", Required: true, BodyName: "shiftLengthMinutes"},
					{Name: "start", Description: "Rotation slot-0 start anchor (ISO-8601 UTC)", Type: "string", Required: true, BodyName: "startAnchor"},
					{Name: "precedence", Description: "Higher wins when layers overlap", Type: "int", Default: 1},
					{Name: "days-mask", Description: "Sun=1..Sat=64 bitmask; 0 = every day", Type: "int", BodyName: "daysOfWeekMask"},
				},
			},
			{
				Name:        "override-add",
				Description: "Add a one-off override (someone covers the pager for a window; wins over layers)",
				ToolName:    "UteamupOnCallOverrideAdd",
				HTTPMethod:  "POST",
				RESTPath:    "{schedule-guid}/overrides",
				Args: []ArgDef{
					{Name: "schedule-guid", Description: "On-call schedule external Guid", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "user", Description: "The covering user's GUID", Type: "string", Required: true, BodyName: "targetUserGuid"},
					{Name: "start", Description: "Window start (ISO-8601 UTC)", Type: "string", Required: true, BodyName: "startAt"},
					{Name: "end", Description: "Window end (ISO-8601 UTC, after start)", Type: "string", Required: true, BodyName: "endAt"},
				},
			},
			{
				Name:        "classify-standby",
				Description: "Classify whether a standby period counts as working time (ECJ test)",
				ToolName:    "UteamupOnCallClassifyStandby",
				HTTPMethod:  "POST",
				RESTPath:    "classify-standby",
				Flags: []FlagDef{
					{Name: "response-minutes", Description: "Required response time in minutes (omit = no hard obligation)", Type: "int", BodyName: "responseTimeMinutes"},
					{Name: "callouts-per-week", Description: "Observed callouts per week", Type: "float", Default: 0.0, BodyName: "calloutsPerWeek"},
					{Name: "freedom", Description: "0=Reachable, 1=Restricted, 2=ConfinedToPremises", Type: "int", Default: 0},
					{Name: "override", Description: "Human determination (wins when set)", Type: "bool", BodyName: "countsAsWorkingTimeOverride"},
				},
			},
		},
	})
}
