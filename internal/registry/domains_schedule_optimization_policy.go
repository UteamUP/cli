package registry

func init() {
	mutationFlags := []FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Caller-generated GUID reused only when retrying the same mutation", Required: true, Type: "string"},
		{Name: "name", BodyName: "name", Description: "Plain-language policy name", Required: true, Type: "string"},
		{Name: "enabled", BodyName: "isEnabled", Description: "Allow the policy to queue review-only proposals", Type: "bool", Default: true},
		{Name: "frequency", BodyName: "frequency", Description: "daily (every selected day) or weekly (one selected day per week)", Type: "string", Default: "daily"},
		{Name: "time-zone", BodyName: "timeZoneId", Description: "IANA or system time zone identifier", Type: "string", Default: "UTC"},
		{Name: "days-mask", BodyName: "daysOfWeekMask", Description: "Sun=1 through Sat=64 selected-day bit mask", Type: "int", Default: 127},
		{Name: "local-time", BodyName: "localExecutionTime", Description: "Local execution time using HH:mm", Type: "string", Default: "06:00"},
		{Name: "horizon-days", BodyName: "horizonDays", Description: "Planning horizon in days (1-30)", Type: "int", Default: 7},
		{Name: "team-guids", BodyName: "teamGuids", Description: "One or more tenant-scoped team GUIDs", Required: true, Type: "stringSlice"},
	}

	Register(&Domain{
		Name:        "schedule-optimization-policy",
		Aliases:     []string{"schedule-policy", "schedule-opt-policy"},
		Description: "Manage recurring policies that queue review-only schedule optimization proposals",
		APIPath:     "/api/schedule/optimization-policies",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List recurring review-only schedule optimization policies",
				ToolName:    "UteamupScheduleOptimizationPolicyList",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "include-archived", BodyName: "includeArchived", Description: "Include archived policies alongside the active ones", Type: "bool", Default: false},
				},
			},
			{
				Name:        "get",
				Description: "Read one recurring policy and its latest outcome",
				ToolName:    "UteamupScheduleOptimizationPolicyGet",
				HTTPMethod:  "GET",
				RESTPath:    "{policyGuid}",
				Args: []ArgDef{
					{Name: "policyGuid", Description: "Schedule optimization policy GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create a policy that queues proposals but never applies them automatically",
				ToolName:    "UteamupScheduleOptimizationPolicyCreate",
				HTTPMethod:  "POST",
				Flags:       mutationFlags,
			},
			{
				Name:        "update",
				Description: "Update recurrence, team scope, horizon, or enabled state",
				ToolName:    "UteamupScheduleOptimizationPolicyUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{policyGuid}",
				Args: []ArgDef{
					{Name: "policyGuid", Description: "Schedule optimization policy GUID", Required: true, Type: "uuid"},
				},
				Flags: mutationFlags,
			},
			{
				Name:        "delete",
				Description: "Soft-delete a recurring policy while preserving audit evidence",
				ToolName:    "UteamupScheduleOptimizationPolicyDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "{policyGuid}",
				Args: []ArgDef{
					{Name: "policyGuid", Description: "Schedule optimization policy GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Caller-generated GUID reused only when retrying the same deletion", Required: true, Type: "string"},
				},
			},
			{
				Name:        "restore",
				Description: "Restore an archived policy as a disabled draft",
				ToolName:    "UteamupScheduleOptimizationPolicyRestore",
				HTTPMethod:  "POST",
				RESTPath:    "{policyGuid}/restore",
				Args: []ArgDef{
					{Name: "policyGuid", Description: "Schedule optimization policy GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
