package registry

// The unpaid tenant billing lifecycle: the current tenant's stage and dates, the owner's one
// restore from the archive, and the Global Admin unarchive / postpone-deletion / legal-hold
// controls (configured Global Admin only; the backend refuses everyone else with 403).
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
		},
	})
}
