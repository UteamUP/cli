package registry

func init() {
	Register(&Domain{
		Name:        "quick-report",
		Aliases:     []string{"qr", "field-report"},
		Description: "Field worker quick fault reports and work order completion",
		Actions: []Action{
			{
				Name:        "create",
				Description: "Create a quick fault report (creates a workorder in Not Started status)",
				ToolName:    "UteamupQuickreportCreate",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Caller-generated GUID reused only when retrying the same report", Required: true, Type: "uuid"},
					{Name: "description", Short: "d", Description: "Fault description", Required: true, Type: "string"},
					{Name: "severity", Short: "s", Description: "Severity: LOW, MEDIUM, HIGH, CRITICAL", Type: "string", Default: "MEDIUM"},
					{Name: "asset-guid", Short: "a", Description: "Tenant-scoped public asset GUID", Type: "uuid"},
					{Name: "asset-code", Short: "c", Description: "Asset code (KKS or other identifier)", Type: "string"},
					{Name: "tags", Short: "t", Description: "Comma-separated tags (e.g., leak,valve,corrosion)", Type: "string"},
					{Name: "latitude", Description: "GPS latitude", Type: "float"},
					{Name: "longitude", Description: "GPS longitude", Type: "float"},
				},
			},
			{
				Name:        "complete",
				Description: "Quick-complete a work order with a summary of work done",
				ToolName:    "UteamupQuickreportComplete",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Caller-generated GUID reused only when retrying the same completion", Required: true, Type: "uuid"},
					{Name: "workorder-guid", Short: "w", Description: "Tenant-scoped public workorder GUID", Required: true, Type: "uuid"},
					{Name: "summary", Short: "s", Description: "Summary of work done", Required: true, Type: "string"},
					{Name: "partial", Short: "p", Description: "Mark as partial completion (keeps WO in progress)", Type: "bool"},
				},
			},
		},
	})
}
