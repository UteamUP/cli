package registry

func init() {
	Register(&Domain{Name: "asset-lifecycle", Description: "Manage asset lifecycle events", Actions: crudActions("AssetLifecycleEvent")})
	Register(&Domain{Name: "asset-rental", Description: "Manage asset rentals", Actions: crudActions("AssetRental")})
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
