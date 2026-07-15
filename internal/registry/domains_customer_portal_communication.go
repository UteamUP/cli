package registry

func init() {
	portalUserArg := []ArgDef{
		{Name: "portalUserGuid", Description: "Customer portal user external GUID", Required: true, Type: "uuid"},
	}

	Register(&Domain{
		Name:        "customer-message",
		Aliases:     []string{"customer-messages"},
		Description: "Manage public-safe customer portal messages by GUID",
		APIPath:     "/api/customerportal",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List messages for a customer portal user GUID",
				ToolName:    "UteamupCustomerPortalMessageList",
				RESTPath:    "by-guid/{portalUserGuid}/messages",
				Args:        portalUserArg,
			},
			{
				Name:        "create",
				Description: "Send a customer portal message with GUID relationships",
				ToolName:    "UteamupCustomerPortalMessageCreate",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{portalUserGuid}/messages",
				Args:        portalUserArg,
				Flags: []FlagDef{
					{Name: "to-user-guid", BodyName: "toUserGuid", Description: "Recipient user external GUID", Required: true, Type: "uuid"},
					{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Optional workorder GUID", Type: "uuid"},
					{Name: "project-guid", BodyName: "projectGuid", Description: "Optional project GUID", Type: "uuid"},
					{Name: "content", Description: "Message content", Required: true, Type: "string"},
				},
			},
		},
	})

	Register(&Domain{
		Name:        "customer-rating",
		Aliases:     []string{"customer-ratings"},
		Description: "Manage customer job ratings by public GUID",
		APIPath:     "/api/customerportal",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List ratings for a customer portal user GUID",
				ToolName:    "UteamupCustomerPortalJobRatingList",
				RESTPath:    "by-guid/{portalUserGuid}/ratings",
				Args:        portalUserArg,
			},
			{
				Name:        "create",
				Description: "Create a workorder rating using GUID relationships",
				ToolName:    "UteamupCustomerPortalJobRatingCreate",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{portalUserGuid}/ratings",
				Args:        portalUserArg,
				Flags: []FlagDef{
					{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Workorder GUID", Required: true, Type: "uuid"},
					{Name: "rating", Description: "Rating from 1 to 5", Required: true, Type: "int"},
					{Name: "comment", Description: "Optional rating comment", Type: "string"},
				},
			},
		},
	})
}
