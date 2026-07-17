package registry

func init() {
	Register(&Domain{
		Name:        "vendor",
		Aliases:     []string{"vendors"},
		Description: "Manage vendors",
		Actions: append(crudActions("Vendor"),
			Action{Name: "search", Description: "Search vendors", ToolName: "UteamupVendorSearch", Args: queryArg(), Flags: paginationFlags()},
			// --- Reseller catalog: vendor's part catalog (stock-reseller-catalog §6) ---
			Action{
				Name:        "catalog",
				Description: "List the part-catalog entries a vendor supplies (vendor part numbers, costs, MOQ, lead times)",
				ToolName:    "UteamupVendorGetCatalog",
				RESTPath:    "by-guid/{guid}/catalog",
				Args:        []ArgDef{{Name: "guid", Description: "Vendor GUID", Required: true, Type: "string"}},
			},
		),
	})

	Register(&Domain{Name: "vendor-portal", Description: "Manage vendor portal", Actions: crudActions("VendorPortal")})
	Register(&Domain{
		Name:        "vendor-performance",
		Aliases:     []string{"vendor-scorecard"},
		Description: "View GUID-first vendor scorecards, trends, events, rankings, and configuration",
		APIPath:     "/api/v1/vendorperformance",
		Actions: []Action{
			{
				Name:        "scorecard",
				Description: "Get a vendor scorecard by public GUID",
				ToolName:    "UteamupVendorScorecardGet",
				RESTPath:    "by-guid/{vendorGuid}/scorecard",
				HTTPMethod:  "GET",
				Args:        []ArgDef{{Name: "vendorGuid", Description: "Vendor GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "events",
				Description: "List vendor performance events by public GUID",
				RESTPath:    "by-guid/{vendorGuid}/events",
				HTTPMethod:  "GET",
				Args:        []ArgDef{{Name: "vendorGuid", Description: "Vendor GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "from", Description: "Optional ISO start date", Type: "string"},
					{Name: "to", Description: "Optional ISO end date", Type: "string"},
				},
			},
			{
				Name:        "trends",
				Description: "Get vendor performance trend snapshots by public GUID",
				RESTPath:    "by-guid/{vendorGuid}/trends",
				HTTPMethod:  "GET",
				Args:        []ArgDef{{Name: "vendorGuid", Description: "Vendor GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "period", Description: "Snapshot period", Type: "string"},
					{Name: "from", Description: "Optional ISO start date", Type: "string"},
					{Name: "to", Description: "Optional ISO end date", Type: "string"},
				},
			},
			{
				Name:        "rankings",
				Description: "List tenant vendor performance rankings",
				RESTPath:    "rankings",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "sort-by", Description: "overall, speed, quality, price, or engagement", Type: "string"},
					{Name: "page", Description: "Page number", Type: "int", Default: 1},
					{Name: "page-size", Description: "Page size", Type: "int", Default: 25},
				},
			},
			{
				Name:        "dashboard",
				Description: "Get the tenant vendor-performance dashboard",
				RESTPath:    "dashboard",
				HTTPMethod:  "GET",
			},
			{
				Name:        "recalculate",
				Description: "Recalculate a vendor scorecard by public GUID",
				RESTPath:    "by-guid/{vendorGuid}/recalculate",
				HTTPMethod:  "POST",
				Args:        []ArgDef{{Name: "vendorGuid", Description: "Vendor GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:         "config",
				Description:  "Get tenant vendor-scoring configuration",
				ToolName:     "UteamupVendorScorecardConfigGet",
				RESTBasePath: "/api/v1/scoringconfiguration",
				HTTPMethod:   "GET",
			},
		},
	})
	Register(&Domain{Name: "vendor-analytics", Description: "View vendor analytics", Actions: listGetActions("VendorAnalytics")})
	Register(&Domain{Name: "vendor-compliance", Description: "Manage vendor compliance", Actions: crudActions("VendorCompliance")})
	Register(&Domain{Name: "vendor-match", Description: "Find matching vendors", Actions: listGetActions("VendorMatch")})
	Register(&Domain{Name: "vendor-message", Description: "Manage vendor messages", Actions: crudActions("VendorMessage")})
	Register(&Domain{Name: "vendor-rating", Description: "Manage vendor ratings", Actions: crudActions("VendorRating")})
}
