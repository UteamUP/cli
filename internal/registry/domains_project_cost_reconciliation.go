package registry

// Reviewed cost provenance coordinates existing ledger and stock evidence without inventory writes.
func init() {
	Register(&Domain{
		Name:        "project-deliverable-cost",
		Description: "Read and edit currency-aware deliverable cost plans",
		APIPath:     "/api/projects",
		Actions: []Action{
			{
				Name: "get-plan", Description: "Read direct stored inputs and updatedAt before editing; not subtree totals",
				ToolName: "UteamupProjectDeliverableCostPlanGet", RESTPath: "{projectGuid}/outputitems/{deliverableGuid}/costs/plan",
				Args: projectResourceArguments("deliverableGuid", "Deliverable GUID"),
			},
			{
				Name: "summary", Description: "Cost summary for the deliverable subtree",
				ToolName: "UteamupProjectCostSummary", RESTPath: "{projectGuid}/outputitems/{deliverableGuid}/costs",
				Args: projectResourceArguments("deliverableGuid", "Deliverable GUID"),
			},
			{
				Name: "plan", Description: "Update supplied plan fields with expectedUpdatedAt; parent amounts include descendants",
				ToolName: "UteamupProjectDeliverableCostPlanUpdate", HTTPMethod: "PUT",
				RESTPath: "{projectGuid}/outputitems/{deliverableGuid}/costs/plan",
				Args:     projectResourceArguments("deliverableGuid", "Deliverable GUID"), Flags: []FlagDef{projectRequestFile()},
			},
		},
	})
	Register(&Domain{
		Name:        "project-cost-reconciliation",
		Description: "Review overlapping project receipt and stock costs with immutable provenance",
		APIPath:     "/api/projects",
		Actions: []Action{
			{
				Name: "preview", Description: "Show exact source costs, receipt candidates and review fingerprints",
				ToolName: "UteamupProjectCostReconciliation", RESTPath: "{projectGuid}/cost-records/reconciliation",
				Args: projectGUIDArgument,
			},
			{
				Name: "history", Description: "Show current, stale and superseded reviews for one stock source",
				ToolName: "UteamupProjectCostReconciliationHistory", RESTPath: "{projectGuid}/cost-records/reconciliation/history",
				Args: []ArgDef{
					{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "string"},
					{Name: "sourceGuid", Description: "Stock transaction source GUID", Required: true, Type: "string", QueryName: "sourceGuid"},
				},
			},
			{
				Name: "review", Description: "Record Independent or CoveredByReceipt evidence; reuse the request GUID and payload for retries",
				ToolName: "UteamupProjectCostReconcile", HTTPMethod: "POST", RESTPath: "{projectGuid}/cost-records/reconciliation",
				Args: projectGUIDArgument, Flags: []FlagDef{projectRequestFile()},
			},
		},
	})
}
