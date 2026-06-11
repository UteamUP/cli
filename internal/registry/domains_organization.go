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
	Register(&Domain{Name: "tenant-holiday", Description: "Manage tenant holidays", Actions: crudActions("TenantHoliday")})
	Register(&Domain{Name: "role", Aliases: []string{"roles"}, Description: "Manage roles", Actions: listGetActions("Role")})
}
