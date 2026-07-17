package registry

func init() {
	Register(&Domain{
		Name:        "asset-lifecycle",
		Description: "Manage asset lifecycle events",
		Actions: append(crudActions("AssetLifecycleEvent"),
			Action{
				Name:        "by-type",
				Description: "List lifecycle events by event type",
				ToolName:    "UteamupAssetlifecycleGetByType",
				HTTPMethod:  "GET",
				RESTPath:    "by-type",
				Flags: []FlagDef{
					{Name: "event-type", BodyName: "eventType", Description: "Lifecycle event type", Type: "string", Required: true},
				},
			},
			Action{
				Name:        "by-date",
				Description: "List lifecycle events within a date range",
				ToolName:    "UteamupAssetlifecycleGetByDateRange",
				HTTPMethod:  "GET",
				RESTPath:    "by-date-range",
				Flags: []FlagDef{
					{Name: "start-date", BodyName: "startDate", Description: "Inclusive ISO-8601 start date", Type: "string", Required: true},
					{Name: "end-date", BodyName: "endDate", Description: "Inclusive ISO-8601 end date", Type: "string", Required: true},
				},
			},
		),
	})
	Register(&Domain{
		Name:        "asset-rental",
		Description: "Manage asset rentals",
		Actions: append(crudActions("AssetRental"),
			Action{
				Name:        "available",
				Description: "List rental-configured assets currently available",
				ToolName:    "UteamupAssetrentalGetAvailable",
				HTTPMethod:  "GET",
				RESTPath:    "available",
			},
			Action{
				Name:        "active",
				Description: "List rental-configured assets currently rented",
				ToolName:    "UteamupAssetrentalGetRented",
				HTTPMethod:  "GET",
				RESTPath:    "rented",
			},
			Action{
				Name:        "expiring",
				Description: "List rentals due back within a planning window",
				ToolName:    "UteamupAssetrentalGetExpiringSoon",
				HTTPMethod:  "GET",
				RESTPath:    "expiring-soon",
				Flags: []FlagDef{
					{Name: "days", Description: "Planning window in days (1-365)", Type: "int", Default: 30},
				},
			},
			Action{
				Name:        "revenue",
				Description: "Read rental revenue for a bounded date range",
				ToolName:    "UteamupAssetrentalRevenueSummary",
				HTTPMethod:  "GET",
				RESTPath:    "revenue-summary",
				Flags: []FlagDef{
					{Name: "start-date", BodyName: "startDate", Description: "Inclusive ISO-8601 start date", Type: "string", Required: true},
					{Name: "end-date", BodyName: "endDate", Description: "Inclusive ISO-8601 end date", Type: "string", Required: true},
				},
			},
		),
	})
	Register(&Domain{Name: "asset-replacement-plan", Description: "Manage asset replacement plans", Actions: crudActions("AssetReplacementPlan")})
	Register(&Domain{Name: "asset-scan-log", Description: "View asset scan logs", Actions: listGetActions("AssetScanLog")})
	Register(&Domain{
		Name:        "asset-booking",
		Aliases:     []string{"asset-calendar"},
		Description: "Manage GUID-first asset calendar bookings",
		Actions: []Action{
			{
				Name: "list", Description: "List bookings for an asset", ToolName: "UteamupAssetCalendarBookingList",
				Args: []ArgDef{{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "from", Description: "Optional UTC period start", Type: "string"},
					{Name: "to", Description: "Optional UTC period end", Type: "string"},
				},
			},
			{
				Name: "conflicts", Description: "Check an asset booking window", ToolName: "UteamupAssetCalendarBookingGetConflicts",
				Args: []ArgDef{
					{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "string"},
					{Name: "start", Description: "Proposed UTC start", Required: true, Type: "string"},
					{Name: "end", Description: "Proposed UTC end", Required: true, Type: "string"},
				},
				Flags: []FlagDef{{Name: "exclude-booking-guid", BodyName: "excludeBookingGuid", Description: "Public booking GUID to exclude", Type: "string"}},
			},
			{
				Name: "create", Description: "Create an asset booking", ToolName: "UteamupAssetCalendarBookingCreate",
				Args:  []ArgDef{{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{jsonFlag()},
			},
			{
				Name: "delete", Description: "Delete an asset booking", ToolName: "UteamupAssetCalendarBookingDelete",
				Args: []ArgDef{
					{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "string"},
					{Name: "bookingGuid", Description: "Public booking GUID", Required: true, Type: "string"},
				},
			},
		},
	})
	Register(&Domain{Name: "asset-condition", Description: "Manage asset condition assessments", Actions: crudActions("AssetConditionAssessment")})
	Register(&Domain{Name: "asset-criticality", Description: "Manage asset criticality assessments", Actions: crudActions("AssetCriticalityAssessment")})
}
