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
				Name:        "reservations",
				Description: "List active stock reservations and backorders (filter by item, workorder, or project)",
				ToolName:    "UteamupStockListReservations",
				RESTPath:    "reservations",
				Flags: append([]FlagDef{
					{Name: "item-guid", Description: "Filter by stock item GUID", Type: "string"},
					{Name: "workorder-guid", Description: "Filter by workorder GUID", Type: "string"},
					{Name: "project-guid", Description: "Filter by project GUID", Type: "string"},
				}, paginationFlags()...),
			},
			Action{
				Name:        "units",
				Description: "List serialized units of a stock item (paged, optional status/serial filters)",
				ToolName:    "UteamupStockListUnits",
				RESTPath:    "items/{itemGuid}/units",
				Flags: append([]FlagDef{
					{Name: "item-guid", Description: "Stock item GUID", Required: true, Type: "string"},
					{Name: "status", Description: "Status filter (InStock, Reserved, InTransit, Installed, RMA, Quarantined, Scrapped, Retired)", Type: "string"},
					{Name: "serial", Description: "Serial number filter (full or partial)", Type: "string"},
				}, paginationFlags()...),
			},
			Action{
				Name:        "units-lookup",
				Description: "Look up a serialized unit by its exact serial number",
				ToolName:    "UteamupStockLookupUnit",
				RESTPath:    "units/lookup/{serial}",
				Args:        []ArgDef{{Name: "serial", Description: "Exact serial number of the unit", Required: true, Type: "string"}},
			},
			Action{
				Name:        "unit-transition",
				Description: "Transition a serialized unit to a new lifecycle status (assetGuid required for Installed)",
				ToolName:    "UteamupStockTransitionUnit",
				HTTPMethod:  "POST",
				RESTPath:    "units/{unitGuid}/transition",
				Args:        []ArgDef{{Name: "unitGuid", Description: "Stock unit GUID", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "target-status", Description: "Target status: InStock, Reserved, InTransit, Installed, RMA, Quarantined, Scrapped, or Retired", Required: true, Type: "string"},
					{Name: "asset-guid", Description: "Asset GUID (required when target status is Installed)", Type: "string"},
					{Name: "workorder-guid", Description: "Workorder GUID linking the transition to a workorder", Type: "string"},
					{Name: "reason", Description: "Reason for the transition", Type: "string"},
				},
			},
			Action{
				Name:        "reserve",
				Description: "Reserve stock for a workorder, project, or ad-hoc hold (returns Reserved or Backordered)",
				ToolName:    "UteamupStockCreateReservation",
				HTTPMethod:  "POST",
				RESTPath:    "reservations",
				Flags: []FlagDef{
					{Name: "stock-item-guid", Description: "Stock item GUID to reserve", Required: true, Type: "string"},
					{Name: "quantity", Description: "Quantity to reserve (minimum 1)", Required: true, Type: "int"},
					{Name: "workorder-guid", Description: "Workorder GUID that owns the hold", Type: "string"},
					{Name: "project-guid", Description: "Project GUID that owns the hold", Type: "string"},
					{Name: "unit-guid", Description: "Specific serialized unit GUID to reserve (quantity must be 1)", Type: "string"},
					{Name: "reserved-until", Description: "Expiry of the hold, RFC3339 (required when no workorder/project owns it)", Type: "string"},
				},
			},
			Action{
				Name:        "release",
				Description: "Release (cancel) a stock reservation, returning the held quantity to ATP",
				ToolName:    "UteamupStockReleaseReservation",
				HTTPMethod:  "POST",
				RESTPath:    "reservations/{reservationGuid}/release",
				Args:        []ArgDef{{Name: "reservationGuid", Description: "Stock reservation GUID", Required: true, Type: "string"}},
			},
			Action{
				Name:        "atp",
				Description: "Get available-to-promise figures for a stock item (on-hand, reserved, ATP, backordered)",
				ToolName:    "UteamupStockGetAtp",
				RESTPath:    "items/{itemGuid}/atp",
				Args:        []ArgDef{{Name: "itemGuid", Description: "Stock item GUID", Required: true, Type: "string"}},
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
