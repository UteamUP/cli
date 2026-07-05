package registry

// Custom AI-credit requests CLI surface — a tenant asks for a non-catalog credit amount (e.g. 500,000/month
// billed annually); global admins list, fulfil, or reject them. Mirrors the backend AiCreditRequestController
// (GUID-first per Guidelines/ApiHowToGuidelinesReadme.md):
//
//   submit   POST /api/aicreditrequest                 (body = flags, camelCased; tenant owner)
//   mine     GET  /api/aicreditrequest/mine            (current tenant's requests)
//   pending  GET  /api/aicreditrequest/pending         (global-admin review queue)
//   fulfill  POST /api/aicreditrequest/{guid}/fulfill  (subscribe the tenant to an AI-credit package plan)
//   reject   POST /api/aicreditrequest/{guid}/reject
//
// The CLI calls these REST routes directly (CallREST); the ToolName is the MCP mirror declaration.

func init() {
	Register(&Domain{
		Name:        "aicreditrequest",
		Aliases:     []string{"aicreditrequests", "creditrequest", "creditrequests"},
		Description: "Submit, list, fulfil, and reject custom AI-credit requests",
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
				Name:        "fulfill",
				Description: "Fulfil a pending request by subscribing the tenant to an AI-credit package plan (global-admin only)",
				ToolName:    "UteamupAiCreditRequestFulfill",
				HTTPMethod:  "POST",
				RESTPath:    "{guid}/fulfill",
				Args:        []ArgDef{{Name: "guid", Description: "Request GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "package-plan-guid", BodyName: "packagePlanGuid", Description: "AI-credit package plan GUID to subscribe the tenant to (required)", Required: true, Type: "string"},
					{Name: "billing-cycle", BodyName: "billingCycle", Description: "Billing cycle: monthly or annual (default annual)", Type: "string"},
					{Name: "quantity", Description: "Number of package units (default 1)", Type: "int"},
					{Name: "resolution-note", BodyName: "resolutionNote", Description: "Optional note captured at fulfilment", Type: "string"},
				},
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
