package registry

func init() {
	Register(&Domain{
		Name:        "stock",
		Aliases:     []string{"stocks"},
		Description: "Manage stock/inventory",
		Actions: append(crudActions("Stock"),
			Action{Name: "search", Description: "Search stock", ToolName: "UteamupStockSearch", Args: queryArg(), Flags: paginationFlags()},
		),
	})

	Register(&Domain{
		Name:        "part",
		Aliases:     []string{"parts"},
		Description: "Manage parts",
		Actions: append(crudActions("Part"),
			Action{Name: "search", Description: "Search parts", ToolName: "UteamupPartSearch", Args: queryArg(), Flags: paginationFlags()},
		),
	})

	Register(&Domain{Name: "chemical", Aliases: []string{"chemicals"}, Description: "Manage chemicals", Actions: crudActions("Chemical")})
	Register(&Domain{Name: "tool", Aliases: []string{"tools"}, Description: "Manage tools/equipment", Actions: crudActions("Tool")})
	Register(&Domain{Name: "inventory", Description: "Manage inventory", Actions: crudActions("Inventory")})
}
