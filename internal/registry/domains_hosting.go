package registry

func init() {
	Register(&Domain{
		Name:        "hosting",
		Description: "Saved UteamUP Azure hosting costs, resource sizes and prices (configured global admin)",
		APIPath:     "/api/globaladmin/hosting",
		Actions: []Action{{
			Name:        "overview",
			Description: "Read stored snapshots without refreshing Azure; missing invoices mean payment coverage is unverified",
			ToolName:    "UteamupHostingCosts",
			HTTPMethod:  "GET",
			Flags: []FlagDef{{
				Name: "month", QueryName: "month", Type: "string",
				Description: "First day of a month in the last 12 months (yyyy-MM-dd); defaults to this month",
			}},
		}},
	})
}
