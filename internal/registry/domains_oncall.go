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
		},
	})
}
