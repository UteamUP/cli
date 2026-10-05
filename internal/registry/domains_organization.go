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
				Description: "Resolve a scanned value (code, serial number, or bin code) to its typed target: stockItem | stockItemUnit | stockBin | asset | assetGroup | part | tool | chemical | workPermit | unknown",
				ToolName:    "UteamupCodeResolve",
				RESTPath:    "resolve/{value}",
				Args:        []ArgDef{{Name: "value", Description: "Scanned/typed value to resolve", Required: true, Type: "string"}},
			},
			Action{
				Name:        "target-list",
				Description: "List the QR codes, barcodes and NFC tags registered on a part, tool, chemical, stock item, stock bin or work permit",
				ToolName:    "UteamupCodeListForTarget",
				HTTPMethod:  "GET",
				RESTPath:    "targets/{targetType}/{targetGuid}",
				Args:        codeTargetArgs(),
			},
			Action{
				Name:        "target-register",
				Description: "Register a scanned QR code, barcode or NFC tag on a part, tool, chemical, stock item, stock bin or work permit",
				ToolName:    "UteamupCodeRegisterForTarget",
				HTTPMethod:  "POST",
				RESTPath:    "targets/{targetType}/{targetGuid}",
				Args:        codeTargetArgs(),
				Flags: []FlagDef{
					{Name: "type", Description: "QR | BARCODE | NFC", Required: true, Type: "string"},
					{Name: "value", Description: "Exactly what the scanner read", Required: true, Type: "string"},
					{Name: "description", Description: "Optional label; defaults to '{type} · {record name}'", Type: "string"},
				},
			},
			Action{
				Name:        "target-generate",
				Description: "Generate a printable barcode (e.g. PRT-7K2M9QXA) on a part, tool, chemical, stock item, stock bin or work permit; returns the existing barcode if it has one",
				ToolName:    "UteamupCodeGenerateForTarget",
				HTTPMethod:  "POST",
				RESTPath:    "targets/{targetType}/{targetGuid}/generate",
				Args:        codeTargetArgs(),
			},
			Action{
				Name:        "target-remove",
				Description: "Remove a code from a part, tool, chemical, stock item, stock bin or work permit (only a code registered on that target)",
				ToolName:    "UteamupCodeRemoveFromTarget",
				HTTPMethod:  "DELETE",
				RESTPath:    "targets/{targetType}/{targetGuid}/{codeGuid}",
				Args: append(codeTargetArgs(),
					ArgDef{Name: "codeGuid", Description: "GUID of the code to remove", Required: true, Type: "non-empty-uuid"}),
			},
		),
	})
	Register(&Domain{Name: "tag", Aliases: []string{"tags"}, Description: "Manage tags", Actions: crudActions("Tag")})
	Register(&Domain{Name: "tenant-holiday", Description: "Manage tenant holidays", Actions: []Action{
		{Name: "year", Description: "List tenant holidays for a year", ToolName: "UteamupTenantHolidayGetByYear", RESTPath: "year/{year}", Args: []ArgDef{{Name: "year", Description: "Holiday year", Required: true, Type: "int"}}},
		{Name: "create", Description: "Create a tenant holiday", ToolName: "UteamupTenantHolidayCreate", Flags: []FlagDef{jsonFlag()}},
		{Name: "update", Description: "Update a tenant holiday by GUID", ToolName: "UteamupTenantHolidayUpdate", Args: []ArgDef{{Name: "holidayGuid", Description: "Tenant holiday GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{holidayGuid}", Flags: []FlagDef{jsonFlag()}},
		{Name: "delete", Description: "Delete a tenant holiday by GUID", ToolName: "UteamupTenantHolidayDelete", Args: []ArgDef{{Name: "holidayGuid", Description: "Tenant holiday GUID", Required: true, Type: "string"}}, RESTPath: "by-guid/{holidayGuid}"},
		{Name: "import", Description: "Import tenant holidays for a country and year", ToolName: "UteamupTenantHolidayImport", HTTPMethod: "POST", RESTPath: "import/{year}", Args: []ArgDef{{Name: "year", Description: "Holiday year", Required: true, Type: "int"}}, Flags: []FlagDef{{Name: "country-code", Description: "ISO 2-letter country code", Default: "IS", Type: "string"}}},
	}})
	Register(&Domain{Name: "role", Aliases: []string{"roles"}, Description: "Manage roles", Actions: listGetActions("Role")})
}

// codeTargetTypes mirrors the backend CodeTargetTypes table (api/codes/targets/{targetType}).
var codeTargetTypes = []string{"part", "tool", "chemical", "stockitem", "stockbin", "workpermit"}

func codeTargetArgs() []ArgDef {
	return []ArgDef{
		{Name: "targetType", Description: "part | tool | chemical | stockitem | stockbin | workpermit", Required: true, Type: "string", AllowedValues: codeTargetTypes},
		{Name: "targetGuid", Description: "GUID of the target record", Required: true, Type: "non-empty-uuid"},
	}
}
