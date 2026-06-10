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
			Action{
				Name:        "transfer",
				Description: "Transfer stock from its current location to a destination location atomically",
				ToolName:    "TransferInventory",
				HTTPMethod:  "POST",
				RESTPath:    "transfers",
				Flags: []FlagDef{
					{Name: "stock-item-guid", Description: "Stock item GUID to transfer", Required: true, Type: "string"},
					{Name: "destination-stock-guid", Description: "Destination stock location GUID", Required: true, Type: "string"},
					{Name: "quantity", Description: "Quantity to transfer (must be available at source)", Required: true, Type: "int"},
					{Name: "destination-bin-guid", Description: "Destination bin GUID (optional)", Type: "string"},
					{Name: "reason", Description: "Reason for the transfer", Type: "string"},
					{Name: "reference", Description: "External reference (e.g. ticket number)", Type: "string"},
				},
			},
			Action{
				Name:        "transfers",
				Description: "List transfer history grouped by transfer (paged, optional location filter)",
				ToolName:    "UteamupStockListTransfers",
				RESTPath:    "transfers",
				Flags: append([]FlagDef{
					{Name: "stock-guid", Description: "Stock location GUID filter", Type: "string"},
				}, paginationFlags()...),
			},
			Action{
				Name:        "po-receive",
				Description: "Receive goods against an Approved purchase order (lines come from a JSON file)",
				ToolName:    "UteamupStockReceivePurchaseOrder",
				HTTPMethod:  "POST",
				RESTPath:    "purchase-orders/{guid}/receive",
				Args:        []ArgDef{{Name: "guid", Description: "Purchase order GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "file", Short: "f", Description: "Path to a JSON file with the received lines: [{\"purchaseOrderItemGuid\":\"…\",\"receivedQuantity\":N}]", Required: true, Type: "string", JSONFile: true, BodyName: "receivedItems"},
				},
			},
			Action{
				Name:        "bulk-adjust",
				Description: "Apply a batch of stock adjustments atomically, all-or-nothing (operations come from a JSON file, max 500)",
				ToolName:    "UteamupStockBulkAdjust",
				HTTPMethod:  "POST",
				RESTPath:    "transactions/bulk",
				Flags: []FlagDef{
					{Name: "file", Short: "f", Description: "Path to a JSON file with the operations: [{\"stockItemGuid\":\"…\",\"action\":\"Add|Remove\",\"quantity\":N,\"reason\":\"…\"}]", Required: true, Type: "string", JSONFile: true, BodyName: "operations"},
				},
			},
			Action{
				Name:        "export",
				Description: "Export all stock items as CSV (round-trippable with import)",
				ToolName:    "UteamupStockExportItems",
				RESTPath:    "items/export",
			},
			Action{
				Name:        "import",
				Description: "Import stock items from a CSV file (upsert keyed by Sku, fallback InternalNumber)",
				ToolName:    "UteamupStockImportItems",
				HTTPMethod:  "POST",
				RESTPath:    "items/import",
				Flags: []FlagDef{
					{Name: "file", Short: "f", Description: "Path to the CSV file to import", Required: true, Type: "string", UploadFile: true},
					{Name: "dry-run", Description: "Run the full import pipeline without persisting", Default: false, Type: "bool", BodyName: "dryrun"},
				},
			},
			Action{
				Name:        "bins",
				Description: "List the bin hierarchy of a stock location",
				ToolName:    "UteamupStockListBins",
				RESTPath:    "locations/{stockGuid}/bins",
				Flags: []FlagDef{
					{Name: "stock-guid", Description: "Stock location GUID", Type: "string", Required: true},
				},
			},
			Action{
				Name:        "bins-create",
				Description: "Create a bin (Zone, Aisle, Rack, Shelf, or Bin) inside a stock location",
				ToolName:    "UteamupStockUpsertBin",
				HTTPMethod:  "POST",
				RESTPath:    "bins",
				Flags: []FlagDef{
					{Name: "stock-guid", Description: "Stock location GUID", Required: true, Type: "string"},
					{Name: "code", Description: "Bin code, unique within the stock location", Required: true, Type: "string"},
					{Name: "name", Description: "Display name (optional)", Type: "string"},
					{Name: "bin-type", Description: "Bin type: Zone, Aisle, Rack, Shelf, or Bin", Default: "Bin", Type: "string"},
					{Name: "parent-bin-guid", Description: "Parent bin GUID for nesting (optional)", Type: "string"},
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
