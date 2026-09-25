package registry

func init() {
	Register(&Domain{
		Name:        "report",
		Aliases:     []string{"reports"},
		Description: "Manage completion reports by stable public GUID",
		APIPath:     "/api/report",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the completion reports you can see (GUID-only items)",
				ToolName:    "UteamupWorkReportList",
				HTTPMethod:  "GET",
				RESTPath:    "worker",
				Flags: append(paginationFlags(),
					FlagDef{Name: "name-filter", Description: "Filter by report description", Type: "string"},
					FlagDef{Name: "workorder-guid", Description: "Filter by stable workorder GUID", Type: "uuid"},
				),
			},
			{
				Name:        "get",
				Description: "Get a completion report by stable public GUID (GUID-only people)",
				ToolName:    "UteamupWorkReportGet",
				HTTPMethod:  "GET",
				RESTPath:    "worker/by-guid/{reportGuid}",
				Args: []ArgDef{
					{Name: "reportGuid", Description: "Stable public report GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "detail",
				Description: "Get enriched report review detail by stable public GUID",
				ToolName:    "UteamupWorkReportDetail",
				HTTPMethod:  "GET",
				RESTPath:    "detail/by-guid/{reportGuid}",
				Args: []ArgDef{
					{Name: "reportGuid", Description: "Stable public report GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create a completion report for a workorder GUID",
				ToolName:    "UteamupWorkReportCreate",
				HTTPMethod:  "POST",
				RESTPath:    "workorder/by-guid/{workorderGuid}",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Stable public workorder GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "description", Description: "Completion report description", Required: true, Type: "string"},
					{Name: "report-date", Description: "Report timestamp", Required: true, Type: "string"},
					{Name: "close-out-notes", Description: "Optional close-out notes", Type: "string"},
					{Name: "time-spent", Description: "Optional hours spent", Type: "float"},
					{Name: "cost-incurred", Description: "Optional cost incurred", Type: "float"},
					{Name: "primary-reporter-guid", Description: "Optional primary reporter user GUID (defaults to you)", Type: "uuid"},
					{Name: "additional-worker-guids", BodyName: "additionalWorkerGuids", Description: "Optional additional worker user GUIDs, repeatable", Type: "stringSlice"},
					{Name: "external-worker-emails", Description: "Optional external worker emails (at most 10)", Type: "stringSlice"},
					{Name: "notify-external-workers", Description: "Email the report to the external workers", Type: "bool"},
					{Name: "idempotency-key", HeaderName: "Idempotency-Key", Description: "Caller-generated GUID reused only for a retry of the same report", Type: "non-empty-uuid"},
				},
			},
			{
				Name: "update",
				Description: "Replace a completion report by stable public GUID. Omitted close-out notes, " +
					"time spent and cost incurred are cleared; omitted worker lists are kept",
				ToolName:   "UteamupWorkReportUpdate",
				HTTPMethod: "PUT",
				RESTPath:   "by-guid/{reportGuid}",
				Args: []ArgDef{
					{Name: "reportGuid", Description: "Stable public report GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "description", Description: "Completion report description (replaces the current one)", Required: true, Type: "string"},
					{Name: "report-date", Description: "Report timestamp (replaces the current one)", Required: true, Type: "string"},
					{Name: "close-out-notes", Description: "Close-out notes (cleared when omitted)", Type: "string"},
					{Name: "time-spent", Description: "Hours spent (cleared when omitted)", Type: "float"},
					{Name: "cost-incurred", Description: "Cost incurred (cleared when omitted)", Type: "float"},
					{Name: "primary-reporter-guid", Description: "Primary reporter user GUID", Type: "uuid"},
					{Name: "additional-worker-guids", BodyName: "additionalWorkerGuids", Description: "Additional worker user GUIDs, repeatable (kept when omitted)", Type: "stringSlice"},
					{Name: "external-worker-emails", Description: "External worker emails, at most 10 (kept when omitted)", Type: "stringSlice"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a completion report by stable public GUID",
				ToolName:    "UteamupWorkReportDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{reportGuid}",
				Args: []ArgDef{
					{Name: "reportGuid", Description: "Stable public report GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
	Register(&Domain{
		Name:        "report-analytics",
		Aliases:     []string{"report-stats"},
		Description: "View bounded report analytics with cost trends, top assets, and completion metrics",
		APIPath:     "/api/report",
		Actions: []Action{
			{
				Name:        "read",
				Description: "Read tenant report analytics for a bounded date range",
				ToolName:    "UteamupReportAnalytics",
				HTTPMethod:  "GET",
				RESTPath:    "analytics",
				Flags: []FlagDef{
					{Name: "start-date", Description: "Inclusive start date (YYYY-MM-DD)", Required: true, Type: "string"},
					{Name: "end-date", Description: "Inclusive end date (whole day, YYYY-MM-DD)", Required: true, Type: "string"},
					{Name: "group-by", Description: "Grouping: day, week (ISO-8601), month, quarter, or year", Default: "month", Type: "string",
						AllowedValues: []string{"day", "week", "month", "quarter", "year"}},
				},
			},
		},
	})
	Register(&Domain{
		Name:        "asset-reports",
		Description: "View reports for a specific asset by its stable public GUID",
		APIPath:     "/api/report",
		Actions: []Action{
			{
				Name:        "get",
				Description: "View paginated report history and summary stats for an asset",
				ToolName:    "UteamupAssetReports",
				HTTPMethod:  "GET",
				RESTPath:    "asset/by-guid/{assetGuid}",
				Args: []ArgDef{
					{Name: "assetGuid", Description: "Stable public asset GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "start-date", Description: "Optional inclusive start date", Type: "string"},
					{Name: "end-date", Description: "Optional inclusive end date", Type: "string"},
					{Name: "page", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Description: "Page size (1-100)", Default: 20, Type: "int"},
				},
			},
		},
	})
	Register(&Domain{Name: "analytics", Description: "View maintenance analytics", Actions: listGetActions("MaintenanceAnalytics")})
	Register(&Domain{Name: "forecast", Aliases: []string{"forecasts"}, Description: "View forecasts", Actions: listGetActions("Forecast")})
	// Routes mirror IftaController (/api/fleet/ifta). The previous generic crudActions
	// derived a phantom /api/ifta base and 404'd on every action; the CSV export stays
	// out because the backend route is deliberately [NonAction]. The GUID-first
	// meter-reading domain lives in domains_meter_reading.go — the duplicate generic
	// registration that used to sit here shadowed it in help output and was removed.
	Register(&Domain{Name: "ifta", Description: "IFTA quarterly fuel-tax reporting (US distance-tax scheme)", APIPath: "/api/fleet/ifta", Actions: []Action{
		{Name: "quarterly-report", HTTPMethod: "GET", RESTPath: "report", Description: "Get the IFTA quarterly fuel-tax report", ToolName: "UteamupIftaGetQuarterlyReport", Flags: []FlagDef{
			{Name: "year", Description: "Report year (e.g. 2026)", Required: true, Type: "int"},
			{Name: "quarter", Description: "Report quarter (1-4)", Required: true, Type: "int"},
		}},
	}})
	Register(&Domain{
		Name:        "cost-overview",
		Aliases:     []string{"costs"},
		Description: "View cost overviews",
		APIPath:     "/api/costoverview",
		Actions: []Action{
			{
				Name:        "list",
				Description: "View the tenant cost overview summary",
				ToolName:    "UteamupCostOverviewSummary",
				HTTPMethod:  "GET",
			},
			{
				Name:        "get",
				Description: "Get cost details for a workorder by its stable public GUID",
				ToolName:    "UteamupCostByWorkorder",
				HTTPMethod:  "GET",
				RESTPath:    "workorders/by-guid/{workorderGuid}",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Workorder GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
