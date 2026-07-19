package registry

func init() {
	Register(&Domain{
		Name:        "fleet-intelligence",
		Aliases:     []string{"fleet-ai"},
		Description: "Review fleet anomalies, tire readiness, and replacement TCO evidence",
		APIPath:     "/api/fleet/intelligence",
		Actions: []Action{
			{
				Name:        "anomalies",
				Description: "Review comparable fuel and idling anomalies with readiness evidence",
				ToolName:    "UteamupFleetIntelligenceGetAnomalies",
				HTTPMethod:  "GET",
				RESTPath:    "fuel-idling-anomalies",
				Flags: []FlagDef{
					{Name: "lookback-days", BodyName: "lookbackDays", Description: "Analysis window in days, 7-365", Type: "int", Default: 90},
					{Name: "minimum-history", BodyName: "minimumHistory", Description: "Minimum comparable history, 5-100", Type: "int", Default: 5},
				},
			},
			{
				Name:        "tire-readiness",
				Description: "Review typed tire measurement and safety readiness for an asset",
				ToolName:    "UteamupFleetIntelligenceGetTireReadiness",
				HTTPMethod:  "GET",
				RESTPath:    "assets/{assetGuid}/tire-readiness",
				Args: []ArgDef{
					{Name: "assetGuid", Description: "Tenant-scoped public asset GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "replacement-tco",
				Description: "Compare keep-versus-replace TCO with explicit single-currency assumptions",
				ToolName:    "UteamupFleetIntelligenceGetReplacementTco",
				HTTPMethod:  "POST",
				RESTPath:    "assets/{assetGuid}/replacement-tco",
				Args: []ArgDef{
					{Name: "assetGuid", Description: "Tenant-scoped public asset GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "currency", Description: "ISO 4217 currency code", Required: true, Type: "string"},
					{Name: "horizon-months", BodyName: "horizonMonths", Description: "Analysis horizon in months, 12-120", Type: "int", Default: 36},
					{Name: "replacement-estimate", BodyName: "replacementEstimate", Description: "Optional sourced replacement estimate", Type: "float"},
					{Name: "residual-value", BodyName: "residualValue", Description: "Explicit residual value assumption", Required: true, Type: "float"},
					{Name: "financing-cost", BodyName: "financingCost", Description: "Explicit financing cost assumption", Required: true, Type: "float"},
					{Name: "downtime-cost-per-hour", BodyName: "downtimeCostPerHour", Description: "Explicit downtime cost per hour", Required: true, Type: "float"},
					{Name: "replacement-annual-maintenance-cost", BodyName: "replacementAnnualMaintenanceCost", Description: "Explicit replacement annual maintenance cost", Required: true, Type: "float"},
					{Name: "replacement-fuel-reduction-percent", BodyName: "replacementFuelReductionPercent", Description: "Expected fuel reduction percentage, 0-100", Required: true, Type: "float"},
				},
			},
		},
	})
}
