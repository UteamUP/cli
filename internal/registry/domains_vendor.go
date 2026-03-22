package registry

func init() {
	Register(&Domain{
		Name:        "vendor",
		Aliases:     []string{"vendors"},
		Description: "Manage vendors",
		Actions: append(crudActions("Vendor"),
			Action{Name: "search", Description: "Search vendors", ToolName: "UteamupVendorSearch", Args: queryArg(), Flags: paginationFlags()},
		),
	})

	Register(&Domain{Name: "vendor-portal", Description: "Manage vendor portal", Actions: crudActions("VendorPortal")})
	Register(&Domain{Name: "vendor-performance", Description: "View vendor performance", Actions: listGetActions("VendorPerformance")})
	Register(&Domain{Name: "vendor-analytics", Description: "View vendor analytics", Actions: listGetActions("VendorAnalytics")})
	Register(&Domain{Name: "vendor-compliance", Description: "Manage vendor compliance", Actions: crudActions("VendorCompliance")})
	Register(&Domain{Name: "vendor-match", Description: "Find matching vendors", Actions: listGetActions("VendorMatch")})
	Register(&Domain{Name: "vendor-message", Description: "Manage vendor messages", Actions: crudActions("VendorMessage")})
	Register(&Domain{Name: "vendor-rating", Description: "Manage vendor ratings", Actions: crudActions("VendorRating")})
	Register(&Domain{Name: "vendor-scorecard", Description: "View vendor scorecards", Actions: listGetActions("VendorScorecard")})
}
