package registry

// Saved report templates are authored in the web builder only, so the CLI has no
// create or update verb and no verb for the unsaved-definition preview, export or
// catalog routes. Every action needs Report.Export.
func init() {
	templateGuidArg := []ArgDef{
		{Name: "templateGuid", Description: "Stable public report template GUID", Required: true, Type: "non-empty-uuid"},
	}

	Register(&Domain{
		Name:        "report-template",
		Aliases:     []string{"rt", "report-templates"},
		Description: "List, preview, export and delete saved report templates by GUID (needs Report.Export)",
		APIPath:     "/api/reports/templates",
		Actions: []Action{
			{
				Name:              "list",
				Description:       "List the saved report templates you can see",
				ToolName:          "UteamupReportTemplateList",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					{Name: "scope", QueryName: "scope", Description: "Which templates: all, mine or shared", Type: "string",
						AllowedValues: []string{"all", "mine", "shared"}},
					{Name: "search", QueryName: "search", Description: "Case-insensitive template name search", Type: "string"},
					{Name: "page", QueryName: "page", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Items per page (1-100)", Default: 50, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get a saved report template by GUID",
				ToolName:    "UteamupReportTemplateGet",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{templateGuid}",
				Args:        templateGuidArg,
			},
			{
				Name:        "preview",
				Description: "Preview a saved report template: at most 50 formatted rows plus the total count",
				ToolName:    "UteamupReportTemplatePreview",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{templateGuid}/preview",
				Args:        templateGuidArg,
			},
			{
				Name:        "export",
				Description: "Save a saved report template as a CSV, Excel or PDF file",
				ToolName:    "UteamupReportTemplateExport",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{templateGuid}/export",
				Args:        templateGuidArg,
				Flags: []FlagDef{
					{Name: "format", QueryName: "format", Description: "File format: csv, xlsx or pdf", Default: "csv", Type: "string",
						AllowedValues: []string{"csv", "xlsx", "pdf"}},
					{Name: "out", Description: "Output file; an existing file is never overwritten " +
						"(default report-template-<templateGuid>.<format>)", Type: "string"},
				},
				DownloadResponseBody: true,
				DownloadOutputFlag:   "out",
				DownloadDefaultName:  "report-template-{templateGuid}.{format}",
			},
			{
				Name:        "delete",
				Description: "Delete a saved report template by GUID (refused while schedules still use it)",
				ToolName:    "UteamupReportTemplateDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{templateGuid}",
				Args:        templateGuidArg,
			},
		},
	})
}
