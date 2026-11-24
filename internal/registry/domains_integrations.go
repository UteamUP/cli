package registry

func init() {
	Register(&Domain{
		Name:        "integrations",
		Description: "Read saved Storage, SharePoint and Business Central state; no live probes or writes",
		Actions: []Action{
			{Name: "storage", Description: "Saved Storage configuration/migration state (TenantStorage.Manage); health is not checked", ToolName: "uteamup_storage_status_get", MCPOnly: true},
			{Name: "sharepoint", Description: "Saved SharePoint validation/capabilities (SharePointIntegration.Manage); no migration or Graph probe", ToolName: "uteamup_sharepoint_status_get", MCPOnly: true},
			{Name: "business-central", Description: "Saved Business Central validation/refresh state (Tenant.Update, Enterprise); emulation is explicit", ToolName: "uteamup_business_central_status_get", MCPOnly: true},
		},
	})
}
