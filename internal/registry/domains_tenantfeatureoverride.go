package registry

func init() {
	Register(&Domain{
		Name:        "tenant-feature-override",
		Aliases:     []string{"tenant-feature-overrides", "tenantfeatureoverride", "feature-override", "feature-overrides"},
		Description: "Grant or revoke catalog modules per tenant, overriding the plan",
		APIPath:     "/api/tenantfeatureoverride",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List a tenant's feature overrides",
				ToolName:    "UteamupTenantFeatureOverrideList",
				RESTPath:    "by-tenant/{tenantGuid}",
				Args:        []ArgDef{{Name: "tenantGuid", Description: "Tenant GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "upsert",
				Description: "Grant (--mode 0) or revoke (--mode 1) a catalog module for a tenant",
				ToolName:    "UteamupTenantFeatureOverrideUpsert",
				HTTPMethod:  "PUT",
				RESTPath:    "by-tenant/{tenantGuid}",
				Args:        []ArgDef{{Name: "tenantGuid", Description: "Tenant GUID (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "feature-catalog-guid", Description: "Feature catalog module GUID (required)", Required: true, Type: "string"},
					{Name: "mode", Description: "Override mode: 0 = Grant, 1 = Revoke (required)", Required: true, Type: "int"},
					{Name: "reason", Description: "Optional reason recorded on the override", Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Remove a tenant's override for one catalog module (plan default applies again)",
				ToolName:    "UteamupTenantFeatureOverrideDelete",
				RESTPath:    "by-tenant/{tenantGuid}/{featureCatalogGuid}",
				Args: []ArgDef{
					{Name: "tenantGuid", Description: "Tenant GUID", Required: true, Type: "string"},
					{Name: "featureCatalogGuid", Description: "Feature catalog module GUID", Required: true, Type: "string"},
				},
			},
		},
	})
}
