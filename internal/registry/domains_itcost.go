package registry

func init() {
	Register(&Domain{
		Name:        "itcost",
		Aliases:     []string{"cloud-cost"},
		Description: "Read cloud spend (Azure, AWS, Google Cloud), budgets, ways to save and licences for the selected tenant",
		// Mirrors the IT cost MCP tools. Cloud billing accounts and their secrets are set up in the web app
		// (Tenant Management → IT & Cloud) only, so there are no account commands. Budget and unusual-spend
		// alerts are IT alerts: `uteamup itmanagement alerts --source costAnomaly`.
		APIPath: "/api/itmanagement/cost",
		Actions: []Action{
			{
				Name:        "summary",
				Description: "Spent this month, expected by month end, last month and savings available",
				ToolName:    "UteamupITManagementCostSummary",
				HTTPMethod:  "GET",
				RESTPath:    "summary",
			},
			{
				Name:        "breakdown",
				Description: "Where the money goes: top items for one grouping over a date range (at most 400 days)",
				ToolName:    "UteamupITManagementCostBreakdown",
				HTTPMethod:  "GET",
				RESTPath:    "breakdown",
				Flags: []FlagDef{
					{Name: "group-by", BodyName: "groupBy", Description: "service, account, resource, tag, businessService, region or resourceGroup", Type: "string", Default: "service"},
					{Name: "from", Description: "First day, inclusive (YYYY-MM-DD)", Type: "string"},
					{Name: "to", Description: "Last day, inclusive (YYYY-MM-DD)", Type: "string"},
					{Name: "tag-key", BodyName: "tagKey", Description: "Tag name, required with --group-by tag", Type: "string"},
					{Name: "amortized", Description: "Spread up-front purchases over the days they cover", Type: "bool"},
					{Name: "account-guid", BodyName: "accountGuid", Description: "Only this cloud billing account", Type: "uuid"},
					{Name: "top", Description: "Items to return (1-50)", Type: "int", Default: 10},
				},
			},
			{
				Name:        "budgets",
				Description: "List cloud budgets with spend so far and the month-end forecast",
				ToolName:    "UteamupITManagementCostBudgetsList",
				HTTPMethod:  "GET",
				RESTPath:    "budgets",
			},
			{
				Name:        "savings",
				Description: "List ways to save, largest first; estimated until measured after the work order",
				ToolName:    "UteamupITManagementCostRecommendationsList",
				HTTPMethod:  "GET",
				RESTPath:    "recommendations",
				Flags: []FlagDef{
					{Name: "status", BodyName: "statuses", Description: "open, planned, inProgress, implemented, dismissed or expired (repeatable)", Type: "stringSlice"},
					{Name: "limit", BodyName: "take", Description: "Items to return (1-200)", Type: "int", Default: 50},
				},
			},
			{
				Name:        "saving-workorder",
				Description: "Create a work order for a saving from an approved template (needs CloudCost.Recommendation.Manage and WorkOrder.Create)",
				ToolName:    "UteamupITManagementCostRecommendationCreateWorkorder",
				HTTPMethod:  "POST",
				RESTPath:    "recommendations/{recommendationGuid}/create-workorder",
				Args: []ArgDef{
					{Name: "recommendation-guid", BodyName: "recommendationGuid", Description: "Saving GUID", Type: "non-empty-uuid", Required: true},
				},
				Flags: []FlagDef{
					{Name: "template-guid", BodyName: "workorderTemplateGuid", Description: "Approved work order template GUID", Type: "non-empty-uuid", Required: true},
				},
			},
			{
				Name:        "licences",
				Description: "Software licences: seats bought and used, cost per seat, renewals within 90 days (needs ContractCost.View)",
				ToolName:    "UteamupITManagementLicencesList",
				HTTPMethod:  "GET",
				RESTPath:    "licences",
			},
		},
	})
}
