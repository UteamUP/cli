package registry

func init() {
	Register(&Domain{
		Name:        "client-security",
		Description: "Read tenant client-security settings or replace reviewed settings with Tenant.Update; unsigned settings do not grant offline access",
		APIPath:     "/api/tenantpolicies/client-security",
		Actions: []Action{
			{
				Name:              "get",
				Description:       "Read management settings for the active tenant; requires Tenant.Update and current membership",
				ToolName:          "UteamupTenantClientSecurityPolicyGet",
				HTTPMethod:        "GET",
				UseDomainBasePath: true,
			},
			{
				Name:        "effective",
				Description: "Read unsigned effective settings as a current tenant member; does not enroll a device or issue an offline lease",
				ToolName:    "UteamupTenantClientSecurityPolicyEffective",
				HTTPMethod:  "GET",
				RESTPath:    "effective",
			},
			{
				Name:            "history",
				Description:     "Read one retained policy-history page for the active tenant; requires Tenant.Update and current membership; unsigned history grants no offline authority",
				ToolName:        "UteamupTenantClientSecurityPolicyHistory",
				HTTPMethod:      "GET",
				RESTPath:        "history",
				RejectExtraArgs: true,
				Flags: []FlagDef{
					{
						Name:        "cursor",
						BodyName:    "cursor",
						QueryName:   "cursor",
						Description: "Opaque nextCursor from the preceding page, at most 1024 characters; omit for the first page",
						Type:        "string",
					},
					{
						Name:        "page-size",
						BodyName:    "pageSize",
						QueryName:   "pageSize",
						Description: "Raw audit positions including skipped snapshots, 1 through 100; the server validates this limit",
						Type:        "int",
						Default:     25,
					},
				},
			},
			{
				Name:              "set",
				Description:       "Replace reviewed client settings with expectedPolicyVersion from the latest read; disconnected devices receive changed limits only after reconnecting",
				ToolName:          "UteamupTenantClientSecurityPolicySet",
				HTTPMethod:        "PUT",
				UseDomainBasePath: true,
				Flags: []FlagDef{
					{
						Name:               "file",
						Description:        "Complete TenantClientSecurityPolicyUpdateModel JSON, including current expectedPolicyVersion",
						Type:               "string",
						Required:           true,
						RootJSONObjectFile: true,
					},
					{
						Name:        "confirm",
						Description: "Explicitly confirm the reviewed offline-access and lock-policy consequences",
						Type:        "bool",
						Required:    true,
						MustBeTrue:  true,
						LocalOnly:   true,
					},
				},
			},
		},
	})
}
