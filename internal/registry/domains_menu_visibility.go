package registry

func init() {
	Register(&Domain{
		Name:        "menu-visibility",
		Aliases:     []string{"sidebar-visibility", "hidden-menu"},
		Description: "Tenant-wide and personal sidebar menu visibility",
		APIPath:     "/api/menuvisibility",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get tenant-wide and personal hidden sidebar menu item keys",
				ToolName:    "UteamupMenuVisibilityGet",
				HTTPMethod:  "GET",
			},
			{
				Name:        "set-tenant",
				Description: "Replace the tenant-wide hidden sidebar list (hidden for every user, regardless of RBAC)",
				ToolName:    "UteamupMenuVisibilityTenantSet",
				HTTPMethod:  "PUT",
				RESTPath:    "tenant",
				Flags: []FlagDef{
					{
						Name:        "hidden-menu-item-keys",
						BodyName:    "hiddenMenuItemKeys",
						Description: "Route links to hide for everyone — repeatable or comma-separated",
						Type:        "stringSlice",
					},
				},
			},
			{
				Name:        "set-me",
				Description: "Replace the signed-in user's personal hidden sidebar list for this tenant",
				ToolName:    "UteamupMenuVisibilityMeSet",
				HTTPMethod:  "PUT",
				RESTPath:    "me",
				Flags: []FlagDef{
					{
						Name:        "hidden-menu-item-keys",
						BodyName:    "hiddenMenuItemKeys",
						Description: "Route links to hide in your sidebar — repeatable or comma-separated",
						Type:        "stringSlice",
					},
				},
			},
		},
	})
}
