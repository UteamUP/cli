package registry

func init() {
	Register(&Domain{
		Name:        "maintenance-analytics",
		Description: "Measured maintenance indicators with explicit data coverage",
		APIPath:     "/api/analytics",
		Actions: []Action{
			{
				Name: "published-window-compliance", Description: "Actual preventive completions against effective published schedule windows",
				ToolName: "UteamupAnalyticsPmScheduleCompliance", RESTPath: "pm-schedule-compliance", HTTPMethod: "GET",
				Flags: []FlagDef{{Name: "timeframe", Description: "Bounded period: 7d, 30d, 90d or ytd", Type: "string", Default: "30d"}},
			},
			{
				Name: "backlog-weeks", Description: "Open estimates divided by recorded capacity for the next full ISO week",
				ToolName: "UteamupAnalyticsMaintenanceBacklogWeeks", RESTPath: "maintenance-backlog-weeks", HTTPMethod: "GET",
			},
			{
				Name: "emergency-percentage", Description: "Explicit emergency classification among actual completed work orders",
				ToolName: "UteamupAnalyticsEmergencyWorkPercentage", RESTPath: "emergency-work-percentage", HTTPMethod: "GET",
				Flags: []FlagDef{
					{Name: "start-utc", QueryName: "startUtc", Description: "UTC window start (defaults to 30 days ago)", Type: "string"},
					{Name: "end-utc", QueryName: "endUtc", Description: "Exclusive UTC window end (defaults to now); maximum window 366 days", Type: "string"},
				},
			},
			{
				Name: "mttf", Description: "Installation to first recorded failure, with missing, invalid and censored coverage",
				ToolName: "UteamupAnalyticsMeanTimeToFirstFailure", RESTPath: "mean-time-to-first-failure", HTTPMethod: "GET",
			},
			{
				Name: "stockout-time-rate", Description: "Time at zero on-hand from continuous applied balance history",
				ToolName: "UteamupAnalyticsStockoutTimeRate", RESTPath: "stockout-time-rate", HTTPMethod: "GET",
				Flags: []FlagDef{
					{Name: "start-utc", QueryName: "startUtc", Description: "UTC window start (defaults to 30 days ago)", Type: "string"},
					{Name: "end-utc", QueryName: "endUtc", Description: "Exclusive UTC window end (defaults to now); maximum window 366 days", Type: "string"},
				},
			},
			{
				Name: "value-overview", Description: "Current stock, asset and combined value (needs InventoryCost.View and/or AssetCost.View)",
				ToolName: "UteamupAnalyticsValueOverview", RESTPath: "value-overview", HTTPMethod: "GET",
			},
		},
	})
}
