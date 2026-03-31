package registry

func init() {
	Register(&Domain{
		Name:        "ai-provider",
		Aliases:     []string{"aip", "byok"},
		Description: "Manage BYOK AI provider configuration (Bring Your Own Key)",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get the current tenant's BYOK AI provider configuration",
				ToolName:    "UteamupGetAIProviderConfig",
			},
			{
				Name:        "update",
				Description: "Update the BYOK AI provider configuration",
				ToolName:    "UteamupUpdateAIProviderConfig",
				Flags: []FlagDef{
					{Name: "provider", Short: "p", Description: "Provider type: AzureAIFoundry, GoogleGemini, OpenAI, AnthropicClaude, OpenAICompatible", Required: true, Type: "string"},
					{Name: "api-key", Short: "k", Description: "API key for the provider", Type: "string"},
					{Name: "base-url", Short: "u", Description: "Provider endpoint URL", Type: "string"},
					{Name: "model", Short: "m", Description: "Model name (e.g., gpt-4o, claude-sonnet-4-20250514)", Required: true, Type: "string"},
					{Name: "display-name", Short: "d", Description: "Friendly display name", Type: "string"},
				},
			},
			{
				Name:        "test-connection",
				Description: "Test connection to an AI provider",
				ToolName:    "UteamupTestAIProviderConnection",
				Flags: []FlagDef{
					{Name: "provider", Short: "p", Description: "Provider type", Required: true, Type: "string"},
					{Name: "api-key", Short: "k", Description: "API key", Required: true, Type: "string"},
					{Name: "base-url", Short: "u", Description: "Provider endpoint URL", Type: "string"},
					{Name: "model", Short: "m", Description: "Model name", Required: true, Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Remove BYOK AI provider configuration (revert to internal)",
				ToolName:    "UteamupDeleteAIProviderConfig",
			},
		},
	})
}
