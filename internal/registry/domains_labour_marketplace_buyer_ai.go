package registry

func init() {
	Register(&Domain{
		Name:        "labour-ai-buyer",
		Aliases:     []string{"job-draft-ai"},
		Description: "Create editable buyer labour-job drafts with explicit AI-credit costs",
		APIPath:     "/api/labour-marketplace/ai/buyer-job-draft",
		Actions: []Action{
			{
				Name:        "cost",
				Description: "Preview the live AI-credit cost and your remaining balance",
				ToolName:    "UteamupLabourMarketplaceBuyerJobDraftCost",
				RESTPath:    "cost",
				HTTPMethod:  "GET",
			},
			{
				Name:        "create",
				Description: "Generate an editable draft; this never creates or publishes a job",
				ToolName:    "UteamupLabourMarketplaceBuyerJobDraftCreate",
				Flags: []FlagDef{
					{Name: "description", Short: "d", Description: "Workers, scope, and outcome needed", Required: true, Type: "string"},
					{Name: "currency-code", Description: "Three-letter currency code", Type: "string", Default: "USD"},
					{Name: "project-guid", Description: "Optional permitted project GUID", Type: "string"},
					{Name: "workorder-guid", Description: "Optional permitted workorder GUID", Type: "string"},
					{Name: "language", Description: "Output language", Type: "string", Default: "English"},
				},
			},
		},
	})
}
