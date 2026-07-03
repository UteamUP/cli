package registry

func init() {
	// Project copilot endpoints live under /api/projects (plural —
	// ProjectCopilotController), NOT the /api/project base the `project`
	// domain auto-derives, hence a dedicated domain with an explicit APIPath.
	Register(&Domain{
		Name:        "project-copilot",
		Aliases:     []string{"projectcopilot", "copilot"},
		Description: "Project copilot: health snapshots, AI summary, AI BOM/WBS/prioritization/risk suggestions, lessons learned, AI image report",
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
			{
				Name:        "image-report",
				Description: "AI close-out report synthesized from a project's analyzed photos (one synthesis call; per-image analysis charged separately)",
				ToolName:    "UteamupProjectGenerateImageReport",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/image-report",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
			// AI Planning Suite (phase B4). Suggest actions persist nothing and
			// charge AI quota (outcomes ride the 200 body); the matching apply
			// actions are free and take the reviewed suggestion via a JSON file,
			// following the bom-apply pattern.
			{
				Name:        "wbs-suggest",
				Description: "AI-suggested work breakdown structure (stages + tasks) for a project (review-first — nothing persisted)",
				ToolName:    "UteamupProjectSuggestWbs",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/wbs/suggest",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "wbs-apply",
				Description: "Create the reviewed WBS stages + workorders on a project (stages come from a JSON file; free)",
				ToolName:    "UteamupProjectApplyWbs",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/wbs/apply",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "file", Short: "f", Description: "Path to a JSON file with the stages: [{\"name\":\"…\",\"order\":1,\"gateCriteria\":null,\"startOffsetDays\":0,\"durationDays\":5,\"tasks\":[{\"title\":\"…\",\"description\":null,\"estimateMinutes\":60,\"suggestedOrder\":1}]}]", Required: true, Type: "string", JSONFile: true, BodyName: "stages"},
				},
			},
			{
				Name:        "estimate",
				Description: "AI-estimated duration + cost for a project or template, from its workorders/stages and similar completed projects (review-first — nothing persisted)",
				ToolName:    "UteamupProjectEstimate",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/estimate",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project or template GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "estimate-apply",
				Description: "Write a reviewed estimate onto a project — estimated cost and/or the end date (free)",
				ToolName:    "UteamupProjectApplyEstimate",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/estimate/apply",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "duration-days", Description: "Estimated duration in days (sets end date = start + days when --apply-duration)", Type: "int", BodyName: "estimatedDurationDays"},
					{Name: "cost", Description: "Estimated cost (sets Estimated Cost when --apply-cost)", Type: "float", BodyName: "estimatedCost"},
					{Name: "apply-duration", Description: "Apply the duration to the project end date", Type: "bool", Default: true, BodyName: "applyDuration"},
					{Name: "apply-cost", Description: "Apply the cost to the project estimated cost", Type: "bool", Default: true, BodyName: "applyCost"},
				},
			},
			{
				Name:        "prioritize-suggest",
				Description: "AI-suggested priority ranking of a project's open workorders (review-first — nothing persisted)",
				ToolName:    "UteamupProjectSuggestPrioritization",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/prioritize/suggest",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "prioritize-apply",
				Description: "Apply the reviewed priorities to a project's workorders (items come from a JSON file; free)",
				ToolName:    "UteamupProjectApplyPrioritization",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/prioritize/apply",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "file", Short: "f", Description: "Path to a JSON file with the items: [{\"workorderGuid\":\"…\",\"priority\":1-5}]", Required: true, Type: "string", JSONFile: true, BodyName: "items"},
				},
			},
			{
				Name:        "risks-suggest",
				Description: "AI-suggested risk register entries for a project (review-first — nothing persisted)",
				ToolName:    "UteamupProjectSuggestRisks",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/risks/suggest",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "risks-apply",
				Description: "Create the reviewed risk suggestions in the project risk register, marked as AI-suggested (risks come from a JSON file; free)",
				ToolName:    "UteamupProjectApplyRisks",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/risks/apply",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "file", Short: "f", Description: "Path to a JSON file with the risks: [{\"title\":\"…\",\"description\":null,\"category\":\"Technical|Schedule|Cost|Resource|Safety|External|Other\",\"probability\":1-5,\"impact\":1-5,\"mitigationPlan\":null}]", Required: true, Type: "string", JSONFile: true, BodyName: "risks"},
				},
			},
			{
				Name:        "lessons-learned",
				Description: "AI lessons-learned summary for a COMPLETED project (persisted on the project; regenerating re-charges and overwrites)",
				ToolName:    "UteamupProjectGenerateLessonsLearned",
				HTTPMethod:  "POST",
				RESTPath:    "{projectGuid}/lessons-learned",
				Args:        []ArgDef{{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"}},
			},
		},
	})
}
