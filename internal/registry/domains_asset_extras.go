package registry

func init() {
	Register(&Domain{Name: "asset-lifecycle", Description: "Manage asset lifecycle events", Actions: crudActions("AssetLifecycleEvent")})
	Register(&Domain{Name: "asset-maintenance-plan", Aliases: []string{"amp"}, Description: "Manage asset maintenance plans", Actions: crudActions("AssetMaintenancePlan")})
	Register(&Domain{Name: "asset-rental", Description: "Manage asset rentals", Actions: crudActions("AssetRental")})
	Register(&Domain{Name: "asset-replacement-plan", Description: "Manage asset replacement plans", Actions: crudActions("AssetReplacementPlan")})
	Register(&Domain{Name: "asset-scan-log", Description: "View asset scan logs", Actions: listGetActions("AssetScanLog")})
	Register(&Domain{Name: "asset-booking", Aliases: []string{"asset-calendar"}, Description: "Manage asset calendar bookings", Actions: crudActions("AssetCalendarBooking")})
	Register(&Domain{Name: "asset-condition", Description: "Manage asset condition assessments", Actions: crudActions("AssetConditionAssessment")})
	Register(&Domain{Name: "asset-criticality", Description: "Manage asset criticality assessments", Actions: crudActions("AssetCriticalityAssessment")})
}
