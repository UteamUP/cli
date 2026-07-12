package registry

func init() {
	Register(&Domain{
		Name:        "labour-marketplace-workspace",
		Aliases:     []string{"labour-workspace", "offers-contracts"},
		Description: "View the authenticated user's private buyer jobs, provider applications, agreements, and dispatches",
		APIPath:     "/api/labour-marketplace",
		Actions: []Action{
			{
				Name:        "me",
				Description: "Show the current user's combined labour marketplace workspace",
				ToolName:    "UteamupLabourMarketplaceWorkspaceMe",
				RESTPath:    "workspace/me",
				HTTPMethod:  "GET",
			},
			{
				Name:        "timesheets",
				Description: "List rate-derived timesheets and dual-approval state for an accessible labour agreement",
				ToolName:    "UteamupLabourAgreementTimesheets",
				RESTPath:    "agreements/{agreementGuid}/timesheets",
				HTTPMethod:  "GET",
				Args: []ArgDef{{
					Name:        "agreementGuid",
					Description: "Labour agreement GUID",
					Required:    true,
					Type:        "string",
				}},
			},
			{
				Name:        "replace-worker",
				Description: "Replace a proposed worker and create a new immutable roster revision",
				ToolName:    "UteamupLabourWorkerDispatchReplace",
				RESTPath:    "dispatches/{dispatchGuid}/replacement",
				HTTPMethod:  "POST",
				Args: []ArgDef{{
					Name:        "dispatchGuid",
					Description: "Dispatch GUID being replaced",
					Required:    true,
					Type:        "string",
				}},
				Flags: []FlagDef{
					{Name: "replacement-workforce-membership-guid", Description: "Registered replacement workforce membership GUID", Type: "string", Required: true},
					{Name: "reason", Description: "Auditable reason for the roster change", Type: "string", Required: true},
				},
			},
		},
	})
}
