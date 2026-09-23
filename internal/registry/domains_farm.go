package registry

// Farm production places use the same public GUID in REST, MCP and CLI.
func init() {
	Register(&Domain{
		Name:        "farm-unit",
		Aliases:     []string{"farm-units"},
		Description: "Read farm fields, housing and aquatic production places",
		APIPath:     "/api/farm/units",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List farm production units",
				ToolName:    "UteamupFarmUnitList",
				Flags: append([]FlagDef{
					{Name: "type", Description: "Filter by unit type", Type: "string"},
					{Name: "is-active", BodyName: "isActive", Description: "Filter by active state", Type: "bool"},
					{Name: "location-guid", BodyName: "locationGuid", Description: "Filter by Location GUID", Type: "string"},
					{Name: "search", Description: "Search unit names", Type: "string"},
				}, paginationFlags()...),
			},
			{
				Name:        "get",
				Description: "Get a farm production unit by GUID",
				ToolName:    "UteamupFarmUnitGet",
				Args:        externalGUIDArg(),
			},
		},
	})
}
