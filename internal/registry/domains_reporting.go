package registry

func init() {
	Register(&Domain{Name: "report", Aliases: []string{"reports"}, Description: "Manage reports", Actions: crudActions("Report")})
	Register(&Domain{Name: "analytics", Description: "View maintenance analytics", Actions: listGetActions("MaintenanceAnalytics")})
	Register(&Domain{Name: "forecast", Aliases: []string{"forecasts"}, Description: "View forecasts", Actions: listGetActions("Forecast")})
	Register(&Domain{Name: "ifta", Description: "Manage IFTA records", Actions: crudActions("Ifta")})
	Register(&Domain{Name: "meter-reading", Aliases: []string{"meter"}, Description: "Manage meter readings", Actions: crudActions("MeterReading")})
	Register(&Domain{Name: "cost-overview", Aliases: []string{"costs"}, Description: "View cost overviews", Actions: listGetActions("CostOverview")})
}
