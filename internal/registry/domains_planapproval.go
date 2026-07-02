package registry

func init() {
	Register(&Domain{
		Name:        "plan-approval",
		Aliases:     []string{"plan-approvals", "planapproval"},
		Description: "Maker-checker approval queue for gated plan changes",
		APIPath:     "/api/planapproval",
		Actions: []Action{
			{
				Name:        "pending",
				Description: "List pending plan change requests (oldest first)",
				ToolName:    "UteamupPlanApprovalPending",
				RESTPath:    "pending",
			},
			{
				Name:        "approve",
				Description: "Approve a pending request and apply the change (requester cannot self-approve)",
				ToolName:    "UteamupPlanApprovalApprove",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{requestGuid}/approve",
				Args:        []ArgDef{{Name: "requestGuid", Description: "Change request GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "reject",
				Description: "Reject a pending request; the proposed change is discarded",
				ToolName:    "UteamupPlanApprovalReject",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{requestGuid}/reject",
				Args:        []ArgDef{{Name: "requestGuid", Description: "Change request GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
		},
	})
}
