package registry

func init() {
	Register(&Domain{
		Name:        "plan-impact",
		Aliases:     []string{"planimpact"},
		Description: "Preview the subscriber impact of a proposed plan price change (read-only)",
		APIPath:     "/api/planimpact",
		Actions: []Action{
			{
				Name:        "preview",
				Description: "Simulate a price change: grandfathered vs exposed subscribers + projected MRR delta. Nothing is modified.",
				ToolName:    "UteamupPlanImpactPreview",
				HTTPMethod:  "POST",
				RESTPath:    "by-plan/{planGuid}/preview",
				Args:        []ArgDef{{Name: "planGuid", Description: "Plan GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "proposed-price-per-license-isk", Description: "Proposed per-license price in ISK (required)", Required: true, Type: "float"},
					{Name: "proposed-price-per-helpdesk-license-isk", Description: "Proposed per-helpdesk-license price in ISK (required)", Required: true, Type: "float"},
				},
			},
		},
	})
}
