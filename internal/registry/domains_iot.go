package registry

func init() {
	Register(&Domain{
		Name:        "iot",
		Aliases:     []string{"internet-of-things"},
		Description: "Inspect the selected tenant's dedicated IoT Beta environment",
		Actions: []Action{
			{
				Name:        "status",
				Description: "Show environment, pricing, usage and lifecycle status",
				ToolName:    "UteamupIoTEnvironmentStatus",
			},
			{
				Name:        "monitoring",
				Description: "Show device health, ingestion usage and recent alerts",
				ToolName:    "UteamupIoTMonitoringDashboard",
			},
			{
				Name:        "telemetry",
				Description: "Query normalized telemetry with bounded cursor pagination",
				ToolName:    "UteamupIoTTelemetryPoints",
				Flags: []FlagDef{
					{Name: "from", Description: "UTC range start (ISO-8601)", Type: "string"},
					{Name: "to", Description: "UTC range end (ISO-8601)", Type: "string"},
					{Name: "device-guid", Description: "Device GUID filter", Type: "string"},
					{Name: "asset-guid", Description: "Asset GUID filter", Type: "string"},
					{Name: "attribute-definition-guid", Description: "Attribute definition GUID filter", Type: "string"},
					{Name: "limit", Description: "Page size (1-500)", Type: "int", Default: 100},
					{Name: "before-received-at", Description: "UTC cursor timestamp", Type: "string"},
					{Name: "before-point-guid", Description: "Point GUID paired with cursor timestamp", Type: "string"},
				},
			},
			{
				Name:        "rules",
				Description: "List baseline, threshold and heartbeat automation rules",
				ToolName:    "UteamupIoTRulesList",
			},
		},
	})
}
