package registry

// Report schedules are created and edited in the web builder only, so the CLI has
// no create or update verb. Every action needs Report.Export and the tenant's
// Scheduled Report Delivery plan feature.
func init() {
	scheduleGuidArg := []ArgDef{
		{Name: "scheduleGuid", Description: "Stable public report schedule GUID", Required: true, Type: "non-empty-uuid"},
	}

	Register(&Domain{
		Name:    "report-schedule",
		Aliases: []string{"rsch", "report-schedules"},
		Description: "List, pause, resume, test-send and delete scheduled report deliveries by GUID " +
			"(needs Report.Export and the Scheduled Report Delivery plan feature)",
		APIPath: "/api/reports/schedules",
		Actions: []Action{
			{
				Name:              "list",
				Description:       "List the report schedules you can manage",
				ToolName:          "UteamupReportScheduleList",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					{Name: "search", QueryName: "search", Description: "Case-insensitive schedule name search", Type: "string"},
					{Name: "active", QueryName: "isActive", Description: "Only active (--active) or paused (--active=false) schedules", Type: "bool"},
					{Name: "mine", QueryName: "mineOnly", Description: "Only schedules you own", Type: "bool"},
					{Name: "page", QueryName: "page", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Items per page (1-100)", Default: 50, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get a report schedule with its recipients and recent runs",
				ToolName:    "UteamupReportScheduleGet",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{scheduleGuid}",
				Args:        scheduleGuidArg,
			},
			{
				Name:        "pause",
				Description: "Pause a report schedule; pausing a paused schedule changes nothing",
				ToolName:    "UteamupReportScheduleSetActive",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{scheduleGuid}/pause",
				Args:        scheduleGuidArg,
			},
			{
				Name:        "resume",
				Description: "Resume a report schedule from now (refused when the owner lost access or no recipient is deliverable)",
				ToolName:    "UteamupReportScheduleSetActive",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{scheduleGuid}/resume",
				Args:        scheduleGuidArg,
			},
			{
				Name:        "send-test",
				Description: "Send the schedule's report once to your own confirmed email without advancing the schedule",
				ToolName:    "UteamupReportScheduleSendTest",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{scheduleGuid}/send-test",
				Args:        scheduleGuidArg,
			},
			{
				Name:        "delete",
				Description: "Delete a report schedule; its run history is kept",
				ToolName:    "UteamupReportScheduleDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{scheduleGuid}",
				Args:        scheduleGuidArg,
			},
		},
	})
}
