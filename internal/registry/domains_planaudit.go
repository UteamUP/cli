package registry

func init() {
	Register(&Domain{
		Name:        "plan-audit",
		Aliases:     []string{"plan-audits", "planaudit"},
		Description: "Read the plan-change audit trail",
		// PlanAuditController routes under /api/planaudit — a different
		// controller from the `plan` domain, so it needs its own base path.
		APIPath: "/api/planaudit",
		Actions: []Action{
			{
				Name:        "history",
				Description: "List a plan's change history (most recent first)",
				ToolName:    "UteamupPlanAuditHistory",
				RESTPath:    "by-plan/{planGuid}",
				Args:        []ArgDef{{Name: "planGuid", Description: "Plan GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "export",
				Description: "Export a plan's change history as CSV",
				ToolName:    "UteamupPlanAuditExport",
				RESTPath:    "by-plan/{planGuid}/export",
				Args:        []ArgDef{{Name: "planGuid", Description: "Plan GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
		},
	})
}
