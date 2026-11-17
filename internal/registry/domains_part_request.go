package registry

// Part request CLI surface — mirrors the uteamup_part_request_{list,get,create,update}
// MCP tools over the REST adapter at /api/part-requests, so it works for both human
// sign-in and API-key sessions. Approve, reject and purchase-order bundling stay in
// the app for procurement reviewers and are intentionally not exposed here.

// partRequestMaxPageSize matches the 1..100 page size the list tool accepts.
const partRequestMaxPageSize = 100

var (
	partRequestStatusValues  = []string{"pending", "approved", "ordered", "fulfilled", "rejected"}
	partRequestUrgencyValues = []string{"low", "medium", "high", "critical"}
)

func partRequestRequestGUIDArg() []ArgDef {
	return []ArgDef{{Name: "requestGuid", Description: "Part request public GUID", Required: true, Type: "non-empty-uuid"}}
}

// partRequestEditableFlags are the fields of PartRequestCreateModel and
// PartRequestUpdateModel. On update the backend replaces every field, so an
// omitted optional field is cleared, except --notes, which keeps its value.
func partRequestEditableFlags(update bool) []FlagDef {
	clearHint := ""
	if update {
		clearHint = "; omission clears it"
	}
	return []FlagDef{
		{Name: "part-name", BodyName: "partName", Description: "Part name or description, up to 256 characters", Required: true, Type: "string"},
		{Name: "quantity", Description: "Quantity to procure, greater than zero (backend default 1 on create)", Required: update, Type: "decimal"},
		{Name: "part-guid", BodyName: "partGuid", Description: "Catalog part public GUID" + clearHint, Type: "non-empty-uuid"},
		{Name: "part-number", BodyName: "partNumber", Description: "OEM/catalog part number, up to 100 characters" + clearHint, Type: "string"},
		{Name: "asset-guid", BodyName: "assetGuid", Description: "Asset public GUID the part is for" + clearHint, Type: "non-empty-uuid"},
		{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Work order public GUID that needs the part" + clearHint, Type: "non-empty-uuid"},
		{Name: "preferred-vendor-guid", BodyName: "preferredVendorGuid", Description: "Preferred vendor public GUID" + clearHint, Type: "non-empty-uuid"},
		{Name: "delivery-location-guid", BodyName: "deliveryLocationGuid", Description: "Delivery location public GUID" + clearHint, Type: "non-empty-uuid"},
		{Name: "uom", Description: "Unit of measure, e.g. pcs, meters, set" + clearHint, Type: "string"},
		{Name: "estimated-unit-cost", BodyName: "estimatedUnitCost", Description: "Estimated unit cost, non-negative" + clearHint, Type: "decimal"},
		{Name: "currency", Description: "Currency code for the estimate, e.g. EUR, ISK" + clearHint, Type: "string"},
		{Name: "urgency", Description: "Urgency: low, medium, high or critical", Default: "medium", Type: "string", AllowedValues: partRequestUrgencyValues},
		{Name: "required-by-date", BodyName: "requiredByDate", Description: "Date the part is required on-site, ISO 8601 (UTC)" + clearHint, Type: "string"},
		{Name: "notes", BodyName: "reviewNotes", Description: "Notes for the procurement reviewer, up to 2,000 characters; on update a non-empty note replaces the current one and omission keeps it", Type: "string"},
	}
}

func init() {
	Register(&Domain{
		Name:        "part-request",
		Aliases:     []string{"part-requests", "partrequest", "preq"},
		Description: "Spare part requisitions: list, read, draft and edit Pending requests",
		APIPath:     "/api/part-requests",
		Actions: []Action{
			{
				Name:                    "list",
				Description:             "List spare part requisitions newest first, filtered by status, urgency, asset or work order. Requires Stock.View, WorkOrder.View or Asset.View",
				ToolName:                "UteamupPartRequestList",
				RejectExtraArgs:         true,
				OffsetPagingMaxPageSize: partRequestMaxPageSize,
				Flags: []FlagDef{
					{Name: "status", Description: "Only this status: pending, approved, ordered, fulfilled or rejected", Type: "string", AllowedValues: partRequestStatusValues},
					{Name: "urgency", Description: "Only this urgency: low, medium, high or critical", Type: "string", AllowedValues: partRequestUrgencyValues},
					{Name: "asset-guid", BodyName: "assetGuid", Description: "Only requisitions for this asset public GUID", Type: "non-empty-uuid"},
					{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Only requisitions for this work order public GUID", Type: "non-empty-uuid"},
					{Name: "page", Short: "p", Description: "Page number, from 1", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Rows per page, 1 to 100", Default: 50, Type: "int"},
				},
			},
			{
				Name:            "get",
				Description:     "Read one spare part requisition by GUID. Requires Stock.View, WorkOrder.View or Asset.View",
				ToolName:        "UteamupPartRequestGet",
				RESTPath:        "{requestGuid}",
				RejectExtraArgs: true,
				Args:            partRequestRequestGUIDArg(),
			},
			{
				Name:            "create",
				Description:     "Draft a Pending spare part requisition; nothing is approved, ordered or taken from stock. Requires WorkOrder.Update, Asset.Update, Stock.View or Stock.Update",
				ToolName:        "UteamupPartRequestCreate",
				RejectExtraArgs: true,
				Flags:           partRequestEditableFlags(false),
			},
			{
				Name:            "update",
				Description:     "Replace the editable fields of a requisition that is still Pending; omitted optional fields are cleared except --notes. Requires WorkOrder.Update, Asset.Update or Stock.Update",
				ToolName:        "UteamupPartRequestUpdate",
				RESTPath:        "{requestGuid}",
				RejectExtraArgs: true,
				Args:            partRequestRequestGUIDArg(),
				Flags:           partRequestEditableFlags(true),
			},
		},
	})
}
