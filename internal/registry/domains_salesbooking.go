package registry

func init() {
	// Sales Booking domain
	Register(&Domain{
		Name:        "salesbooking",
		Aliases:     []string{"booking", "sales-booking"},
		Description: "Manage sales bookings (demos, contact sales, support calls, onboarding)",
		Actions: []Action{
			{
				Name:        "slots",
				Description: "Get available booking slots for a date",
				ToolName:    "UteamupSalesBookingGetAvailableSlots",
				Flags: []FlagDef{
					{Name: "date", Short: "d", Description: "Date to check (YYYY-MM-DD)", Required: true, Type: "string"},
					{Name: "type", Short: "t", Description: "Booking type (ContactSales, RequestDemo, SupportCall, Onboarding)", Default: "RequestDemo", Type: "string"},
				},
			},
			{
				Name:        "create",
				Description: "Create a new sales booking",
				ToolName:    "UteamupSalesBookingCreate",
				Flags: []FlagDef{
					{Name: "type", Short: "t", Description: "Booking type (ContactSales, RequestDemo, SupportCall, Onboarding)", Required: true, Type: "string"},
					{Name: "start", Description: "Scheduled start time (UTC ISO 8601)", Required: true, Type: "string"},
					{Name: "name", Description: "Guest name", Required: true, Type: "string"},
					{Name: "email", Description: "Guest email", Required: true, Type: "string"},
					{Name: "language", Description: "Guest language (en, is, pl, de, es)", Default: "en", Type: "string"},
					{Name: "message", Description: "Custom message", Type: "string"},
					{Name: "company", Description: "Company name", Type: "string"},
					{Name: "plan-id", Description: "Plan ID (for ContactSales)", Type: "int"},
					{Name: "from-json", Description: "JSON file with booking data", Type: "string"},
				},
			},
			{
				Name:        "verify",
				Description: "Verify a booking via email token",
				ToolName:    "UteamupSalesBookingVerify",
				Args:        []ArgDef{{Name: "id", Description: "Booking ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "token", Short: "t", Description: "Verification token", Required: true, Type: "string"},
				},
			},
			{
				Name:        "list",
				Description: "List sales bookings by date range",
				ToolName:    "UteamupSalesBookingList",
				Flags: []FlagDef{
					{Name: "start", Short: "s", Description: "Start date (UTC ISO 8601)", Required: true, Type: "string"},
					{Name: "end", Short: "e", Description: "End date (UTC ISO 8601)", Required: true, Type: "string"},
					{Name: "type", Short: "t", Description: "Filter by booking type", Type: "string"},
					{Name: "status", Description: "Filter by status (Pending, Confirmed, Cancelled, Completed, NoShow, Rescheduled, PendingVerification)", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get a sales booking by ID",
				ToolName:    "UteamupSalesBookingGet",
				Args:        []ArgDef{{Name: "id", Description: "Booking ID", Required: true, Type: "int"}},
			},
			{
				Name:        "update-status",
				Description: "Update booking status (admin)",
				ToolName:    "UteamupSalesBookingUpdateStatus",
				Args:        []ArgDef{{Name: "id", Description: "Booking ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "status", Short: "s", Description: "New status (Pending, Confirmed, Cancelled, Completed, NoShow, Rescheduled)", Required: true, Type: "string"},
					{Name: "reason", Short: "r", Description: "Reason for status change", Type: "string"},
				},
			},
			{
				Name:        "cancel",
				Description: "Cancel a sales booking",
				ToolName:    "UteamupSalesBookingCancel",
				Args:        []ArgDef{{Name: "id", Description: "Booking ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "reason", Short: "r", Description: "Cancellation reason", Type: "string"},
				},
			},
		},
	})

	// Sales Availability domain
	Register(&Domain{
		Name:        "salesavailability",
		Aliases:     []string{"availability", "sales-availability"},
		Description: "Manage sales availability schedules and overrides",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get the availability schedule",
				ToolName:    "UteamupSalesAvailabilityGet",
				Flags: []FlagDef{
					{Name: "type", Short: "t", Description: "Filter by booking type (ContactSales, RequestDemo, SupportCall, Onboarding)", Type: "string"},
				},
			},
			{
				Name:        "update",
				Description: "Create or update an availability window",
				ToolName:    "UteamupSalesAvailabilityUpdate",
				Flags: []FlagDef{
					{Name: "id", Description: "Existing availability ID (for update, omit for create)", Type: "int"},
					{Name: "day", Short: "d", Description: "Day of week (Sunday=0..Saturday=6)", Required: true, Type: "int"},
					{Name: "start", Short: "s", Description: "Start time UTC (HH:mm)", Required: true, Type: "string"},
					{Name: "end", Short: "e", Description: "End time UTC (HH:mm)", Required: true, Type: "string"},
					{Name: "duration", Description: "Slot duration in minutes", Default: 30, Type: "int"},
					{Name: "buffer", Description: "Buffer between slots in minutes", Default: 15, Type: "int"},
					{Name: "type", Short: "t", Description: "Booking type", Required: true, Type: "string"},
					{Name: "active", Description: "Whether this window is active", Default: true, Type: "bool"},
					{Name: "from-json", Description: "JSON file with availability data", Type: "string"},
				},
			},
			{
				Name:        "create-override",
				Description: "Create an availability override (holiday/blocked day)",
				ToolName:    "UteamupSalesAvailabilityCreateOverride",
				Flags: []FlagDef{
					{Name: "date", Short: "d", Description: "Override date (YYYY-MM-DD)", Required: true, Type: "string"},
					{Name: "blocked", Short: "b", Description: "Whether the day is fully blocked", Default: true, Type: "bool"},
					{Name: "start", Short: "s", Description: "Override start time UTC (HH:mm, for partial block)", Type: "string"},
					{Name: "end", Short: "e", Description: "Override end time UTC (HH:mm, for partial block)", Type: "string"},
					{Name: "reason", Short: "r", Description: "Reason for override", Type: "string"},
				},
			},
			{
				Name:        "delete-override",
				Description: "Delete an availability override",
				ToolName:    "UteamupSalesAvailabilityDeleteOverride",
				Args:        []ArgDef{{Name: "id", Description: "Override ID", Required: true, Type: "int"}},
			},
		},
	})
}
