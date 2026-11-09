package registry

// The unpaid tenant billing lifecycle: the current tenant's stage and dates, the owner's one
// restore from the archive, the Global Admin unarchive / postpone-deletion / legal-hold /
// always-free-demo controls and the platform settings (configured Global Admin only; the backend refuses
// everyone else with 403).
func init() {
	Register(&Domain{
		Name:        "tenant-lifecycle",
		Aliases:     []string{"lifecycle"},
		Description: "Unpaid tenant billing lifecycle: grace, archive, deletion and Global Admin controls",
		APIPath:     "/api/billing/lifecycle",
		Actions: []Action{
			{
				Name:              "get",
				Description:       "Stage, unpaid/grace/archive/deletion dates and allowed actions for the current tenant",
				ToolName:          "TenantLifecycleGet",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
			},
			{
				Name:        "restore",
				Description: "Restore the archived current tenant to complete payment (owner only, once per unpaid episode)",
				ToolName:    "TenantLifecycleRestore",
				HTTPMethod:  "POST",
				RESTPath:    "restore",
			},
			{
				Name:         "admin-get",
				Description:  "Any tenant's lifecycle with projected dates and dedicated resources (Global Admin)",
				ToolName:     "AdminGetTenantLifecycle",
				HTTPMethod:   "GET",
				RESTBasePath: "/api/globaladmin/tenants",
				RESTPath:     "{tenantGuid}/lifecycle",
				Args:         []ArgDef{{Name: "tenantGuid", Description: "Tenant GUID", Required: true, Type: "non-empty-uuid"}},
			},
			{
				Name:         "unarchive",
				Description:  "Take a tenant out of the archive for 1-180 days (Global Admin)",
				ToolName:     "AdminUnarchiveTenant",
				HTTPMethod:   "POST",
				RESTBasePath: "/api/globaladmin/tenants",
				RESTPath:     "{tenantGuid}/lifecycle/unarchive",
				Args:         []ArgDef{{Name: "tenantGuid", Description: "Tenant GUID", Required: true, Type: "non-empty-uuid"}},
				Flags: []FlagDef{
					{Name: "extend-days", BodyName: "extendDays", Description: "Days before the tenant can be archived again (1-180)", Type: "int", Default: 30},
				},
			},
			{
				Name:         "postpone-deletion",
				Description:  "Move an archived tenant's deletion date, at most 365 days ahead (Global Admin)",
				ToolName:     "AdminPostponeTenantDeletion",
				HTTPMethod:   "POST",
				RESTBasePath: "/api/globaladmin/tenants",
				RESTPath:     "{tenantGuid}/lifecycle/postpone-deletion",
				Args:         []ArgDef{{Name: "tenantGuid", Description: "Tenant GUID", Required: true, Type: "non-empty-uuid"}},
				Flags: []FlagDef{
					{Name: "deletion-date", BodyName: "deletionDueAt", Description: "New deletion date, yyyy-MM-dd (UTC)", Required: true, Type: "string"},
				},
			},
			{
				Name:         "legal-hold",
				Description:  "Place or release a legal hold that stops archive and deletion (Global Admin)",
				ToolName:     "AdminSetTenantLegalHold",
				HTTPMethod:   "PUT",
				RESTBasePath: "/api/globaladmin/tenants",
				RESTPath:     "{tenantGuid}/lifecycle/legal-hold",
				Args:         []ArgDef{{Name: "tenantGuid", Description: "Tenant GUID", Required: true, Type: "non-empty-uuid"}},
				Flags: []FlagDef{
					{Name: "hold", Description: "true to place the hold, false to release it", Required: true, Type: "bool"},
					{Name: "reason", Description: "Why the hold is placed (required to place it)", Type: "string"},
				},
			},
			{
				Name:         "always-free-demo",
				Description:  "Make a tenant a permanently free demo workspace or restore normal billing; refused while a paid agreement, pending payment or unpaid invoice is open (Global Admin)",
				ToolName:     "AdminSetTenantAlwaysFreeDemoWorkspace",
				HTTPMethod:   "PUT",
				RESTBasePath: "/api/globaladmin/tenants",
				RESTPath:     "{tenantGuid}/lifecycle/always-free-demo-workspace",
				Args:         []ArgDef{{Name: "tenantGuid", Description: "Tenant GUID", Required: true, Type: "non-empty-uuid"}},
				Flags: []FlagDef{
					{Name: "enabled", Description: "true to make the workspace permanently free, false to restore normal billing", Required: true, Type: "bool"},
					{Name: "reason", Description: "Why the exemption changes (kept in the audit log, max 500 characters)", Required: true, Type: "string"},
				},
			},
			{
				Name:         "settings-get",
				Description:  "Platform lifecycle settings: configured and effective mode, purchase availability and periods (Global Admin)",
				ToolName:     "AdminGetTenantLifecycleSettings",
				HTTPMethod:   "GET",
				RESTBasePath: "/api/globaladmin/tenant-lifecycle",
				RESTPath:     "settings",
			},
			{
				Name:         "settings-update",
				Description:  "Change lifecycle settings; omitted flags keep their value. Enforce archives and deletes unpaid workspaces (Global Admin)",
				ToolName:     "AdminUpdateTenantLifecycleSettings",
				HTTPMethod:   "PUT",
				RESTBasePath: "/api/globaladmin/tenant-lifecycle",
				RESTPath:     "settings",
				Flags: []FlagDef{
					{Name: "mode", Description: "Off, DryRun or Enforce", Type: "string"},
					{Name: "card-grace-days", BodyName: "cardGraceDays", Description: "Days every member keeps access after a failed card payment (1-60)", Type: "int"},
					{Name: "archive-after-days", BodyName: "archiveAfterDays", Description: "Unpaid days before a workspace is archived (30-365)", Type: "int"},
					{Name: "delete-after-archive-days", BodyName: "deleteAfterArchiveDays", Description: "Days in the archive before deletion (30-365)", Type: "int"},
					{Name: "restore-window-days", BodyName: "restoreWindowDays", Description: "Days an owner-restored workspace has to pay (1-180)", Type: "int"},
					{Name: "final-export-retention-days", BodyName: "finalExportRetentionDays", Description: "Days the final export is kept after deletion (7-365)", Type: "int"},
					{Name: "max-transitions-per-run", BodyName: "maxTransitionsPerRun", Description: "Most transitions of one kind per run before it halts (1-500)", Type: "int"},
					{Name: "contact-email", BodyName: "contactEmail", Description: "Address named in notices for data requests", Type: "string"},
					{Name: "deprovision-dedicated-resources", BodyName: "deprovisionDedicatedResources", Description: "Deprovision a deleted tenant's dedicated database after a verified export", Type: "bool"},
				},
			},
		},
	})
}
