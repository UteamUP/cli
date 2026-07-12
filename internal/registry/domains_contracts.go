package registry

func init() {
	Register(&Domain{Name: "contract", Aliases: []string{"contracts"}, Description: "Manage contracts", Actions: crudActions("Contract")})
	Register(&Domain{Name: "contractor", Aliases: []string{"contractors"}, Description: "Manage contractor profiles", Actions: crudActions("ContractorProfile")})
	Register(&Domain{Name: "contractor-workorder", Description: "Manage contractor work orders", Actions: crudActions("ContractorWorkOrder")})
	Register(&Domain{
		Name:        "labour-rate",
		Description: "Manage GUID-first labour rate rules and schedule modifiers",
		APIPath:     "/api/labourrate",
		Actions: []Action{
			{
				Name:        "list-rules",
				Description: "List labour rate rules",
				ToolName:    "UteamupLabourRateGetRules",
				RESTPath:    "rules",
				HTTPMethod:  "GET",
			},
			{
				Name:        "create-rule",
				Description: "Create a labour rate rule from JSON",
				ToolName:    "UteamupLabourRateCreateRule",
				RESTPath:    "rules",
				HTTPMethod:  "POST",
				Flags:       []FlagDef{jsonFlag()},
			},
			{
				Name:        "update-rule",
				Description: "Update a labour rate rule by public GUID",
				ToolName:    "UteamupLabourRateUpdateRule",
				RESTPath:    "rules/{ruleGuid}",
				HTTPMethod:  "PUT",
				Args: []ArgDef{{
					Name:        "ruleGuid",
					Description: "Labour rate rule public GUID",
					Required:    true,
					Type:        "string",
				}},
				Flags: []FlagDef{jsonFlag()},
			},
			{
				Name:        "delete-rule",
				Description: "Delete a labour rate rule by public GUID",
				ToolName:    "UteamupLabourRateDeleteRule",
				RESTPath:    "rules/{ruleGuid}",
				HTTPMethod:  "DELETE",
				Args: []ArgDef{{
					Name:        "ruleGuid",
					Description: "Labour rate rule public GUID",
					Required:    true,
					Type:        "string",
				}},
			},
			{
				Name:        "list-modifiers",
				Description: "List after-hours, call-out, holiday, and other rate modifiers",
				ToolName:    "UteamupLabourRateGetModifiers",
				RESTPath:    "modifiers",
				HTTPMethod:  "GET",
			},
			{
				Name:        "create-modifier",
				Description: "Create a labour rate modifier from JSON",
				ToolName:    "UteamupLabourRateCreateModifier",
				RESTPath:    "modifiers",
				HTTPMethod:  "POST",
				Flags:       []FlagDef{jsonFlag()},
			},
			{
				Name:        "update-modifier",
				Description: "Update a labour rate modifier by public GUID",
				ToolName:    "UteamupLabourRateUpdateModifier",
				RESTPath:    "modifiers/{modifierGuid}",
				HTTPMethod:  "PUT",
				Args: []ArgDef{{
					Name:        "modifierGuid",
					Description: "Labour rate modifier public GUID",
					Required:    true,
					Type:        "string",
				}},
				Flags: []FlagDef{jsonFlag()},
			},
			{
				Name:        "delete-modifier",
				Description: "Delete a labour rate modifier by public GUID",
				ToolName:    "UteamupLabourRateDeleteModifier",
				RESTPath:    "modifiers/{modifierGuid}",
				HTTPMethod:  "DELETE",
				Args: []ArgDef{{
					Name:        "modifierGuid",
					Description: "Labour rate modifier public GUID",
					Required:    true,
					Type:        "string",
				}},
			},
		},
	})
	Register(&Domain{Name: "rental-rate", Description: "Manage rental rates", Actions: crudActions("RentalRate")})
	Register(&Domain{Name: "warranty", Aliases: []string{"warranties"}, Description: "Manage warranties", Actions: crudActions("Warranty")})
	Register(&Domain{Name: "commission", Aliases: []string{"commissions"}, Description: "Manage commissions", Actions: crudActions("Commission")})
}
