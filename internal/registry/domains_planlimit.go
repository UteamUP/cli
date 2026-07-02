package registry

func init() {
	Register(&Domain{
		Name:        "plan-limit",
		Aliases:     []string{"plan-limits", "planlimit", "planlimits"},
		Description: "Read and upsert per-plan resource limits",
		// PlanLimitController routes under /api/planlimit — a different
		// controller from the `plan` domain, so it needs its own base path.
		APIPath: "/api/planlimit",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List a plan's resource limits",
				ToolName:    "UteamupPlanLimitList",
				RESTPath:    "by-plan/{planGuid}",
				Args:        []ArgDef{{Name: "planGuid", Description: "Plan GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "upsert",
				Description: "Create or update the limit for one dimension on a plan (omit --max-value for unlimited)",
				ToolName:    "UteamupPlanLimitUpsert",
				HTTPMethod:  "PUT",
				RESTPath:    "by-plan/{planGuid}",
				Args:        []ArgDef{{Name: "planGuid", Description: "Plan GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "dimension", Description: "Limit dimension (PlanLimitDimension enum value, required)", Required: true, Type: "int"},
					{Name: "max-value", Description: "Cap for the dimension (>= 0); omit for unlimited / to remove the cap", Type: "int"},
				},
			},
		},
	})
}
