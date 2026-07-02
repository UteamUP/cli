package registry

func init() {
	Register(&Domain{
		Name:        "plan-migration",
		Aliases:     []string{"plan-migrations", "planmigration"},
		Description: "Bulk-migrate subscriptions between plans (per-tenant partial success)",
		APIPath:     "/api/planmigration",
		Actions: []Action{
			{
				Name:        "migrate",
				Description: "Migrate every subscription from one plan to another. Dry-run by default — pass --dry-run=false to execute.",
				ToolName:    "UteamupPlanMigrationRun",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "from-plan-guid", Description: "Source plan GUID (required)", Required: true, Type: "string"},
					{Name: "to-plan-guid", Description: "Target plan GUID (required)", Required: true, Type: "string"},
					// Default true mirrors PlanMigrationRequestModel.DryRun — the safe
					// default for a bulk mutation; the flag is always sent so the CLI
					// never relies on the backend default silently changing.
					{Name: "dry-run", Description: "List the affected tenants only; nothing changes (default true)", Default: true, Type: "bool"},
				},
			},
		},
	})
}
