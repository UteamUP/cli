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
		// Routes mirror OperationalRoute{,Schedule,Execution}Controller. Without APIPath the REST
		// fallback derived /api/route — no such controller, so every verb 404'd. Actions whose REST
		// adapter could not be confirmed against a real [Http*] template are MCPOnly (their governed
		// MCP tool is the working surface) rather than pointed at a guessed URL.
		APIPath: "/api/operationalroutes",
		Actions: []Action{
			{Name: "list", Description: "List operational routes", ToolName: "UteamupOperationalRouteList", HTTPMethod: "GET", Flags: append(paginationFlags(), FlagDef{Name: "is-active", BodyName: "isActive", Description: "Filter active status", Type: "bool"})},
			{Name: "get", Description: "Get an operational route by GUID", ToolName: "UteamupOperationalRouteGet", HTTPMethod: "GET", RESTPath: "by-guid/{routeGuid}", Args: routeGUIDArg},
			{Name: "schedules", Description: "List schedules for a route GUID", ToolName: "UteamupInspectionScheduleList", MCPOnly: true, Args: routeGUIDArg},
			{Name: "overdue", Description: "List overdue route schedules", ToolName: "UteamupInspectionScheduleGetOverdue", HTTPMethod: "GET", RESTBasePath: "/api/operationalrouteschedules", RESTPath: "overdue"},
			{Name: "executions", Description: "List route executions", ToolName: "UteamupInspectionExecutionList", HTTPMethod: "GET", RESTBasePath: "/api/operationalrouteexecutions", Flags: append(paginationFlags(), FlagDef{Name: "route-guid", BodyName: "routeGuid", Description: "Filter by public route GUID", Type: "string"})},
			{Name: "execution", Description: "Get a route execution by GUID", ToolName: "UteamupInspectionExecutionGet", HTTPMethod: "GET", RESTBasePath: "/api/operationalrouteexecutions", RESTPath: "{executionGuid}", Args: executionGUIDArg},
			{Name: "start", Description: "Start a route execution using GUID data", ToolName: "UteamupInspectionExecutionStart", HTTPMethod: "POST", RESTBasePath: "/api/operationalrouteexecutions", RESTPath: "by-guid/start", Flags: []FlagDef{jsonFlag()}},
			{Name: "complete-stop", Description: "Complete a route stop by GUID", ToolName: "UteamupInspectionStopComplete", HTTPMethod: "PUT", RESTBasePath: "/api/operationalrouteexecutions", RESTPath: "by-guid/stops/{stopGuid}/complete", Args: []ArgDef{
				{Name: "stopGuid", Description: "Public route-template stop GUID", Required: true, Type: "string"},
			}, Flags: []FlagDef{jsonFlag()}},
			{Name: "flag-issue", Description: "Flag an issue at a route stop GUID", ToolName: "UteamupInspectionIssueflag", HTTPMethod: "POST", RESTBasePath: "/api/operationalrouteexecutions", RESTPath: "by-guid/stops/{stopGuid}/flag-issue", Args: []ArgDef{
				{Name: "stopGuid", Description: "Public route-template stop GUID", Required: true, Type: "string"},
			}, Flags: []FlagDef{jsonFlag()}},
			{Name: "complete", Description: "Complete a route execution by GUID", ToolName: "UteamupInspectionExecutionComplete", HTTPMethod: "PUT", RESTBasePath: "/api/operationalrouteexecutions", RESTPath: "by-guid/{executionGuid}/complete", Args: executionGUIDArg},
			{Name: "abandon", Description: "Abandon a route execution by GUID", ToolName: "UteamupInspectionExecutionAbandon", HTTPMethod: "PUT", RESTBasePath: "/api/operationalrouteexecutions", RESTPath: "by-guid/{executionGuid}/abandon", Args: append(executionGUIDArg, ArgDef{
				Name: "reason", Description: "Reason for abandoning the execution", Required: true, Type: "string",
			})},
			{Name: "analytics", Description: "Get tenant route analytics", ToolName: "UteamupInspectionAnalyticsOverview", MCPOnly: true},
			{Name: "anomalies", Description: "List route anomalies", ToolName: "UteamupInspectionAnomalyList", MCPOnly: true, Flags: []FlagDef{
				{Name: "route-guid", BodyName: "routeGuid", Description: "Filter by public route GUID", Type: "string"},
				{Name: "status", Description: "Filter by anomaly status", Type: "string"},
			}},
			{Name: "asset-health", Description: "Get inspection health for an asset GUID", ToolName: "UteamupInspectionAssetHealthscore", MCPOnly: true, Args: assetGUIDArg},
			{
				Name:         "optimize",
				Description:  "Get a deterministic review-only route optimization by public GUID",
				ToolName:     "UteamupOperationalRouteOptimize",
				HTTPMethod:   "GET",
				RESTBasePath: "/api/inspectionai",
				RESTPath:     "routes/by-guid/{routeGuid}/optimize",
				Args:         routeGUIDArg,
			},
		},
	})

	Register(&Domain{Name: "automation", Aliases: []string{"automations"}, Description: "Manage automations", Actions: crudActions("Automation")})
	Register(&Domain{Name: "notification", Aliases: []string{"notifications"}, Description: "Manage notifications", Actions: crudActions("Notification")})
	Register(&Domain{Name: "helpdesk", Description: "Manage helpdesk", Actions: crudActions("Helpdesk")})
	Register(&Domain{Name: "extension", Aliases: []string{"extensions"}, Description: "Manage extensions", Actions: listGetActions("Extension")})
	Register(&Domain{Name: "weather", Description: "Get weather data", Actions: listGetActions("Weather")})
}
