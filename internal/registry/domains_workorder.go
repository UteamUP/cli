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
					{Name: "priority", Description: "Filter by priority (1=Low, 2=Medium, 3=High, 4=Urgent, 5=Critical)", Type: "string"},
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
					{Name: "priority", Description: "Priority (1=Low, 2=Medium, 3=High, 4=Urgent, 5=Critical)", Default: "Medium", Type: "string"},
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
					{Name: "priority", Description: "New priority (1=Low, 2=Medium, 3=High, 4=Urgent, 5=Critical)", Type: "string"},
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
			{
				Name:        "by-code",
				Description: "List work orders by coding system code branch prefix",
				ToolName:    "UteamupCodingsystemWorkorders",
				Args:        []ArgDef{{Name: "prefix", Description: "Code branch prefix (e.g., '1-HLA')", Required: true, Type: "string"}},
			},
			{
				Name:        "quick-close",
				Description: "Create and immediately close a work order from a Quick Close template in a single action. Subject to the MCP/CLI rate-limit tier (5/min, 50/day).",
				ToolName:    "UteamupWorkorderQuickClose",
				Flags: []FlagDef{
					{Name: "template", Description: "External GUID of the Quick Close template (required)", Required: true, Type: "string"},
					{Name: "asset", Description: "External GUID of the asset the work was performed on (required)", Required: true, Type: "string"},
					{Name: "note", Description: "Resolution note — what was done (3–4000 chars, required)", Required: true, Type: "string"},
					{Name: "idempotency-key", Description: "Client-generated GUID used to deduplicate retries. Optional — if omitted the CLI generates one per invocation.", Type: "string"},
					{Name: "industry-code", Description: "Optional external GUID of the industry/coding-catalog entry (informational)", Type: "string"},
					{Name: "performed-at", Description: "Optional UTC timestamp of when the work was performed (ISO 8601). Must be within the last 30 days and not in the future.", Type: "string"},
				},
			},
		},
	})
}
