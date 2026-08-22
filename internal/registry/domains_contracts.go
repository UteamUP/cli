package registry

func init() {
	// APIPath is explicit because buildRESTPath would otherwise derive "/api/contract" from the
	// singular domain name, while ContractsController is [Route("api/[controller]")] and so
	// serves the PLURAL "/api/contracts" — every command 404'd.
	Register(&Domain{Name: "contract", Aliases: []string{"contracts"}, APIPath: "/api/contracts", Description: "Manage contracts", Actions: crudActions("Contract")})
	Register(&Domain{
		Name:        "contractor",
		Aliases:     []string{"contractors"},
		Description: "View contractor profiles by public GUID",
		APIPath:     "/api/v1/contractor",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List contractor profiles",
				ToolName:    "UteamupContractorProfileList",
				MCPOnly:     true,
				Flags:       paginationFlags(),
			},
			{
				Name:        "get",
				Description: "Get a contractor profile by public GUID",
				ToolName:    "UteamupContractorProfileGet",
				RESTPath:    "profile/by-guid/{profileGuid}",
				HTTPMethod:  "GET",
				Args: []ArgDef{{
					Name:        "profileGuid",
					Description: "Contractor profile public GUID",
					Required:    true,
					Type:        "uuid",
				}},
			},
		},
	})
	Register(&Domain{
		Name:        "contractor-workorder",
		Description: "Assign contractors and review bids using public GUIDs",
		APIPath:     "/api/v1/contractor/assignments",
		Actions: []Action{
			{
				Name:        "assign",
				Description: "Assign a contractor from a GUID-only JSON model",
				ToolName:    "UteamupContractorAssign",
				MCPOnly:     true,
				Flags: []FlagDef{{
					Name:        "from-json",
					BodyName:    "model",
					Description: "JSON file containing workOrderGuid and contractorProfileGuid",
					Required:    true,
					Type:        "string",
					JSONFile:    true,
				}},
			},
			{
				Name:        "assignments",
				Description: "List assignments for a work order public GUID",
				ToolName:    "UteamupContractorAssignmentList",
				RESTPath:    "workorder/{workOrderGuid}",
				HTTPMethod:  "GET",
				Args: []ArgDef{{
					Name: "workOrderGuid", Description: "Work order public GUID", Required: true, Type: "uuid",
				}},
			},
			{
				Name:        "bids",
				Description: "List bids for a work order public GUID",
				ToolName:    "UteamupContractorBidList",
				RESTPath:    "workorder/{workOrderGuid}/bids",
				HTTPMethod:  "GET",
				Args: []ArgDef{{
					Name: "workOrderGuid", Description: "Work order public GUID", Required: true, Type: "uuid",
				}},
			},
			{
				Name:        "accept-bid",
				Description: "Accept a contractor bid by public GUID",
				ToolName:    "UteamupContractorBidAccept",
				RESTPath:    "bids/{bidGuid}/accept",
				HTTPMethod:  "POST",
				Args: []ArgDef{{
					Name: "bidGuid", Description: "Contractor bid public GUID", Required: true, Type: "uuid",
				}},
			},
		},
	})
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
	// Same plural-controller mismatch as "contract" above: WarrantiesController serves
	// "/api/warranties", not the derived "/api/warranty".
	Register(&Domain{Name: "warranty", Aliases: []string{"warranties"}, APIPath: "/api/warranties", Description: "Manage warranties", Actions: crudActions("Warranty")})
	Register(&Domain{Name: "commission", Aliases: []string{"commissions"}, Description: "Manage commissions", Actions: crudActions("Commission")})
}
