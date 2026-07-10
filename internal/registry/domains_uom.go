package registry

// Units of measure (stock-reseller-catalog pack-conversion §8). Base path /api/uom:
// list gates on Stock.View, create/delete on Stock.Update.
func init() {
	Register(&Domain{
		Name:        "uom",
		Aliases:     []string{"units-of-measure"},
		Description: "Manage units of measure (shared system set + tenant units)",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the units of measure available to the tenant (shared system set + tenant units)",
				ToolName:    "UteamupUomList",
			},
			{
				Name:        "create",
				Description: "Create a tenant unit of measure (duplicate code is rejected)",
				ToolName:    "UteamupUomCreate",
				Flags: []FlagDef{
					{Name: "code", Description: "Short code, unique per tenant (max 20 characters)", Required: true, Type: "string"},
					{Name: "name", Description: "Display name (max 100 characters)", Required: true, Type: "string"},
					{Name: "category", Description: "Category, e.g. Count, Volume, Mass, Length (optional)", Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a tenant unit of measure (the shared system set is not deletable)",
				ToolName:    "UteamupUomDelete",
				RESTPath:    "{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "Unit of measure GUID", Required: true, Type: "string"}},
			},
		},
	})
}
