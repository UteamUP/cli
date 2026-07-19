package registry

func init() {
	Register(&Domain{
		Name:        "registered-report",
		Aliases:     []string{"approved-report", "report-query"},
		Description: "Discover and run allowlisted deterministic reports without arbitrary SQL",
		APIPath:     "/api/reports/registered",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List registered report definitions and accepted filters",
				ToolName:    "UteamupReportList",
			},
			{
				Name:        "run",
				Description: "Run one registered report with a validated bounded date range",
				ToolName:    "UteamupReportGenerate",
				HTTPMethod:  "POST",
				RESTPath:    "query",
				Flags: []FlagDef{
					{Name: "report-key", BodyName: "reportKey", Description: "Registered report key", Required: true, Type: "string"},
					{Name: "start-date", BodyName: "startDate", Description: "Optional ISO-8601 period start", Type: "string"},
					{Name: "end-date", BodyName: "endDate", Description: "Optional ISO-8601 period end", Type: "string"},
				},
			},
		},
	})
}
