package registry

func init() {
	Register(&Domain{
		Name:        "admin-users",
		Aliases:     []string{"adminusers", "gausers"},
		Description: "Global-admin user management (list, detail, login events, disable/enable, password reset)",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List users across all tenants (global-admin only)",
				ToolName:    "UteamupAdminUserList",
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page (max 100)", Default: 25, Type: "int"},
					{Name: "search", Short: "q", Description: "Search email, first name, last name", Type: "string"},
					{Name: "sort", Description: "Sort order: lastLogin, email, name", Type: "string"},
					{Name: "include-deleted", Description: "Include soft-deleted users", Default: false, Type: "bool"},
					{Name: "include-disabled", Description: "Include disabled users", Default: true, Type: "bool"},
				},
			},
			{
				Name:        "get",
				Description: "Get a user's full detail by GUID",
				ToolName:    "UteamupAdminUserGet",
				Args:        []ArgDef{{Name: "guid", Description: "User GUID", Required: true, Type: "string"}},
			},
			{
				Name:        "login-events",
				Description: "Paginated login/logout history for a user",
				ToolName:    "UteamupAdminUserLoginEvents",
				Args:        []ArgDef{{Name: "guid", Description: "User GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page (max 100)", Default: 50, Type: "int"},
				},
			},
			{
				Name:        "disable",
				Description: "Disable (ban) a user. Revokes refresh tokens immediately.",
				ToolName:    "UteamupAdminUserDisable",
				Args:        []ArgDef{{Name: "guid", Description: "User GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "confirm-email", Description: "Target user's email (case-insensitive match required)", Type: "string", Required: true},
					{Name: "reason", Description: "Optional reason captured on the audit log", Type: "string"},
				},
			},
			{
				Name:        "enable",
				Description: "Re-enable a previously disabled user",
				ToolName:    "UteamupAdminUserEnable",
				Args:        []ArgDef{{Name: "guid", Description: "User GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "reason", Description: "Optional reason captured on the audit log", Type: "string"},
				},
			},
			{
				Name:        "reset-password",
				Description: "Admin-initiated password reset (email-link OR one-time temp password)",
				ToolName:    "UteamupAdminUserResetPassword",
				Args:        []ArgDef{{Name: "guid", Description: "User GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "mode", Description: "EmailLink (send reset link) or TempPassword (return one-time password)", Type: "string", Required: true},
					{Name: "confirm-email", Description: "Target user's email (case-insensitive match required)", Type: "string", Required: true},
					{Name: "reason", Description: "Optional reason captured on the audit log", Type: "string"},
				},
			},
		},
	})
}
