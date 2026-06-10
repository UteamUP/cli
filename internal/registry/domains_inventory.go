package registry

func init() {
	Register(&Domain{
		Name:        "stock",
		Aliases:     []string{"stocks"},
		Description: "Manage stock/inventory",
		Actions: append(crudActions("Stock"),
			Action{
				Name:        "search",
				Description: "Search stock items across locations (low-stock lookups included)",
				ToolName:    "UteamupStockSearchItems",
				RESTPath:    "items/search",
				Flags: append([]FlagDef{
					{Name: "q", Description: "Free-text search term (name, SKU, part number, GTIN)", Type: "string"},
					{Name: "type", Description: "Item type filter (Part, Tool, Chemical)", Type: "string"},
					{Name: "stock-guid", Description: "Stock location GUID filter", Type: "string"},
				}, paginationFlags()...),
			},
			Action{
				Name:        "alerts",
				Description: "List active alerts for a stock location",
				ToolName:    "UteamupStockListAlerts",
				RESTPath:    "locations/{stockGuid}/alerts",
				Flags: []FlagDef{
					{Name: "stock-guid", Description: "Stock location GUID", Type: "string", Required: true},
				},
			},
			Action{
				Name:        "ack-alert",
				Description: "Acknowledge a stock alert",
				ToolName:    "UteamupStockAcknowledgeAlert",
				HTTPMethod:  "POST",
				RESTPath:    "alerts/{alertGuid}/acknowledge",
				Args:        []ArgDef{{Name: "alertGuid", Description: "Stock alert GUID", Required: true, Type: "string"}},
			},
			Action{
				Name:        "po-list",
				Description: "List purchase orders (paged, optional status filter)",
				ToolName:    "UteamupStockListPurchaseOrders",
				RESTPath:    "purchase-orders",
				Flags: append([]FlagDef{
					{Name: "status", Description: "Status filter (Draft, Submitted, Approved, Received, Cancelled)", Type: "string"},
				}, paginationFlags()...),
			},
			Action{
				Name:        "po-get",
				Description: "Get a purchase order by GUID",
				ToolName:    "UteamupStockGetPurchaseOrder",
				RESTPath:    "purchase-orders/{guid}",
				Args:        []ArgDef{{Name: "guid", Description: "Purchase order GUID", Required: true, Type: "string"}},
			},
			Action{
				Name:        "po-submit",
				Description: "Submit a Draft purchase order for approval",
				ToolName:    "UteamupStockSubmitPurchaseOrder",
				HTTPMethod:  "POST",
				RESTPath:    "purchase-orders/{guid}/submit",
				Args:        []ArgDef{{Name: "guid", Description: "Purchase order GUID", Required: true, Type: "string"}},
			},
			Action{
				Name:        "po-approve",
				Description: "Approve a Submitted purchase order",
				ToolName:    "UteamupStockApprovePurchaseOrder",
				HTTPMethod:  "POST",
				RESTPath:    "purchase-orders/{guid}/approve",
				Args:        []ArgDef{{Name: "guid", Description: "Purchase order GUID", Required: true, Type: "string"}},
			},
			Action{
				Name:        "po-cancel",
				Description: "Cancel a purchase order",
				ToolName:    "UteamupStockCancelPurchaseOrder",
				HTTPMethod:  "POST",
				RESTPath:    "purchase-orders/{guid}/cancel",
				Args:        []ArgDef{{Name: "guid", Description: "Purchase order GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					// Always sent (Default "") so the POST carries a JSON body — the
					// backend binds [FromBody] CancelPurchaseOrderRequestModel.
					{Name: "reason", Description: "Cancellation reason", Default: "", Type: "string"},
				},
			},
			Action{
				Name:        "reorder-policy-get",
				Description: "Get the tenant's stock auto-replenishment policy",
				ToolName:    "UteamupStockGetReorderPolicy",
				RESTPath:    "reorder-policy",
			},
			Action{
				Name:        "reorder-policy-set",
				Description: "Update the tenant's stock auto-replenishment policy (caps/vendors are managed in the web app and reset by this full upsert)",
				ToolName:    "UteamupStockUpdateReorderPolicy",
				HTTPMethod:  "PUT",
				RESTPath:    "reorder-policy",
				Flags: []FlagDef{
					{Name: "enabled", Description: "Enable automatic reorder evaluation", Default: false, Type: "bool", BodyName: "isEnabled"},
					{Name: "cron-schedule", Description: "Cron schedule for reorder evaluation", Default: "0 3 * * *", Type: "string"},
					{Name: "auto-submit", Description: "Auto-submit generated purchase orders for approval", Default: false, Type: "bool", BodyName: "autoSubmitEnabled"},
					{Name: "auto-deduct", Description: "Auto-deduct stock on workorder completion", Default: false, Type: "bool", BodyName: "autoDeductOnWorkorderCompletionEnabled"},
				},
			},
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
