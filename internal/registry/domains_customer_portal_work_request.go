package registry

func init() {
	Register(&Domain{
		Name:        "customer-portal-work-request",
		Aliases:     []string{"customer-work-request", "cpwr"},
		Description: "Manage customer portal work requests by external GUID",
		APIPath:     "/api/customerportalworkrequests",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List customer portal work requests",
				ToolName:    "UteamupCustomerPortalWorkRequestList",
				Flags: []FlagDef{
					{Name: "portal-user-guid", BodyName: "portalUserGuid", Description: "Optional portal user external GUID filter", Type: "uuid"},
				},
			},
			{
				Name:        "get",
				Description: "Get a customer portal work request by external GUID",
				ToolName:    "UteamupCustomerPortalWorkRequestGet",
				RESTPath:    "by-guid/{workRequestGuid}",
				Args: []ArgDef{
					{Name: "workRequestGuid", Description: "Customer portal work request external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create a customer portal work request using public GUIDs",
				ToolName:    "UteamupCustomerPortalWorkRequestCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "portal-user-guid", BodyName: "customerPortalUserExternalGuid", Description: "Portal user external GUID", Required: true, Type: "uuid"},
					{Name: "customer-guid", BodyName: "customerExternalGuid", Description: "Customer external GUID", Required: true, Type: "uuid"},
					{Name: "description", Description: "Requested work description", Required: true, Type: "string"},
					{Name: "priority-hint", BodyName: "priorityHint", Description: "Optional priority hint from 0 to 5", Type: "int"},
				},
			},
			{
				Name:        "status-update",
				Description: "Update a customer portal work request status by external GUID",
				ToolName:    "UteamupCustomerPortalWorkRequestStatusUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{workRequestGuid}/status",
				Args: []ArgDef{
					{Name: "workRequestGuid", Description: "Customer portal work request external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "status", Description: "New status value", Required: true, Type: "int"},
					{Name: "review-notes", BodyName: "reviewNotes", Description: "Optional review notes", Type: "string"},
				},
			},
		},
	})
}
