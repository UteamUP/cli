package registry

func init() {
	Register(&Domain{
		Name:        "tenant",
		Aliases:     []string{"tenants"},
		Description: "Manage tenants (organizations)",
		Actions: append(listGetActions("Tenant"), []Action{
			{
				Name:        "industry-profiles-get",
				Description: "Get the active tenant's primary and additional industry profiles",
				ToolName:    "UteamupTenantIndustryProfilesGet",
				MCPOnly:     true,
			},
			{
				Name:        "additional-industries-set",
				Description: "Replace the active tenant's additional industry profiles without changing its primary coding profile",
				ToolName:    "UteamupTenantAdditionalIndustryProfilesSet",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "industry-profile-guid", BodyName: "additionalIndustryProfileGuids", Description: "Complete additional industry-profile GUID set; repeatable, omit all values to clear", Type: "stringSlice", Default: []string{}},
				},
			},
			{
				Name:        "modules-get",
				Description: "List active-tenant module families, plan availability, and effective state",
				ToolName:    "UteamupTenantModulesGet",
				MCPOnly:     true,
			},
			{
				Name:        "module-set",
				Description: "Enable or disable one active-tenant module family within the plan entitlement ceiling",
				ToolName:    "UteamupTenantModuleSet",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "familyKey", Description: "Lower-case module family key, for example fleet", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "enabled", BodyName: "isEnabled", Description: "Enable or disable the family", Required: true, Type: "bool"},
				},
			},
			{
				Name:        "invite-defaults-get",
				Description: "Get the per-tenant invite-defaults configuration (auto-assign license + role)",
				ToolName:    "UteamupTenantInviteDefaultsGet",
				Args: []ArgDef{
					{Name: "tenantGuidToGet", Description: "Tenant GUID", Required: true, Type: "string"},
				},
			},
			{
				Name:        "invite-defaults-set",
				Description: "Update the per-tenant invite-defaults configuration",
				ToolName:    "UteamupTenantInviteDefaultsSet",
				Args: []ArgDef{
					{Name: "tenantGuidToUpdate", Description: "Tenant GUID", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "auto-license", Description: "Enable auto-assign license on invite", Type: "bool"},
					{Name: "license-type", Description: "License type to auto-assign: 0=Regular, 1=Helpdesk", Type: "int"},
					{Name: "auto-role", Description: "Enable auto-assign role on invite", Type: "bool"},
					{Name: "role-id", Description: "GUID of the default tenant-scoped role to auto-assign", Type: "string"},
				},
			},
			{
				Name:        "extend-trial",
				Description: "Extend a tenant's trial period (revives expired trials; requires Tenant.ExtendTrial)",
				ToolName:    "UteamupTenantExtendTrial",
				HTTPMethod:  "POST",
				RESTPath:    "{tenantGuid}/extend-trial",
				Args:        []ArgDef{{Name: "tenantGuid", Description: "Tenant GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "extend-by-days", BodyName: "extendByDays", Description: "Days to add to the current trial end (1-365); exactly one of this or --new-trial-end", Type: "int"},
					{Name: "new-trial-end", BodyName: "newTrialEndDate", Description: "Absolute new trial end (ISO 8601 UTC); exactly one of this or --extend-by-days", Type: "string"},
					{Name: "note", Description: "Audit note stored with the extension", Type: "string"},
				},
			},
		}...),
	})
}
