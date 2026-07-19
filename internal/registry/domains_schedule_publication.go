package registry

func init() {
	Register(&Domain{
		Name:        "schedule-draft",
		Aliases:     []string{"schedule-drafts"},
		Description: "Plan and publish versioned tenant-local weekly schedules",
		APIPath:     "/api/scheduledraft",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List schedule drafts, optionally for one tenant-local ISO week",
				ToolName:    "UteamupScheduleDraftList",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "week-start", BodyName: "weekStart", Description: "Optional tenant-local ISO week start", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Read one schedule draft and its entries by public GUID",
				ToolName:    "UteamupScheduleDraftGet",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{draftGuid}",
				Args: []ArgDef{
					{Name: "draftGuid", Description: "Schedule draft GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create an explicit weekly draft, optionally cloned from a publication",
				ToolName:    "UteamupScheduleDraftCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "week-start", BodyName: "weekStart", Description: "Tenant-local ISO week start", Required: true, Type: "string"},
					{Name: "base-publication-guid", BodyName: "basePublicationGuid", Description: "Optional publication GUID to clone", Type: "string"},
				},
			},
			{
				Name:        "entry-add",
				Description: "Add one shift or work-order entry to an editable draft",
				ToolName:    "UteamupScheduleDraftEntryAdd",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{draftGuid}/entries",
				Args: []ArgDef{
					{Name: "draftGuid", Description: "Schedule draft GUID", Required: true, Type: "uuid"},
				},
				Flags: scheduleDraftEntryFlags(false),
			},
			{
				Name:        "entry-update",
				Description: "Update one schedule draft entry using optimistic concurrency",
				ToolName:    "UteamupScheduleDraftEntryUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{draftGuid}/entries/{entryGuid}",
				Args: []ArgDef{
					{Name: "draftGuid", Description: "Schedule draft GUID", Required: true, Type: "uuid"},
					{Name: "entryGuid", Description: "Schedule draft entry GUID", Required: true, Type: "uuid"},
				},
				Flags: scheduleDraftEntryFlags(true),
			},
			{
				Name:        "entry-delete",
				Description: "Delete one schedule draft entry using optimistic concurrency",
				ToolName:    "UteamupScheduleDraftEntryDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{draftGuid}/entries/{entryGuid}",
				Args: []ArgDef{
					{Name: "draftGuid", Description: "Schedule draft GUID", Required: true, Type: "uuid"},
					{Name: "entryGuid", Description: "Schedule draft entry GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "expected-row-version", BodyName: "expectedRowVersion", Description: "Latest draft entry concurrency token", Required: true, Type: "int"},
				},
			},
			{
				Name:        "validate",
				Description: "Validate a draft against qualifications, availability, policy, and capacity",
				ToolName:    "UteamupScheduleDraftValidate",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{draftGuid}/validate",
				Args: []ArgDef{
					{Name: "draftGuid", Description: "Schedule draft GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "publish",
				Description: "Publish a validated draft transactionally",
				ToolName:    "UteamupScheduleDraftPublish",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{draftGuid}/publish",
				Args: []ArgDef{
					{Name: "draftGuid", Description: "Schedule draft GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "expected-row-version", BodyName: "expectedRowVersion", Description: "Latest draft concurrency token", Required: true, Type: "int"},
					{Name: "acknowledged-warning-code", BodyName: "acknowledgedWarningCodes", Description: "Overrideable warning codes explicitly acknowledged", Type: "stringSlice"},
					{Name: "override-reason", BodyName: "overrideReason", Description: "Required reason when acknowledging warnings", Type: "string"},
				},
			},
		},
	})

	Register(&Domain{
		Name:        "schedule-publication",
		Aliases:     []string{"schedule-publications"},
		Description: "Inspect, diff, recall, and roll back immutable schedule publications",
		APIPath:     "/api/schedulepublication",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Read one immutable schedule publication by public GUID",
				ToolName:    "UteamupSchedulePublicationGet",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{publicationGuid}",
				Args: []ArgDef{
					{Name: "publicationGuid", Description: "Schedule publication GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "history",
				Description: "List publication versions for one tenant-local ISO week",
				ToolName:    "UteamupSchedulePublicationHistory",
				HTTPMethod:  "GET",
				RESTPath:    "history",
				Flags: []FlagDef{
					{Name: "week-start", BodyName: "weekStart", Description: "Tenant-local ISO week start", Required: true, Type: "string"},
				},
			},
			{
				Name:        "diff",
				Description: "Diff two publication versions by public GUID",
				ToolName:    "UteamupSchedulePublicationDiff",
				HTTPMethod:  "GET",
				RESTPath:    "diff/{fromPublicationGuid}/{toPublicationGuid}",
				Args: []ArgDef{
					{Name: "fromPublicationGuid", Description: "Older schedule publication GUID", Required: true, Type: "uuid"},
					{Name: "toPublicationGuid", Description: "Newer schedule publication GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "recall",
				Description: "Recall future, not-started work with an audit reason",
				ToolName:    "UteamupSchedulePublicationRecall",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{publicationGuid}/recall",
				Args: []ArgDef{
					{Name: "publicationGuid", Description: "Schedule publication GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "reason", Description: "Required audit reason", Required: true, Type: "string"},
				},
			},
			{
				Name:        "rollback",
				Description: "Create a new monotonic version from an older publication snapshot",
				ToolName:    "UteamupSchedulePublicationRollback",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{publicationGuid}/rollback",
				Args: []ArgDef{
					{Name: "publicationGuid", Description: "Schedule publication GUID to restore", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "reason", Description: "Required audit reason", Required: true, Type: "string"},
				},
			},
		},
	})

	Register(&Domain{
		Name:        "workforce-capacity",
		Aliases:     []string{"workforce-reservation"},
		Description: "Check worker availability and reserve partial project capacity",
		APIPath:     "/api/workforcecapacity",
		Actions: []Action{
			{
				Name:        "availability",
				Description: "Resolve all availability effects for one tenant worker",
				ToolName:    "UteamupWorkforceAvailabilityCheck",
				HTTPMethod:  "GET",
				RESTPath:    "availability/{workerGuid}",
				Args: []ArgDef{
					{Name: "workerGuid", Description: "Tenant worker GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "from-utc", BodyName: "fromUtc", Description: "Inclusive availability window start in UTC", Required: true, Type: "string"},
					{Name: "to-utc", BodyName: "toUtc", Description: "Exclusive availability window end in UTC", Required: true, Type: "string"},
				},
			},
			{
				Name:        "reservation-readiness",
				Description: "Check partial project capacity without persisting a reservation",
				ToolName:    "UteamupWorkforceReservationReadiness",
				HTTPMethod:  "POST",
				RESTPath:    "reservation-readiness",
				Flags:       workforceReservationFlags(false),
			},
			{
				Name:        "reservation-create",
				Description: "Revalidate and reserve partial project capacity transactionally",
				ToolName:    "UteamupWorkforceReservationCreate",
				HTTPMethod:  "POST",
				RESTPath:    "reservations",
				Flags:       workforceReservationFlags(true),
			},
		},
	})
}

func scheduleDraftEntryFlags(includeConcurrency bool) []FlagDef {
	flags := []FlagDef{
		{Name: "kind", Description: "Entry kind: shift or workorder", Required: true, Type: "string"},
		{Name: "shift-guid", BodyName: "shiftGuid", Description: "Shift GUID for shift entries", Type: "string"},
		{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Workorder GUID for work-order entries", Type: "string"},
		{Name: "workforce-group-guid", BodyName: "workforceGroupGuid", Description: "Optional workforce group GUID", Type: "string"},
		{Name: "location-guid", BodyName: "locationGuid", Description: "Optional location GUID", Type: "string"},
		{Name: "start-utc", BodyName: "startUtc", Description: "Entry start in UTC", Required: true, Type: "string"},
		{Name: "end-utc", BodyName: "endUtc", Description: "Entry end in UTC", Required: true, Type: "string"},
		{Name: "allocation-percent", BodyName: "allocationPercent", Description: "Entry allocation from 0.01 to 100 percent", Default: 100.0, Type: "float"},
		{Name: "open-shift", BodyName: "isOpenShift", Description: "Leave the entry open for eligible worker claims", Type: "bool"},
		{Name: "notes", Description: "Optional planning notes", Type: "string"},
		{Name: "workers-json", BodyName: "workers", Description: "Path to a JSON array of workerGuid/allocationPercent objects", Type: "string", JSONFile: true},
	}
	if includeConcurrency {
		flags = append(flags, FlagDef{
			Name:        "expected-row-version",
			BodyName:    "expectedRowVersion",
			Description: "Latest draft entry concurrency token",
			Required:    true,
			Type:        "int",
		})
	}
	return flags
}

func workforceReservationFlags(includeCreateFields bool) []FlagDef {
	flags := []FlagDef{
		{Name: "worker-guid", BodyName: "workerGuid", Description: "Tenant worker GUID", Required: true, Type: "string"},
		{Name: "project-guid", BodyName: "projectGuid", Description: "Tenant project GUID", Required: true, Type: "string"},
		{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Optional project workorder GUID", Type: "string"},
		{Name: "start-utc", BodyName: "start", Description: "Reservation start in UTC", Required: true, Type: "string"},
		{Name: "end-utc", BodyName: "end", Description: "Reservation end in UTC", Required: true, Type: "string"},
		{Name: "allocation-percent", BodyName: "allocationPercent", Description: "Capacity claimed from 0.01 to 100 percent", Required: true, Type: "float"},
	}
	if includeCreateFields {
		flags = append(
			flags,
			FlagDef{Name: "team-guid", BodyName: "teamGuid", Description: "Optional tenant team GUID", Type: "string"},
			FlagDef{Name: "handoff-notes", BodyName: "handoffNotes", Description: "Optional reservation handoff notes", Type: "string"},
		)
	}
	return flags
}
