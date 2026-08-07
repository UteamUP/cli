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
				ToolName:    "UteamupVendorPerformanceEventsList",
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
				ToolName:    "UteamupVendorPerformanceTrendsGet",
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
				ToolName:    "UteamupVendorPerformanceRankingsList",
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
				ToolName:    "UteamupVendorPerformanceDashboardGet",
				RESTPath:    "dashboard",
				HTTPMethod:  "GET",
			},
			{
				Name:        "recalculate",
				Description: "Recalculate a vendor scorecard by public GUID",
				ToolName:    "UteamupVendorScorecardRecalculate",
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
	Register(&Domain{
		Name:        "vendor-match",
		Description: "Find matching vendors by work order public GUID",
		APIPath:     "/api/v1/vendor/match",
		Actions: []Action{
			{
				Name:        "match",
				Description: "Calculate vendor matches from a GUID-only JSON model",
				ToolName:    "UteamupVendorMatchVendors",
				MCPOnly:     true,
				Flags: []FlagDef{{
					Name:        "from-json",
					BodyName:    "model",
					Description: "JSON file containing workOrderGuid and maxResults",
					Required:    true,
					Type:        "string",
					JSONFile:    true,
				}},
			},
			{
				Name:        "list",
				Description: "List existing matches for a work order public GUID",
				ToolName:    "UteamupVendorMatchGet",
				RESTPath:    "workorder/{workOrderGuid}",
				HTTPMethod:  "GET",
				Args: []ArgDef{{
					Name: "workOrderGuid", Description: "Work order public GUID", Required: true, Type: "uuid",
				}},
			},
		},
	})
	Register(&Domain{
		Name:        "vendor-message",
		Description: "Manage vendor messages by public GUID",
		APIPath:     "/api/v1/vendor/messages",
		Actions: []Action{
			{
				Name:        "send",
				Description: "Send a vendor message from a GUID-only JSON model",
				ToolName:    "UteamupVendorMessageSend",
				MCPOnly:     true,
				Flags: []FlagDef{{
					Name:        "from-json",
					BodyName:    "model",
					Description: "JSON file containing vendorGuid and message content",
					Required:    true,
					Type:        "string",
					JSONFile:    true,
				}},
			},
			{
				Name:        "list",
				Description: "List messages for a vendor public GUID",
				ToolName:    "UteamupVendorMessageList",
				RESTPath:    "vendor/{vendorGuid}",
				HTTPMethod:  "GET",
				Args: []ArgDef{{
					Name: "vendorGuid", Description: "Vendor public GUID", Required: true, Type: "uuid",
				}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int", QueryName: "page"},
					{Name: "page-size", Short: "s", Description: "Page size", Default: 25, Type: "int", QueryName: "pageSize"},
				},
			},
			{
				Name:        "unread",
				Description: "Get unread message count for a vendor public GUID",
				ToolName:    "UteamupVendorMessageUnreadCount",
				RESTPath:    "vendor/{vendorGuid}/unread",
				HTTPMethod:  "GET",
				Args: []ArgDef{{
					Name: "vendorGuid", Description: "Vendor public GUID", Required: true, Type: "uuid",
				}},
			},
		},
	})
	Register(&Domain{
		Name:        "vendor-rating",
		Description: "Manage vendor ratings by public GUID",
		APIPath:     "/api/v1/vendorratings",
		Actions: []Action{
			{
				Name:        "submit",
				Description: "Submit a vendor rating from a GUID-only JSON model",
				ToolName:    "UteamupVendorRatingSubmit",
				MCPOnly:     true,
				Flags: []FlagDef{{
					Name:        "from-json",
					BodyName:    "model",
					Description: "JSON file containing vendorGuid and rating details",
					Required:    true,
					Type:        "string",
					JSONFile:    true,
				}},
			},
			{
				Name:        "list",
				Description: "List ratings for a vendor public GUID",
				ToolName:    "UteamupVendorRatingList",
				RESTPath:    "by-guid/{vendorGuid}",
				HTTPMethod:  "GET",
				Args: []ArgDef{{
					Name: "vendorGuid", Description: "Vendor public GUID", Required: true, Type: "uuid",
				}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int", QueryName: "page"},
					{Name: "page-size", Short: "s", Description: "Page size", Default: 25, Type: "int", QueryName: "pageSize"},
				},
			},
			{
				Name:        "aggregate",
				Description: "Get aggregate ratings for a vendor public GUID",
				ToolName:    "UteamupVendorRatingAggregate",
				RESTPath:    "by-guid/{vendorGuid}/aggregate",
				HTTPMethod:  "GET",
				Args: []ArgDef{{
					Name: "vendorGuid", Description: "Vendor public GUID", Required: true, Type: "uuid",
				}},
			},
			{
				Name:        "flag",
				Description: "Flag a vendor rating by public GUID",
				ToolName:    "UteamupVendorRatingFlag",
				MCPOnly:     true,
				Args: []ArgDef{{
					Name: "ratingGuid", Description: "Vendor rating public GUID", Required: true, Type: "uuid",
				}},
				Flags: []FlagDef{{
					Name:        "from-json",
					BodyName:    "model",
					Description: "JSON file containing the moderation flag reason",
					Required:    true,
					Type:        "string",
					JSONFile:    true,
				}},
			},
		},
	})
}
