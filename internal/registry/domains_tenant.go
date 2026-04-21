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
		},
	})
}
