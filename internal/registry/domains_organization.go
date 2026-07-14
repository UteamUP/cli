package registry

func init() {
	Register(&Domain{Name: "location", Aliases: []string{"locations", "loc"}, Description: "Manage locations", Actions: crudActions("Location")})
	Register(&Domain{Name: "floor-plan", Description: "Manage floor plans", Actions: crudActions("FloorPlan")})
	Register(&Domain{Name: "category", Aliases: []string{"categories", "cat"}, Description: "Manage categories", Actions: crudActions("Category")})
	Register(&Domain{Name: "currency", Aliases: []string{"currencies"}, Description: "Manage currencies", Actions: crudActions("Currency")})
	Register(&Domain{
		Name:        "code",
		Aliases:     []string{"codes"},
		Description: "Manage codes",
		// CodesController routes at api/codes (plural) — the auto-derived
		// "/api/code" base never matched a backend route.
		APIPath: "/api/codes",
		Actions: append(crudActions("Code"),
			Action{
				Name:        "resolve",
				Description: "Resolve a scanned value (code, serial number, or bin code) to its typed target: stockItem | stockItemUnit | stockBin | asset | unknown",
				ToolName:    "UteamupCodeResolve",
				RESTPath:    "resolve/{value}",
				Args:        []ArgDef{{Name: "value", Description: "Scanned/typed value to resolve", Required: true, Type: "string"}},
			},
		),
	})
	Register(&Domain{Name: "tag", Aliases: []string{"tags"}, Description: "Manage tags", Actions: crudActions("Tag")})
	Register(&Domain{Name: "tenant", Aliases: []string{"tenants"}, Description: "Manage tenants", Actions: listGetActions("Tenant")})
	Register(&Domain{Name: "tenant-holiday", Description: "Manage tenant holidays", Actions: []Action{
		{Name: "year", Description: "List tenant holidays for a year", ToolName: "UteamupTenantHolidayGetByYear", RESTPath: "year/{year}", Args: []ArgDef{{Name: "year", Description: "Holiday year", Required: true, Type: "int"}}},
		{Name: "create", Description: "Create a tenant holiday", ToolName: "UteamupTenantHolidayCreate", Flags: []FlagDef{jsonFlag()}},
		{Name: "update", Description: "Update a tenant holiday by GUID", ToolName: "UteamupTenantHolidayUpdate", Args: []ArgDef{{Name: "holidayGuid", Description: "Tenant holiday GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{holidayGuid}", Flags: []FlagDef{jsonFlag()}},
		{Name: "delete", Description: "Delete a tenant holiday by GUID", ToolName: "UteamupTenantHolidayDelete", Args: []ArgDef{{Name: "holidayGuid", Description: "Tenant holiday GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{holidayGuid}"},
		{Name: "import", Description: "Import tenant holidays for a country and year", ToolName: "UteamupTenantHolidayImport", HTTPMethod: "POST", RESTPath: "import/{year}", Args: []ArgDef{{Name: "year", Description: "Holiday year", Required: true, Type: "int"}}, Flags: []FlagDef{{Name: "country-code", Description: "ISO 2-letter country code", Default: "IS", Type: "string"}}},
	}})
	Register(&Domain{Name: "role", Aliases: []string{"roles"}, Description: "Manage roles", Actions: listGetActions("Role")})
}
