package registry

func init() {
	Register(&Domain{
		Name:        "quick-report",
		Aliases:     []string{"qr", "field-report"},
		Description: "Field worker quick fault reports and work order completion",
		Actions: []Action{
			{
				Name:        "create",
				Description: "Create a quick fault report (creates work order in Requested status)",
				ToolName:    "uteamup_quickreport_create",
				Flags: []FlagDef{
					{Name: "description", Short: "d", Description: "Fault description", Required: true, Type: "string"},
					{Name: "severity", Short: "s", Description: "Severity: LOW, MEDIUM, HIGH, CRITICAL", Type: "string", Default: "MEDIUM"},
					{Name: "asset-id", Short: "a", Description: "Asset ID (from QR scan or manual selection)", Type: "int"},
					{Name: "asset-code", Short: "c", Description: "Asset code (KKS or other identifier)", Type: "string"},
					{Name: "tags", Short: "t", Description: "Comma-separated tags (e.g., leak,valve,corrosion)", Type: "string"},
					{Name: "latitude", Description: "GPS latitude", Type: "float"},
					{Name: "longitude", Description: "GPS longitude", Type: "float"},
				},
			},
			{
				Name:        "complete",
				Description: "Quick-complete a work order with a summary of work done",
				ToolName:    "uteamup_quickreport_complete",
				Flags: []FlagDef{
					{Name: "workorder-id", Short: "w", Description: "Work order ID to complete", Required: true, Type: "int"},
					{Name: "summary", Short: "s", Description: "Summary of work done", Required: true, Type: "string"},
					{Name: "partial", Short: "p", Description: "Mark as partial completion (keeps WO in progress)", Type: "bool"},
				},
			},
		},
	})
}
