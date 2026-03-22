package registry

func init() {
	Register(&Domain{Name: "schedule", Aliases: []string{"schedules"}, Description: "Manage schedules", Actions: crudActions("Schedule")})
	Register(&Domain{Name: "schedule-assignment", Description: "Manage schedule assignments", Actions: crudActions("ScheduleAssignment")})
	Register(&Domain{Name: "shift", Aliases: []string{"shifts"}, Description: "Manage shifts", Actions: crudActions("Shift")})
	Register(&Domain{Name: "shift-template", Description: "Manage shift templates", Actions: crudActions("ShiftTemplate")})
	Register(&Domain{Name: "shift-instance", Description: "Manage shift instances", Actions: crudActions("ShiftInstance")})
	Register(&Domain{Name: "shift-request", Description: "Manage shift requests", Actions: crudActions("ShiftRequest")})
	Register(&Domain{Name: "shift-assignment", Description: "Manage shift user assignments", Actions: crudActions("ShiftUserAssignment")})
	Register(&Domain{Name: "shift-handover", Description: "Manage shift handovers", Actions: crudActions("ShiftHandover")})
	Register(&Domain{Name: "time-entry", Aliases: []string{"time", "timesheet"}, Description: "Manage time entries", Actions: crudActions("TimeEntry")})
}
