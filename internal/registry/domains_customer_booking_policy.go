package registry

func init() {
	Register(&Domain{
		Name:        "customer-booking-policy",
		Description: "Customer booking policy configuration lookups",
		APIPath:     "/api/customerbookingpolicies",
		Actions: []Action{
			{
				Name:        "service-types",
				Description: "Search tenant service-type names and public GUIDs for a booking policy",
				ToolName:    "UteamupCustomerBookingServiceTypeList",
				HTTPMethod:  "GET",
				RESTPath:    "service-types",
				Flags: []FlagDef{
					{Name: "search", BodyName: "search", Description: "Service-type name search, maximum 200 characters", Type: "string"},
					{Name: "service-type-guid", BodyName: "serviceTypeGuid", Description: "Resolve one service-type public GUID in the current tenant", Type: "string"},
					{Name: "page", BodyName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", BodyName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
		},
	})
}
