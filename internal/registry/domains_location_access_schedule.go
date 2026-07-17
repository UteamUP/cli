package registry

func init() {
	writeFlags := []FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Caller-generated GUID reused only when retrying the same mutation", Required: true, Type: "uuid"},
		{Name: "name", BodyName: "name", Description: "Plain-language access window name", Required: true, Type: "string"},
		{Name: "description", BodyName: "description", Description: "Optional site-entry instructions", Type: "string"},
		{Name: "days-mask", BodyName: "daysOfWeekMask", Description: "Sun=1 through Sat=64 selected-day bit mask", Type: "int", Default: 127},
		{Name: "access-start", BodyName: "accessStartLocal", Description: "Local access start using HH:mm", Type: "string"},
		{Name: "access-end", BodyName: "accessEndLocal", Description: "Local access end using HH:mm", Type: "string"},
		{Name: "all-day", BodyName: "isAllDay", Description: "Allow access all day on selected days", Type: "bool", Default: false},
		{Name: "effective-from", BodyName: "effectiveFromDate", Description: "Optional first effective date in ISO 8601 format", Type: "string"},
		{Name: "effective-to", BodyName: "effectiveToDate", Description: "Optional last effective date in ISO 8601 format", Type: "string"},
		{Name: "time-zone", BodyName: "timeZoneId", Description: "IANA or system time zone identifier", Type: "string", Default: "UTC"},
		{Name: "active", BodyName: "isActive", Description: "Enforce this window during scheduling", Type: "bool", Default: true},
	}

	Register(&Domain{
		Name:        "location-access-schedule",
		Aliases:     []string{"location-access", "site-access"},
		Description: "Manage recurring site-access windows used as hard scheduling constraints",
		APIPath:     "/api/location",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List access windows for one public location GUID",
				ToolName:    "UteamupLocationAccessScheduleList",
				HTTPMethod:  "GET",
				RESTPath:    "{locationGuid}/access-schedules",
				Args: []ArgDef{
					{Name: "locationGuid", Description: "Location GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "get",
				Description: "Read one access window and its concurrency evidence",
				ToolName:    "UteamupLocationAccessScheduleGet",
				HTTPMethod:  "GET",
				RESTPath:    "access-schedules/{scheduleGuid}",
				Args: []ArgDef{
					{Name: "scheduleGuid", Description: "Location access schedule GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create an idempotent recurring site-access window",
				ToolName:    "UteamupLocationAccessScheduleCreate",
				HTTPMethod:  "POST",
				RESTPath:    "{locationGuid}/access-schedules",
				Args: []ArgDef{
					{Name: "locationGuid", Description: "Location GUID", Required: true, Type: "uuid"},
				},
				Flags: writeFlags,
			},
			{
				Name:        "update",
				Description: "Update a site-access window with optimistic concurrency",
				ToolName:    "UteamupLocationAccessScheduleUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "access-schedules/{scheduleGuid}",
				Args: []ArgDef{
					{Name: "scheduleGuid", Description: "Location access schedule GUID", Required: true, Type: "uuid"},
				},
				Flags: append([]FlagDef{
					{Name: "expected-updated-at", BodyName: "expectedUpdatedAtUtc", Description: "Current UpdatedAtUtc value in ISO 8601 format", Required: true, Type: "string"},
				}, writeFlags...),
			},
			{
				Name:        "delete",
				Description: "Soft-delete a site-access window with optimistic concurrency",
				ToolName:    "UteamupLocationAccessScheduleDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "access-schedules/{scheduleGuid}",
				Args: []ArgDef{
					{Name: "scheduleGuid", Description: "Location access schedule GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Caller-generated retry-safe GUID", Required: true, Type: "uuid"},
					{Name: "expected-updated-at", BodyName: "expectedUpdatedAtUtc", Description: "Current UpdatedAtUtc value in ISO 8601 format", Required: true, Type: "string"},
				},
			},
		},
	})
}
