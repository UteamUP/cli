package registry

func init() {
	Register(&Domain{
		Name:        "labour-marketplace-workspace",
		Aliases:     []string{"labour-workspace", "offers-contracts"},
		Description: "View the authenticated user's private buyer jobs, provider applications, agreements, and dispatches",
		APIPath:     "/api/labour-marketplace",
		Actions: []Action{
			{
				Name: "accept-dispatch", Description: "Explicitly accept or decline only the signed-in worker's own current dispatch",
				ToolName: "UteamupLabourDispatchAccept", HTTPMethod: "POST", RESTPath: "dispatches/{dispatchGuid}/acceptance", RejectExtraArgs: true,
				Args: []ArgDef{{Name: "dispatchGuid", Description: "Personal dispatch public GUID", Required: true, Type: "non-empty-uuid"}},
				Flags: []FlagDef{
					{Name: "accept", Description: "Explicit decision: --accept=true or --accept=false", Type: "bool", Required: true},
					{Name: "reason", Description: "Optional reason for the personal decision", Type: "string"},
				},
			},
			{
				Name: "create-my-timesheet", Description: "Submit only the signed-in worker's accepted dispatch entries using the original stable request key",
				ToolName: "UteamupLabourTimesheetCreateMine", HTTPMethod: "POST", RESTPath: "agreements/{agreementGuid}/my-timesheets", RejectExtraArgs: true,
				Args:  []ArgDef{{Name: "agreementGuid", Description: "Personal agreement public GUID", Required: true, Type: "non-empty-uuid"}},
				Flags: []FlagDef{{Name: "from-json", Description: "Canonical personal time JSON with nonempty idempotencyKey, periodStart, periodEnd and own entries; retain unchanged for retries", Type: "string", Required: true, RootJSONObjectFile: true}},
			},
			{
				Name: "my-timesheets", Description: "Read only the signed-in worker's timesheets for one agreement",
				ToolName: "UteamupLabourTimesheetsMine", HTTPMethod: "GET", RESTPath: "agreements/{agreementGuid}/my-timesheets", RejectExtraArgs: true,
				Args: []ArgDef{{Name: "agreementGuid", Description: "Personal agreement public GUID", Required: true, Type: "non-empty-uuid"}},
			},
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
