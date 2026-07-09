package registry

func init() {
	Register(&Domain{Name: "schedule", Aliases: []string{"schedules"}, Description: "Manage schedules", Actions: crudActions("Schedule")})
	Register(&Domain{Name: "schedule-assignment", Description: "Manage schedule assignments", Actions: crudActions("ScheduleAssignment")})
	Register(&Domain{Name: "shift", Aliases: []string{"shifts"}, Description: "Manage shifts", Actions: crudActions("Shift")})
	Register(&Domain{Name: "shift-template", Description: "Manage shift templates", Actions: crudActions("ShiftTemplate")})
	Register(&Domain{Name: "shift-instance", Description: "Manage shift instances", Actions: crudActions("ShiftInstance")})
	Register(&Domain{Name: "shift-request", Description: "Manage shift requests", Actions: crudActions("ShiftRequest")})
	Register(&Domain{Name: "shift-assignment", Description: "Manage shift user assignments", Actions: crudActions("ShiftUserAssignment")})
	shiftHandoverActions := crudActions("ShiftHandover")
	shiftHandoverActions = append(shiftHandoverActions, Action{
		Name:        "generate-summary",
		Description: "Generate an editable AI handover draft (5 credits; never submits)",
		ToolName:    "UteamupShiftHandoverGenerateSummary",
		HTTPMethod:  "POST",
		RESTPath:    "by-guid/{handoverGuid}/generate-summary",
		Args:        []ArgDef{{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "string"}},
		Flags: []FlagDef{
			{Name: "language", Description: "Response language code", Default: "en", Type: "string"},
			{Name: "summary-type", Description: "brief, detailed, or technical", Default: "detailed", Type: "string", BodyName: "summaryType"},
		},
	})
	Register(&Domain{Name: "shift-handover", Description: "Manage shift handovers", Actions: shiftHandoverActions})
	Register(&Domain{Name: "time-entry", Aliases: []string{"time", "timesheet"}, Description: "Manage time entries", Actions: crudActions("TimeEntry")})
}
