package registry

func init() {
	Register(&Domain{
		Name:        "hosting",
		Description: "Saved UteamUP Azure hosting costs, resource sizes and prices (configured global admin)",
		APIPath:     "/api/globaladmin/hosting",
		Actions: []Action{{
			Name:              "overview",
			Description:       "Read stored snapshots without refreshing Azure; missing invoices mean payment coverage is unverified",
			ToolName:          "UteamupHostingCosts",
			HTTPMethod:        "GET",
			UseDomainBasePath: true,
			Flags: []FlagDef{{
				Name: "month", QueryName: "month", Type: "string",
				Description: "First day of a month in the last 12 months (yyyy-MM-dd); defaults to this month",
			}},
		}, {
			Name:        "retry-permissions",
			Description: "Retry 401/403 collection after Azure access repair; enforces cooldown and active throttling (configured Global Admin only)",
			ToolName:    "UteamupHostingRetryPermissions",
			HTTPMethod:  "POST",
			RESTPath:    "retry-permissions",
		}, {
			Name:        "refresh",
			Description: "Fetch and save new Azure evidence; respects fifteen-minute cooldown and provider backoff (configured Global Admin only)",
			ToolName:    "UteamupHostingRefresh",
			HTTPMethod:  "POST",
			RESTPath:    "refresh",
			Flags: []FlagDef{{
				Name: "month", QueryName: "month", Type: "string",
				Description: "Optional first day of a month in the last twelve months (yyyy-MM-dd)",
			}},
		}},
	})
}
