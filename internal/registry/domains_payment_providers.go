package registry

func init() {
	Register(&Domain{
		Name:        "payment-providers",
		Aliases:     []string{"pp", "providers"},
		Description: "Inspect payment providers (global admin; Kling is the only customer rail)",
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
		},
	})
}
