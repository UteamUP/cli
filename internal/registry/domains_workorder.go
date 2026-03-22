package registry

func init() {
	Register(&Domain{
		Name:        "workorder",
		Aliases:     []string{"wo", "workorders"},
		Description: "Manage work orders",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List work orders with pagination and filtering",
				ToolName:    "UteamupWorkOrderList",
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
					{Name: "status", Description: "Filter by status", Type: "string"},
					{Name: "priority", Description: "Filter by priority", Type: "string"},
					{Name: "sort-by", Description: "Sort field", Default: "CreatedAt", Type: "string"},
					{Name: "sort-order", Description: "Sort direction (asc or desc)", Default: "desc", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get work order details by ID",
				ToolName:    "UteamupWorkOrderGet",
				Args:        []ArgDef{{Name: "id", Description: "Work order ID", Required: true, Type: "int"}},
			},
			{
				Name:        "create",
				Description: "Create a new work order",
				ToolName:    "UteamupWorkOrderCreate",
				Flags: []FlagDef{
					{Name: "title", Description: "Work order title", Required: true, Type: "string"},
					{Name: "description", Description: "Work order description", Type: "string"},
					{Name: "priority", Description: "Priority (Low, Medium, High, Critical)", Default: "Medium", Type: "string"},
					{Name: "asset-id", Description: "Associated asset ID", Type: "int"},
					{Name: "assigned-to", Description: "Assigned user ID", Type: "int"},
					{Name: "from-json", Description: "JSON file with work order data", Type: "string"},
				},
			},
			{
				Name:        "update",
				Description: "Update an existing work order",
				ToolName:    "UteamupWorkOrderUpdate",
				Args:        []ArgDef{{Name: "id", Description: "Work order ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "title", Description: "New title", Type: "string"},
					{Name: "status", Description: "New status", Type: "string"},
					{Name: "priority", Description: "New priority", Type: "string"},
					{Name: "from-json", Description: "JSON file with update data", Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a work order by ID",
				ToolName:    "UteamupWorkOrderDelete",
				Args:        []ArgDef{{Name: "id", Description: "Work order ID", Required: true, Type: "int"}},
			},
			{
				Name:        "search",
				Description: "Search work orders by title or description",
				ToolName:    "UteamupWorkOrderSearch",
				Args:        []ArgDef{{Name: "query", Description: "Search term", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
				},
			},
		},
	})
}
