package registry

func resourceReservationGUIDArgument() []ArgDef {
	return []ArgDef{{Name: "reservationGuid", Description: "Reservation public GUID", Required: true, Type: "uuid"}}
}

func resourceReservationDecisionAction(name, description, step string, withNote bool) Action {
	action := Action{
		Name:        name,
		Description: description,
		ToolName:    "UteamupResourceReservationDecide",
		HTTPMethod:  "POST",
		RESTPath:    "{reservationGuid}/" + step,
		Args:        resourceReservationGUIDArgument(),
	}
	if withNote {
		action.Flags = []FlagDef{
			{Name: "note", BodyName: "note", Description: "Decision note shown to the requester, maximum 1000 characters", Type: "string"},
		}
	}
	return action
}

func init() {
	Register(&Domain{
		Name:        "resource-reservation",
		Aliases:     []string{"resource-reservations"},
		Description: "Reserve technicians, tools, equipment, vehicles and facilities and decide reservation requests",
		APIPath:     "/api/resourcereservations",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List reservations visible to the caller, newest window first",
				ToolName:    "UteamupResourceReservationList",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "statuses", QueryName: "statuses", Description: "Statuses to include: pendingVerification, requested, confirmed, checkedOut, returned, cancelled, rejected, expired (repeatable)", Type: "stringSlice"},
					{Name: "from-utc", QueryName: "fromUtc", Description: "Only reservations ending after this ISO 8601 UTC time", Type: "string"},
					{Name: "to-utc", QueryName: "toUtc", Description: "Only reservations starting before this ISO 8601 UTC time", Type: "string"},
					{Name: "resource-guid", QueryName: "resourceGuid", Description: "Bookable-resource public GUID", Type: "uuid"},
					{Name: "mine", QueryName: "mine", Description: "Only the caller's own reservations", Type: "bool"},
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one reservation",
				ToolName:    "UteamupResourceReservationGet",
				HTTPMethod:  "GET",
				RESTPath:    "{reservationGuid}",
				Args:        resourceReservationGUIDArgument(),
			},
			{
				Name:        "create",
				Description: "Reserve one resource, named by its source GUID or its bookable-resource GUID",
				ToolName:    "UteamupResourceReservationCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "resource-type", BodyName: "resourceType", Description: "0=technician, 3=equipment, 4=vehicle, 5=facility, 7=tool (pools, contractors and crews cannot be reserved directly)", Required: true, Type: "int"},
					{Name: "source-guid", BodyName: "sourceGuid", Description: "Source public GUID (user, asset, tool or location); exactly one of this or --resource-guid", Type: "uuid"},
					{Name: "resource-guid", BodyName: "resourceGuid", Description: "Bookable-resource public GUID; exactly one of this or --source-guid", Type: "uuid"},
					{Name: "start-utc", BodyName: "startUtc", Description: "Reservation start, ISO 8601 UTC", Required: true, Type: "string"},
					{Name: "end-utc", BodyName: "endUtc", Description: "Reservation end, ISO 8601 UTC (at most 31 days after --start-utc)", Required: true, Type: "string"},
					{Name: "capacity-requested", BodyName: "capacityRequested", Description: "Capacity to reserve, greater than zero", Default: 1.0, Type: "float"},
					{Name: "purpose", BodyName: "purpose", Description: "Why the resource is needed, maximum 500 characters", Type: "string"},
					{Name: "notes", BodyName: "notes", Description: "Additional notes, maximum 2000 characters", Type: "string"},
					{Name: "idempotency-key", HeaderName: "Idempotency-Key", Required: true, Type: "uuid",
						Description: "Stable request GUID; reuse it with the same reservation after an uncertain response"},
				},
			},
			resourceReservationDecisionAction("approve", "Approve a requested reservation", "approve", true),
			resourceReservationDecisionAction("reject", "Reject a requested reservation", "reject", true),
			resourceReservationDecisionAction("cancel", "Cancel a reservation that has not been checked out", "cancel", true),
			resourceReservationDecisionAction("checkout", "Record that a confirmed tool, equipment or vehicle was handed out", "checkout", false),
			resourceReservationDecisionAction("return", "Record that a checked-out tool, equipment or vehicle came back", "return", false),
		},
	})
}
