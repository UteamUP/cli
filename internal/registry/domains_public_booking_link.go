package registry

func publicBookingLinkFlags() []FlagDef {
	return []FlagDef{
		{Name: "name", BodyName: "name", Description: "Page name, 2-200 characters", Required: true, Type: "string"},
		{Name: "description", BodyName: "description", Description: "Text shown to visitors, maximum 1000 characters", Type: "string"},
		{Name: "is-active", BodyName: "isActive", Description: "Whether the public page accepts bookings", Default: true, Type: "bool"},
		{Name: "allowed-resource-types", BodyName: "allowedResourceTypes", Description: "Bookable types: technician, tool, equipment, vehicle (repeatable, at least one)", Required: true, Type: "stringSlice"},
		{Name: "requires-approval", BodyName: "requiresApproval", Description: "Hold verified bookings for an approver", Default: true, Type: "bool"},
		{Name: "minimum-lead-time-hours", BodyName: "minimumLeadTimeHours", Description: "Earliest booking start from now, 0-720 hours", Default: 2, Type: "int"},
		{Name: "booking-horizon-days", BodyName: "bookingHorizonDays", Description: "Latest bookable day from now, 1-365", Default: 30, Type: "int"},
		{Name: "max-duration-minutes", BodyName: "maxDurationMinutes", Description: "Longest booking, 15-10080 minutes and at least one slot", Default: 480, Type: "int"},
		{Name: "slot-minutes", BodyName: "slotMinutes", Description: "Slot length: 15, 30 or 60", Default: 30, Type: "int"},
		{Name: "approval-deadline-hours", BodyName: "approvalDeadlineHours", Description: "Hours an approver has before a request expires, 1-336", Default: 48, Type: "int"},
		{Name: "max-submissions-per-hour", BodyName: "maxSubmissionsPerHour", Description: "Public submissions accepted per hour, 1-500", Default: 20, Type: "int"},
	}
}

func init() {
	Register(&Domain{
		Name:        "public-booking-link",
		Aliases:     []string{"public-booking-links"},
		Description: "Manage public booking pages that let people outside the tenant request resources",
		APIPath:     "/api/publicbookinglinks",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the tenant's public booking pages",
				ToolName:    "UteamupPublicBookingLinkList",
				HTTPMethod:  "GET",
			},
			{
				Name:        "create",
				Description: "Create a public booking page",
				ToolName:    "UteamupPublicBookingLinkCreate",
				HTTPMethod:  "POST",
				Flags:       publicBookingLinkFlags(),
			},
			{
				Name:        "update",
				Description: "Replace a public booking page's settings",
				ToolName:    "UteamupPublicBookingLinkUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{linkGuid}",
				Args: []ArgDef{
					{Name: "linkGuid", Description: "Public booking page GUID", Required: true, Type: "uuid"},
				},
				Flags: publicBookingLinkFlags(),
			},
		},
	})
}
