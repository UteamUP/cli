package registry

func init() {
	routeGUIDArg := []ArgDef{{
		Name:        "routeGuid",
		Description: "Public operational route GUID",
		Required:    true,
		Type:        "string",
	}}
	executionGUIDArg := []ArgDef{{
		Name:        "executionGuid",
		Description: "Public route execution GUID",
		Required:    true,
		Type:        "string",
	}}
	assetGUIDArg := []ArgDef{{
		Name:        "assetGuid",
		Description: "Public asset GUID",
		Required:    true,
		Type:        "string",
	}}

	Register(&Domain{
		Name:        "route",
		Aliases:     []string{"routes", "operational-route"},
		Description: "Manage operational routes",
		Actions: []Action{
			{Name: "list", Description: "List operational routes", ToolName: "UteamupOperationalRouteList", Flags: append(paginationFlags(), FlagDef{Name: "is-active", BodyName: "isActive", Description: "Filter active status", Type: "bool"})},
			{Name: "get", Description: "Get an operational route by GUID", ToolName: "UteamupOperationalRouteGet", Args: routeGUIDArg},
			{Name: "schedules", Description: "List schedules for a route GUID", ToolName: "UteamupInspectionScheduleList", Args: routeGUIDArg},
			{Name: "overdue", Description: "List overdue route schedules", ToolName: "UteamupInspectionScheduleGetOverdue"},
			{Name: "executions", Description: "List route executions", ToolName: "UteamupInspectionExecutionList", Flags: append(paginationFlags(), FlagDef{Name: "route-guid", BodyName: "routeGuid", Description: "Filter by public route GUID", Type: "string"})},
			{Name: "execution", Description: "Get a route execution by GUID", ToolName: "UteamupInspectionExecutionGet", Args: executionGUIDArg},
			{Name: "start", Description: "Start a route execution using GUID data", ToolName: "UteamupInspectionExecutionStart", Flags: []FlagDef{jsonFlag()}},
			{Name: "complete-stop", Description: "Complete a route stop by GUID", ToolName: "UteamupInspectionStopComplete", Args: []ArgDef{
				{Name: "stopGuid", Description: "Public route-template stop GUID", Required: true, Type: "string"},
			}, Flags: []FlagDef{jsonFlag()}},
			{Name: "flag-issue", Description: "Flag an issue at a route stop GUID", ToolName: "UteamupInspectionIssueflag", Args: []ArgDef{
				{Name: "stopGuid", Description: "Public route-template stop GUID", Required: true, Type: "string"},
			}, Flags: []FlagDef{jsonFlag()}},
			{Name: "complete", Description: "Complete a route execution by GUID", ToolName: "UteamupInspectionExecutionComplete", Args: executionGUIDArg},
			{Name: "abandon", Description: "Abandon a route execution by GUID", ToolName: "UteamupInspectionExecutionAbandon", Args: append(executionGUIDArg, ArgDef{
				Name: "reason", Description: "Reason for abandoning the execution", Required: true, Type: "string",
			})},
			{Name: "analytics", Description: "Get tenant route analytics", ToolName: "UteamupInspectionAnalyticsOverview"},
			{Name: "anomalies", Description: "List route anomalies", ToolName: "UteamupInspectionAnomalyList", Flags: []FlagDef{
				{Name: "route-guid", BodyName: "routeGuid", Description: "Filter by public route GUID", Type: "string"},
				{Name: "status", Description: "Filter by anomaly status", Type: "string"},
			}},
			{Name: "asset-health", Description: "Get inspection health for an asset GUID", ToolName: "UteamupInspectionAssetHealthscore", Args: assetGUIDArg},
			{
				Name:        "optimize",
				Description: "Get a deterministic review-only route optimization by public GUID",
				ToolName:    "UteamupOperationalRouteOptimize",
				Args:        routeGUIDArg,
			},
		},
	})

	Register(&Domain{Name: "automation", Aliases: []string{"automations"}, Description: "Manage automations", Actions: crudActions("Automation")})
	Register(&Domain{Name: "notification", Aliases: []string{"notifications"}, Description: "Manage notifications", Actions: crudActions("Notification")})
	Register(&Domain{Name: "helpdesk", Description: "Manage helpdesk", Actions: crudActions("Helpdesk")})
	Register(&Domain{Name: "extension", Aliases: []string{"extensions"}, Description: "Manage extensions", Actions: listGetActions("Extension")})
	Register(&Domain{Name: "weather", Description: "Get weather data", Actions: listGetActions("Weather")})
}
