package registry

func init() {
	// Project copilot endpoints live under /api/projects (plural —
	// ProjectCopilotController), NOT the /api/project base the `project`
	// domain auto-derives, hence a dedicated domain with an explicit APIPath.
	Register(&Domain{
		Name:        "project-copilot",
		Aliases:     []string{"projectcopilot", "copilot"},
		Description: "Project copilot: health snapshots, AI summary, AI BOM suggestions",
		APIPath:     "/api/projects",
		Actions: []Action{
			{
				Name:        "health-compute",
				Description: "Compute and store a deterministic health snapshot for a project (schedule risk + budget burn; last 12 kept)",
				ToolName:    "UteamupProjectComputeHealth",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/health/compute",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "health",
				Description: "Latest stored health snapshot for a project",
				ToolName:    "UteamupProjectGetHealth",
				RESTPath:    "{projectGuid}/health",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "summary",
				Description: "AI summary of the latest health snapshot (cached per snapshot; a fresh call charges AI quota)",
				ToolName:    "UteamupProjectGenerateSummary",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/summary",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "bom-suggest",
				Description: "Catalog-grounded AI BOM suggestion for a project (review-first — nothing persisted)",
				ToolName:    "UteamupProjectSuggestBom",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/bom/suggest",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					// Always sent (Default "") so the POST carries a JSON body — the
					// backend binds [FromBody] SuggestProjectBomRequestModel.
					{Name: "description", Description: "Extra free-text description of the planned work to ground the suggestion", Default: "", Type: "string"},
				},
			},
			{
				Name:        "bom-apply",
				Description: "Apply reviewed BOM lines to a project (lines come from a JSON file)",
				ToolName:    "UteamupProjectApplyBom",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/bom/apply",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "file", Short: "f", Description: "Path to a JSON file with the lines: [{\"itemType\":\"Part|Tool|Chemical|StockItem\",\"itemGuid\":\"…\",\"quantity\":N}]", Required: true, Type: "string", JSONFile: true, BodyName: "lines"},
				},
			},
		},
	})
}
