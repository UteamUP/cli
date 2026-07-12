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
	shiftHandoverActions = append(shiftHandoverActions,
		Action{
			Name:        "submit",
			Description: "Submit a draft handover as its designated outgoing operator",
			ToolName:    "UteamupShiftHandoverSubmit",
			HTTPMethod:  "PUT",
			RESTPath:    "by-guid/{handoverGuid}/submit",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: handoverMutationFlags(),
		},
		Action{
			Name:        "pending-acceptances",
			Description: "List handovers awaiting acceptance by the current incoming operator",
			ToolName:    "UteamupShiftHandoverGetPendingAcceptances",
			HTTPMethod:  "GET",
			RESTPath:    "acceptances/pending",
		},
		Action{
			Name:        "start-review",
			Description: "Start reviewing a submitted handover as its designated incoming operator",
			ToolName:    "UteamupShiftHandoverStartReview",
			HTTPMethod:  "PUT",
			RESTPath:    "by-guid/{handoverGuid}/start-review",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: handoverMutationFlags(),
		},
		Action{
			Name:        "accept",
			Description: "Accept a handover as its designated incoming operator",
			ToolName:    "UteamupShiftHandoverAccept",
			HTTPMethod:  "PUT",
			RESTPath:    "by-guid/{handoverGuid}/accept",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: append(handoverMutationFlags(),
				FlagDef{Name: "notes", Description: "Optional acceptance notes", Type: "string"},
			),
		},
		Action{
			Name:        "complete",
			Description: "Complete an accepted handover as its designated incoming operator",
			ToolName:    "UteamupShiftHandoverComplete",
			HTTPMethod:  "PUT",
			RESTPath:    "by-guid/{handoverGuid}/complete",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: handoverMutationFlags(),
		},
		Action{
			Name:        "decline-acceptance",
			Description: "Decline a handover as its designated incoming operator",
			ToolName:    "UteamupShiftHandoverDeclineAcceptance",
			HTTPMethod:  "PUT",
			RESTPath:    "by-guid/{handoverGuid}/decline-acceptance",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: append(handoverMutationFlags(),
				FlagDef{Name: "notes", Description: "Optional decline reason", Type: "string"},
			),
		},
		Action{
			Name:        "sections",
			Description: "List handover sections by stable handover GUID",
			ToolName:    "UteamupShiftHandoverGetSections",
			HTTPMethod:  "GET",
			RESTPath:    "by-guid/{handoverGuid}/sections",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
		},
		Action{
			Name:        "section-create",
			Description: "Add a validated section to a draft or rejected handover",
			ToolName:    "UteamupShiftHandoverCreateSection",
			HTTPMethod:  "POST",
			RESTPath:    "by-guid/{handoverGuid}/sections",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: []FlagDef{
				{Name: "section-type", BodyName: "sectionType", Required: true, Type: "int", Description: "Section type enum value"},
				{Name: "title", Type: "string", Description: "Section title; required for custom sections"},
				{Name: "content", Type: "string", Description: "Section content"},
				{Name: "sort-order", BodyName: "sortOrder", Type: "int", Description: "Optional unique sort order"},
				{Name: "required", BodyName: "isRequired", Type: "bool", Description: "Require completion before submit"},
				handoverConcurrencyFlag(),
			},
		},
		Action{
			Name:        "section-update",
			Description: "Update a handover section by stable GUID",
			ToolName:    "UteamupShiftHandoverUpdateSection",
			HTTPMethod:  "PUT",
			RESTPath:    "by-guid/{handoverGuid}/sections/{sectionGuid}",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "sectionGuid", Description: "Section ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: []FlagDef{
				{Name: "title", Type: "string", Description: "Updated title"},
				{Name: "content", Type: "string", Description: "Updated content"},
				{Name: "completed", BodyName: "isCompleted", Type: "bool", Description: "Completion state"},
				{Name: "sort-order", BodyName: "sortOrder", Type: "int", Description: "Unique sort order"},
				handoverConcurrencyFlag(),
			},
		},
		Action{
			Name:        "section-delete",
			Description: "Delete an optional handover section by stable GUID",
			ToolName:    "UteamupShiftHandoverDeleteSection",
			HTTPMethod:  "DELETE",
			RESTPath:    "by-guid/{handoverGuid}/sections/{sectionGuid}",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "sectionGuid", Description: "Section ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: []FlagDef{handoverConcurrencyFlag()},
		},
		Action{
			Name:        "sections-reorder",
			Description: "Replace the absolute order of every handover section",
			ToolName:    "UteamupShiftHandoverReorderSections",
			HTTPMethod:  "PUT",
			RESTPath:    "by-guid/{handoverGuid}/sections/reorder",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: []FlagDef{
				{Name: "section-guids", BodyName: "sectionGuids", Required: true, Type: "stringSlice", Description: "Every section GUID in the desired order"},
				handoverConcurrencyFlag(),
			},
		},
	)
	Register(&Domain{Name: "shift-handover", Description: "Manage shift handovers", Actions: shiftHandoverActions})
	Register(&Domain{Name: "time-entry", Aliases: []string{"time", "timesheet"}, Description: "Manage time entries", Actions: crudActions("TimeEntry")})
}

func handoverConcurrencyFlag() FlagDef {
	return FlagDef{
		Name:        "concurrency-token",
		BodyName:    "concurrencyToken",
		Description: "Latest opaque concurrency token from the handover response",
		Required:    true,
		Type:        "string",
	}
}

func handoverMutationFlags() []FlagDef {
	return []FlagDef{
		{
			Name:        "concurrency-token",
			BodyName:    "concurrencyToken",
			Description: "Latest opaque concurrency token from the handover response",
			Required:    true,
			Type:        "string",
		},
		{
			Name:        "idempotency-key",
			HeaderName:  "Idempotency-Key",
			Description: "Client-generated GUID that remains stable across retries",
			Required:    true,
			Type:        "string",
		},
	}
}
