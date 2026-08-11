package registry

func init() {
	Register(&Domain{
		Name:        "payment-providers",
		Aliases:     []string{"pp", "providers"},
		Description: "Inspect and manage payment providers (global admin; Kling is the only live rail)",
		APIPath:     "/api/globaladmin",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List configured payment providers with status (active rail, parked rails, disabled reasons)",
				ToolName:    "UteamupPaymentProvidersList",
				HTTPMethod:  "GET",
				RESTPath:    "payment-providers",
			},
			{
				Name:        "health-check",
				Description: "Health-check a payment provider by name (e.g. kling)",
				ToolName:    "UteamupPaymentProviderHealthCheck",
				HTTPMethod:  "POST",
				RESTPath:    "payment-providers/{providerName}/health-check",
				Args:        []ArgDef{{Name: "providerName", Description: "Provider name, e.g. kling", Required: true, Type: "string"}},
			},
			{
				Name:        "set-active",
				Description: "Set the active payment provider (global admin only; parked rails are rejected by the backend guards)",
				ToolName:    "UteamupPaymentProviderSetActive",
				HTTPMethod:  "POST",
				RESTPath:    "payment-providers/{providerName}/set-active",
				Args:        []ArgDef{{Name: "providerName", Description: "Provider name to activate, e.g. kling", Required: true, Type: "string"}},
			},
		},
	})
}
