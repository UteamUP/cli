package registry

func init() {
	permissionFlags := []FlagDef{
		{Name: "is-active", BodyName: "isActive", Description: "Whether the portal user is active", Default: true, Type: "bool"},
		{Name: "can-view-projects", BodyName: "canViewProjects", Description: "Allow project visibility", Default: true, Type: "bool"},
		{Name: "can-request-work", BodyName: "canRequestWork", Description: "Allow work requests", Default: true, Type: "bool"},
		{Name: "can-approve-workorders", BodyName: "canApproveWorkorders", Description: "Allow workorder approval", Default: false, Type: "bool"},
		{Name: "can-track-fleet", BodyName: "canTrackFleet", Description: "Allow fleet tracking", Default: false, Type: "bool"},
		{Name: "can-send-messages", BodyName: "canSendMessages", Description: "Allow portal messaging", Default: true, Type: "bool"},
		{Name: "can-rate-jobs", BodyName: "canRateJobs", Description: "Allow job ratings", Default: true, Type: "bool"},
	}

	createFlags := append([]FlagDef{
		{Name: "customer-external-guid", BodyName: "customerExternalGuid", Description: "Customer external GUID", Required: true, Type: "string"},
		{Name: "email", Description: "Portal user email address", Required: true, Type: "string"},
		{Name: "password", Description: "Optional initial password", Type: "string", Sensitive: true},
	}, permissionFlags...)

	Register(&Domain{
		Name:        "customer-portal-user",
		Aliases:     []string{"customer-portal", "cpu"},
		Description: "Administer customer portal users by external GUID",
		APIPath:     "/api/customerportalusers",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List customer portal users",
				ToolName:    "UteamupCustomerPortalUserList",
			},
			{
				Name:        "get",
				Description: "Get a customer portal user by external GUID",
				ToolName:    "UteamupCustomerPortalUserGet",
				RESTPath:    "by-guid/{userExternalGuid}",
				Args: []ArgDef{
					{Name: "userExternalGuid", Description: "Customer portal user external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create a customer portal user for a customer GUID",
				ToolName:    "UteamupCustomerPortalUserCreate",
				HTTPMethod:  "POST",
				Flags:       createFlags,
			},
			{
				Name:        "update",
				Description: "Update a customer portal user by external GUID",
				ToolName:    "UteamupCustomerPortalUserUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{userExternalGuid}",
				Args: []ArgDef{
					{Name: "userExternalGuid", Description: "Customer portal user external GUID", Required: true, Type: "uuid"},
				},
				Flags: permissionFlags,
			},
			{
				Name:        "delete",
				Description: "Delete a customer portal user by external GUID",
				ToolName:    "UteamupCustomerPortalUserDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{userExternalGuid}",
				Args: []ArgDef{
					{Name: "userExternalGuid", Description: "Customer portal user external GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
