package registry

// Mirrors the MCP UteamupDistanceTax* tools backed by VehicleComplianceController
// (api/fleet/compliance) — the globally-modelled distance-tax surface. Iceland's
// kilometragjald is scheme "is.kilometragjald"; rates are effective-dated data rows.
func init() {
	Register(&Domain{
		Name:        "distance-tax",
		Aliases:     []string{"kilometragjald", "dt"},
		Description: "Distance-based road tax schemes, reports and statutory readings due",
		APIPath:     "/api/fleet/compliance",
		Actions: []Action{
			{
				Name:        "schemes",
				Description: "List the distance-tax schemes visible to the tenant (global packs + tenant overrides)",
				ToolName:    "UteamupDistanceTaxSchemes",
				HTTPMethod:  "GET",
				RESTPath:    "distance-tax/schemes",
			},
			{
				Name:        "report",
				Description: "Compute a distance-tax report for a scheme and period from the statutory meter ledger",
				ToolName:    "UteamupDistanceTaxReport",
				HTTPMethod:  "GET",
				RESTPath:    "distance-tax/report",
				Flags: []FlagDef{
					{Name: "scheme", Description: "Scheme key (e.g. is.kilometragjald)", Required: true, Type: "string"},
					{Name: "from", Description: "Period start (ISO 8601 date)", Required: true, Type: "string"},
					{Name: "to", Description: "Period end (ISO 8601 date)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "statutory-due",
				Description: "List rental vehicles whose statutory odometer reading is due within 30 days",
				ToolName:    "UteamupDistanceTaxStatutoryDue",
				HTTPMethod:  "GET",
				RESTPath:    "statutory-readings/due",
			},
		},
	})
}
