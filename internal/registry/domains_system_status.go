package registry

func init() {
	Register(&Domain{
		Name:        "system-status",
		Aliases:     []string{"status", "ss"},
		Description: "Live status of every external service UteamUP depends on (global admin)",
		APIPath:     "/api/globaladmin",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Check AI providers, grounded search, database, storage, Key Vault, payments, email, sign-in and Redis (cached 60 s)",
				ToolName:    "UteamupExternalServiceStatus",
				HTTPMethod:  "GET",
				RESTPath:    "external-services/status",
				Flags: []FlagDef{
					{Name: "refresh", QueryName: "refresh", Description: "Run every check again instead of using results from the last minute (calls paid AI APIs)", Type: "bool"},
				},
			},
		},
	})
}
