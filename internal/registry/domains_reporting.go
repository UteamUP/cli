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
				Description: "List completion reports",
				ToolName:    "UteamupReportList",
				HTTPMethod:  "GET",
				Flags: append(paginationFlags(),
					FlagDef{Name: "name-filter", Description: "Filter by report description", Type: "string"},
					FlagDef{Name: "workorder-guid", Description: "Filter by stable workorder GUID", Type: "uuid"},
				),
			},
			{
				Name:        "get",
				Description: "Get a completion report by stable public GUID",
				ToolName:    "UteamupReportGet",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{reportGuid}",
				Args: []ArgDef{
					{Name: "reportGuid", Description: "Stable public report GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "detail",
				Description: "Get enriched report review detail by stable public GUID",
				ToolName:    "UteamupReportDetail",
				HTTPMethod:  "GET",
				RESTPath:    "detail/by-guid/{reportGuid}",
				Args: []ArgDef{
					{Name: "reportGuid", Description: "Stable public report GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create a completion report for a workorder GUID",
				ToolName:    "UteamupReportCreate",
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
					{Name: "primary-reporter-id", Description: "Optional identity user key", Type: "string"},
					{Name: "additional-worker-ids", Description: "Optional identity user keys", Type: "stringSlice"},
					{Name: "external-worker-emails", Description: "Optional external worker emails", Type: "stringSlice"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a completion report by stable public GUID",
				ToolName:    "UteamupReportDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{reportGuid}",
				Args: []ArgDef{
					{Name: "reportGuid", Description: "Stable public report GUID", Required: true, Type: "uuid"},
				},
			},
			// TODO(debug): V02 - expose GUID report update when the CLI supports this route's multipart form model.
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
					{Name: "end-date", Description: "Inclusive end date (YYYY-MM-DD)", Required: true, Type: "string"},
					{Name: "group-by", Description: "Grouping: day, week, month, quarter, or year", Default: "month", Type: "string"},
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
					{Name: "page", Description: "Page number", Default: "1", Type: "int"},
					{Name: "page-size", Description: "Page size", Default: "20", Type: "int"},
				},
			},
		},
	})
	Register(&Domain{Name: "analytics", Description: "View maintenance analytics", Actions: listGetActions("MaintenanceAnalytics")})
	Register(&Domain{Name: "forecast", Aliases: []string{"forecasts"}, Description: "View forecasts", Actions: listGetActions("Forecast")})
	Register(&Domain{Name: "ifta", Description: "Manage IFTA records", Actions: crudActions("Ifta")})
	Register(&Domain{Name: "meter-reading", Aliases: []string{"meter"}, Description: "Manage meter readings", Actions: crudActions("MeterReading")})
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
