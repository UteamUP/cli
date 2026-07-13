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
			Name:        "operational-baton",
			Description: "Show the deterministic action-first handover baton (zero AI credits)",
			ToolName:    "UteamupShiftHandoverGetOperationalBaton",
			HTTPMethod:  "GET",
			RESTPath:    "by-guid/{handoverGuid}/operational-baton",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
		},
		Action{
			Name:        "history",
			Description: "Verify and show immutable handover versions, events, and signatures",
			ToolName:    "UteamupShiftHandoverGetHistory",
			HTTPMethod:  "GET",
			RESTPath:    "by-guid/{handoverGuid}/history",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
		},
		Action{
			Name:        "signature-create",
			Description: "Add idempotent manager or compliance evidence to the latest immutable version",
			ToolName:    "UteamupShiftHandoverCreateSignature",
			HTTPMethod:  "POST",
			RESTPath:    "by-guid/{handoverGuid}/signatures",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: []FlagDef{
				{Name: "purpose", Description: "managerApproval or complianceAttestation", Required: true, Type: "string"},
				{Name: "method", Description: "Authenticated or device attestation method", Required: true, Type: "string"},
				{Name: "meaning", Description: "Explicit signer meaning (10-1000 characters)", Required: true, Type: "string"},
				{Name: "device-key-identifier", BodyName: "deviceKeyIdentifier", Description: "Optional managed-device key identifier", Type: "string"},
				{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Client-generated GUID stable across retries", Required: true, Type: "uuid"},
			},
		},
		Action{
			Name:        "audit-export",
			Description: "Return a portable audit manifest with verified history and manifest hash",
			ToolName:    "UteamupShiftHandoverExportAudit",
			HTTPMethod:  "GET",
			RESTPath:    "by-guid/{handoverGuid}/audit-export",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
		},
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
			Name:        "carryovers",
			Description: "List carry-over items by stable handover GUID",
			ToolName:    "UteamupShiftHandoverGetCarryOvers",
			HTTPMethod:  "GET",
			RESTPath:    "by-guid/{handoverGuid}/carryovers",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
		},
		Action{
			Name:        "carryover-create",
			Description: "Add a carry-over item to a draft or rejected handover",
			ToolName:    "UteamupShiftHandoverCreateCarryOver",
			HTTPMethod:  "POST",
			RESTPath:    "by-guid/{handoverGuid}/carryovers",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: []FlagDef{
				{Name: "description", Required: true, Type: "string", Description: "Operational issue that must continue to the next shift"},
				{Name: "priority", Type: "int", Description: "Priority from 0 (highest) to 5 (lowest)"},
				{Name: "original-handover-guid", BodyName: "originalShiftHandoverGuid", Type: "uuid", Description: "Optional origin handover GUID for migrated issues"},
				handoverConcurrencyFlag(),
			},
		},
		Action{
			Name:        "carryover-update",
			Description: "Update a carry-over item by stable GUID",
			ToolName:    "UteamupShiftHandoverUpdateCarryOver",
			HTTPMethod:  "PUT",
			RESTPath:    "by-guid/{handoverGuid}/carryovers/{carryOverGuid}",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "carryOverGuid", Description: "Carry-over ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: []FlagDef{
				{Name: "description", Type: "string", Description: "Updated operational issue"},
				{Name: "status", Type: "string", Description: "active, resolved, or escalated; conversion uses carryover-convert"},
				{Name: "priority", Type: "int", Description: "Priority from 0 (highest) to 5 (lowest)"},
				handoverConcurrencyFlag(),
			},
		},
		Action{
			Name:        "carryover-delete",
			Description: "Delete an unconverted carry-over item by stable GUID",
			ToolName:    "UteamupShiftHandoverDeleteCarryOver",
			HTTPMethod:  "DELETE",
			RESTPath:    "by-guid/{handoverGuid}/carryovers/{carryOverGuid}",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "carryOverGuid", Description: "Carry-over ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: []FlagDef{handoverConcurrencyFlag()},
		},
		Action{
			Name:        "carryover-convert",
			Description: "Convert a carry-over into one traceable workorder exactly once",
			ToolName:    "UteamupShiftHandoverConvertCarryOverToWorkOrder",
			HTTPMethod:  "POST",
			RESTPath:    "by-guid/{handoverGuid}/carryovers/{carryOverGuid}/convert-to-workorder",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "carryOverGuid", Description: "Carry-over ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: handoverMutationFlags(),
		},
		Action{
			Name:        "carryover-links",
			Description: "List operational records linked to a carry-over item",
			ToolName:    "UteamupShiftHandoverGetCarryOverLinks",
			HTTPMethod:  "GET",
			RESTPath:    "by-guid/{handoverGuid}/carryovers/{carryOverGuid}/links",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "carryOverGuid", Description: "Carry-over ExternalGuid", Required: true, Type: "uuid"},
			},
		},
		Action{
			Name:        "carryover-link-create",
			Description: "Link an operational record to a carry-over without changing the source record",
			ToolName:    "UteamupShiftHandoverCreateCarryOverLink",
			HTTPMethod:  "POST",
			RESTPath:    "by-guid/{handoverGuid}/carryovers/{carryOverGuid}/links",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "carryOverGuid", Description: "Carry-over ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: handoverItemLinkCreateFlags(),
		},
		Action{
			Name:        "carryover-link-delete",
			Description: "Remove a carry-over operational link without changing the source record",
			ToolName:    "UteamupShiftHandoverDeleteCarryOverLink",
			HTTPMethod:  "DELETE",
			RESTPath:    "by-guid/{handoverGuid}/carryovers/{carryOverGuid}/links/{linkGuid}",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "carryOverGuid", Description: "Carry-over ExternalGuid", Required: true, Type: "uuid"},
				{Name: "linkGuid", Description: "Link ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: []FlagDef{handoverConcurrencyFlag()},
		},
		Action{
			Name:        "carryover-critical-acknowledge",
			Description: "Acknowledge an unresolved P0/P1 carry-over without resolving it",
			ToolName:    "UteamupShiftHandoverAcknowledgeCriticalCarryOver",
			HTTPMethod:  "PUT",
			RESTPath:    "by-guid/{handoverGuid}/carryovers/{carryOverGuid}/acknowledge-critical",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "carryOverGuid", Description: "Carry-over ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: criticalItemAcknowledgmentFlags(),
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
		Action{
			Name:        "section-links",
			Description: "List operational records linked to a handover section",
			ToolName:    "UteamupShiftHandoverGetSectionLinks",
			HTTPMethod:  "GET",
			RESTPath:    "by-guid/{handoverGuid}/sections/{sectionGuid}/links",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "sectionGuid", Description: "Section ExternalGuid", Required: true, Type: "uuid"},
			},
		},
		Action{
			Name:        "section-link-create",
			Description: "Link an operational record to a section without changing the source record",
			ToolName:    "UteamupShiftHandoverCreateSectionLink",
			HTTPMethod:  "POST",
			RESTPath:    "by-guid/{handoverGuid}/sections/{sectionGuid}/links",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "sectionGuid", Description: "Section ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: handoverItemLinkCreateFlags(),
		},
		Action{
			Name:        "section-link-delete",
			Description: "Remove a section operational link without changing the source record",
			ToolName:    "UteamupShiftHandoverDeleteSectionLink",
			HTTPMethod:  "DELETE",
			RESTPath:    "by-guid/{handoverGuid}/sections/{sectionGuid}/links/{linkGuid}",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "sectionGuid", Description: "Section ExternalGuid", Required: true, Type: "uuid"},
				{Name: "linkGuid", Description: "Link ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: []FlagDef{handoverConcurrencyFlag()},
		},
		Action{
			Name:        "section-critical-acknowledge",
			Description: "Acknowledge a completed critical section as the incoming operator",
			ToolName:    "UteamupShiftHandoverAcknowledgeCriticalSection",
			HTTPMethod:  "PUT",
			RESTPath:    "by-guid/{handoverGuid}/sections/{sectionGuid}/acknowledge-critical",
			Args: []ArgDef{
				{Name: "handoverGuid", Description: "Shift handover ExternalGuid", Required: true, Type: "uuid"},
				{Name: "sectionGuid", Description: "Section ExternalGuid", Required: true, Type: "uuid"},
			},
			Flags: criticalItemAcknowledgmentFlags(),
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

func handoverItemLinkCreateFlags() []FlagDef {
	return []FlagDef{
		{
			Name:        "linked-entity-type",
			BodyName:    "linkedEntityType",
			Description: "project, workOrder, asset, location, workPermit, incident, journal, or document",
			Required:    true,
			Type:        "string",
		},
		{
			Name:        "linked-entity-guid",
			BodyName:    "linkedEntityGuid",
			Description: "ExternalGuid of the operational record to link",
			Required:    true,
			Type:        "uuid",
		},
		handoverConcurrencyFlag(),
	}
}

func criticalItemAcknowledgmentFlags() []FlagDef {
	return []FlagDef{
		{
			Name:        "notes",
			Description: "Optional acknowledgement notes",
			Type:        "string",
		},
		handoverConcurrencyFlag(),
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
