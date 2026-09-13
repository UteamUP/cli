package registry

func init() {
	Register(&Domain{
		Name:        "itmanagement",
		Aliases:     []string{"it", "it-management"},
		Description: "Inspect the selected tenant's IT Management connection, alerts, rules, resources and network flows",
		Actions: []Action{
			{
				Name:        "dashboard",
				Description: "Show connection health, open alerts by severity, servers, patching and unmapped resources",
				ToolName:    "UteamupITManagementDashboard",
			},
			{
				Name:        "alerts",
				Description: "List IT alerts with state, severity, source, asset and time filters",
				ToolName:    "UteamupITManagementAlertsList",
				Flags: []FlagDef{
					{Name: "state", Description: "open, acknowledged or resolved", Type: "string"},
					{Name: "severity", BodyName: "minimumSeverity", Description: "Minimum severity: info, warning, error or critical", Type: "string"},
					{Name: "source", Description: "Source filter (for example securityAlert, heartbeat, networkFlow)", Type: "string"},
					{Name: "asset-guid", BodyName: "assetGuid", Description: "Asset GUID filter", Type: "string"},
					{Name: "rule-guid", BodyName: "ruleGuid", Description: "Rule GUID filter", Type: "string"},
					{Name: "from", Description: "UTC range start (ISO-8601)", Type: "string"},
					{Name: "to", Description: "UTC range end (ISO-8601)", Type: "string"},
					{Name: "search", Description: "Title, resource or computer contains", Type: "string"},
					{Name: "limit", BodyName: "pageSize", Description: "Page size (1-200)", Type: "int", Default: 50},
					{Name: "page", Description: "Page number", Type: "int", Default: 1},
				},
			},
			{
				Name:        "alert",
				Description: "Show one IT alert",
				ToolName:    "UteamupITManagementAlertGet",
				Flags: []FlagDef{
					{Name: "alert-guid", BodyName: "alertGuid", Description: "Alert GUID", Type: "string", Required: true},
				},
			},
			{
				Name:        "alert-ack",
				Description: "Acknowledge an IT alert with a reason code",
				ToolName:    "UteamupITManagementAlertAcknowledge",
				Flags: []FlagDef{
					{Name: "alert-guid", BodyName: "alertGuid", Description: "Alert GUID", Type: "string", Required: true},
					{Name: "reason", BodyName: "reasonCode", Description: "falsePositive, fixed, duplicate, plannedMaintenance, expected or other", Type: "string"},
					{Name: "note", Description: "Optional note", Type: "string"},
				},
			},
			{
				Name:        "alert-resolve",
				Description: "Resolve an IT alert with a reason code",
				ToolName:    "UteamupITManagementAlertResolve",
				Flags: []FlagDef{
					{Name: "alert-guid", BodyName: "alertGuid", Description: "Alert GUID", Type: "string", Required: true},
					{Name: "reason", BodyName: "reasonCode", Description: "falsePositive, fixed, duplicate, plannedMaintenance, expected or other", Type: "string"},
					{Name: "note", Description: "Optional note", Type: "string"},
				},
			},
			{
				Name:        "rules",
				Description: "List IT alert rules with their noise controls and simulation state",
				ToolName:    "UteamupITManagementRulesList",
			},
			{
				Name:        "connections",
				Description: "Show Azure connection health (status only, never secrets)",
				ToolName:    "UteamupITManagementConnectionsList",
			},
			{
				Name:        "resources",
				Description: "List discovered Azure resources and their asset mapping",
				ToolName:    "UteamupITManagementResourcesList",
				Flags: []FlagDef{
					{Name: "unmapped", BodyName: "unmappedOnly", Description: "Only resources without an asset", Type: "bool"},
					{Name: "limit", BodyName: "pageSize", Description: "Page size (1-200)", Type: "int", Default: 50},
					{Name: "page", Description: "Page number", Type: "int", Default: 1},
				},
			},
			{
				Name:        "servicemap",
				Description: "Summarise the service map: nodes, accepted and proposed connections, busiest edges",
				ToolName:    "UteamupITManagementServiceMapSummary",
				Flags: []FlagDef{
					{Name: "window-hours", BodyName: "windowHours", Description: "Flow window in hours (1-168)", Type: "int", Default: 24},
				},
			},
			{
				Name:        "flows",
				Description: "List network flows for an asset over a window",
				ToolName:    "UteamupITManagementFlowsSummary",
				Flags: []FlagDef{
					{Name: "source-asset-guid", BodyName: "sourceAssetGuid", Description: "Source asset GUID", Type: "string"},
					{Name: "target-asset-guid", BodyName: "targetAssetGuid", Description: "Target asset GUID", Type: "string"},
					{Name: "window", Description: "24h or 7d", Type: "string", Default: "24h"},
				},
			},
		},
	})
}
