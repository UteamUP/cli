package registry

func init() {
	Register(&Domain{Name: "report", Aliases: []string{"reports"}, Description: "Manage reports (enriched detail includes cost breakdown, checklists, meter readings, labour, tool usage)", Actions: crudActions("Report")})
	Register(&Domain{Name: "report-analytics", Aliases: []string{"report-stats"}, Description: "View report analytics with cost trends, top assets, and completion metrics (params: startDate, endDate, groupBy)", Actions: listGetActions("ReportAnalytics")})
	Register(&Domain{Name: "asset-reports", Description: "View reports for a specific asset with summary stats (params: assetId, startDate, endDate)", Actions: listGetActions("AssetReports")})
	Register(&Domain{Name: "analytics", Description: "View maintenance analytics", Actions: listGetActions("MaintenanceAnalytics")})
	Register(&Domain{Name: "forecast", Aliases: []string{"forecasts"}, Description: "View forecasts", Actions: listGetActions("Forecast")})
	Register(&Domain{Name: "ifta", Description: "Manage IFTA records", Actions: crudActions("Ifta")})
	Register(&Domain{Name: "meter-reading", Aliases: []string{"meter"}, Description: "Manage meter readings", Actions: crudActions("MeterReading")})
	Register(&Domain{Name: "cost-overview", Aliases: []string{"costs"}, Description: "View cost overviews", Actions: listGetActions("CostOverview")})
}
