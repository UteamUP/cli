package registry

func init() {
	Register(&Domain{
		Name:        "tenant",
		Aliases:     []string{"tenants"},
		Description: "Manage tenants (organizations)",
		Actions: []Action{
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
		},
	})
}
