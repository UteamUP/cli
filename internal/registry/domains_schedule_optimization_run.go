package registry

func init() {
	Register(&Domain{
		Name:        "schedule-optimization-run",
		Aliases:     []string{"schedule-optimization", "schedule-opt"},
		Description: "Create, inspect, apply, revert, and cancel durable schedule optimization proposals",
		APIPath:     "/api/schedule/optimization-runs",
		Actions: []Action{
			{
				Name:        "create",
				Description: "Create a review-first optimization run that never applies schedule changes automatically",
				ToolName:    "UteamupScheduleOptimizationRunCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Caller-generated GUID reused only when retrying the same request", Required: true, Type: "string"},
					{Name: "week-start", BodyName: "weekStart", Description: "UTC start of the planning week", Required: true, Type: "string"},
					{Name: "workorder-guids", BodyName: "workorderGuids", Description: "Tenant-scoped workorder GUIDs", Required: true, Type: "stringSlice"},
					{Name: "technician-guids", BodyName: "technicianGuids", Description: "Eligible tenant member GUIDs", Required: true, Type: "stringSlice"},
					{Name: "team-guid", BodyName: "teamGuid", Description: "Optional tenant-scoped team GUID", Type: "string"},
					{Name: "horizon-days", BodyName: "horizonDays", Description: "Planning horizon in days (1-30)", Type: "int", Default: 7},
				},
			},
			{
				Name:        "get",
				Description: "Read durable status, evidence, and the optional optimization proposal",
				ToolName:    "UteamupScheduleOptimizationRunGet",
				HTTPMethod:  "GET",
				RESTPath:    "{runGuid}",
				Args: []ArgDef{
					{Name: "runGuid", Description: "Schedule optimization run GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "cancel",
				Description: "Request cancellation without deleting the optimization audit record",
				ToolName:    "UteamupScheduleOptimizationRunCancel",
				HTTPMethod:  "POST",
				RESTPath:    "{runGuid}/cancel",
				Args: []ArgDef{
					{Name: "runGuid", Description: "Schedule optimization run GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "reason", Description: "Optional plain-language cancellation reason", Type: "string"},
				},
			},
			{
				Name:        "apply",
				Description: "Apply all or selected workorders from a completed optimization proposal after revalidating evidence",
				ToolName:    "UteamupScheduleOptimizationRunApply",
				HTTPMethod:  "POST",
				RESTPath:    "{runGuid}/apply",
				Args: []ArgDef{
					{Name: "runGuid", Description: "Schedule optimization run GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Caller-generated GUID reused only when retrying the same apply", Required: true, Type: "string"},
					{Name: "selected-workorder-guids", BodyName: "selectedWorkorderGuids", Description: "Optional workorder GUID subset; omit to apply every proposal item", Type: "stringSlice"},
				},
			},
			{
				Name:        "revert",
				Description: "Cancel exactly the assignments created by an applied optimization run while preserving audit evidence",
				ToolName:    "UteamupScheduleOptimizationRunRevert",
				HTTPMethod:  "POST",
				RESTPath:    "{runGuid}/revert",
				Args: []ArgDef{
					{Name: "runGuid", Description: "Schedule optimization run GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Caller-generated GUID reused only when retrying the same revert", Required: true, Type: "string"},
					{Name: "reason", BodyName: "reason", Description: "Optional plain-language reason for restoring the previous plan", Type: "string"},
				},
			},
		},
	})
}
