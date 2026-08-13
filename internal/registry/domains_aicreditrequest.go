package registry

// Custom AI-credit requests CLI surface — a tenant asks for a non-catalog credit amount (e.g. 500,000/month
// billed annually); global admins list or reject them. Paid fulfilment is owned by the provider-neutral billing
// order and settlement flow. Mirrors the backend AiCreditRequestController
// (GUID-first per Guidelines/ApiGuidelines.md):
//
//   submit   POST /api/aicreditrequest                 (body = flags, camelCased; tenant owner)
//   mine     GET  /api/aicreditrequest/mine            (current tenant's requests)
//   pending  GET  /api/aicreditrequest/pending         (global-admin review queue)
//   reject   POST /api/aicreditrequest/{guid}/reject
//
// The CLI calls these REST routes directly (CallREST); the ToolName is the MCP mirror declaration.

func init() {
	Register(&Domain{
		Name:        "aicreditrequest",
		Aliases:     []string{"aicreditrequests", "creditrequest", "creditrequests"},
		Description: "Submit, list, and reject custom AI-credit requests",
		APIPath:     "/api/aicreditrequest",
		Actions: []Action{
			{
				Name:        "submit",
				Description: "Submit a custom AI-credit request for the current tenant (tenant owner).",
				ToolName:    "UteamupAiCreditRequestSubmit",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "requested-monthly-credits", BodyName: "requestedMonthlyCredits", Description: "Credits wanted per month (required)", Required: true, Type: "int"},
					{Name: "billing-cycle", BodyName: "billingCycle", Description: "Billing cycle: monthly or annual (default annual)", Type: "string"},
					{Name: "note", Description: "Optional context (use case, timeline, budget)", Type: "string"},
				},
			},
			{
				Name:        "mine",
				Description: "List the current tenant's custom AI-credit requests",
				ToolName:    "UteamupAiCreditRequestMine",
				RESTPath:    "mine",
			},
			{
				Name:        "pending",
				Description: "List pending custom AI-credit requests across all tenants (global-admin only)",
				ToolName:    "UteamupAiCreditRequestPending",
				RESTPath:    "pending",
			},
			{
				Name:        "reject",
				Description: "Reject a pending request with a reason (global-admin only)",
				ToolName:    "UteamupAiCreditRequestReject",
				HTTPMethod:  "POST",
				RESTPath:    "{guid}/reject",
				Args:        []ArgDef{{Name: "guid", Description: "Request GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "reason", Description: "Reason for rejecting the request (required)", Required: true, Type: "string"},
				},
			},
		},
	})
}
