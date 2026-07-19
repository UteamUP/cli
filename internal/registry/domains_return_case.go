package registry

func returnCaseIdempotencyFlag() FlagDef {
	return FlagDef{
		Name:               "idempotency-key",
		BodyName:           "idempotencyKey",
		HeaderName:         "Idempotency-Key",
		MirrorHeaderInBody: true,
		Description:        "Stable retry UUID, sent identically in the required header and compatibility body field",
		Required:           true,
		Type:               "string",
	}
}

func returnCaseTransitionFlags() []FlagDef {
	return []FlagDef{
		returnCaseIdempotencyFlag(),
		{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Exact reviewed UpdatedAt timestamp", Required: true, Type: "string"},
		{Name: "reason", BodyName: "reason", Description: "Reviewed transition reason", Type: "string"},
		{Name: "shipment-reference", BodyName: "shipmentReference", Description: "Reviewed shipment reference", Type: "string"},
		{Name: "carrier", BodyName: "carrier", Description: "Reviewed carrier name", Type: "string"},
		{Name: "tracking-number", BodyName: "trackingNumber", Description: "Reviewed tracking number", Type: "string"},
		{Name: "evidence-document-guid", BodyName: "evidenceDocumentGuids", Description: "Existing tenant evidence document GUID (repeatable)", Type: "stringSlice"},
		{Name: "credit-amount", BodyName: "creditAmount", Description: "Optional partial credit amount", Type: "float"},
		{Name: "currency", BodyName: "currency", Description: "Required currency for an explicit partial credit", Type: "string"},
		{Name: "dispositions-file", BodyName: "dispositions", Description: "JSON file containing reviewed per-line physical dispositions", Type: "string", JSONFile: true},
	}
}

func returnCaseTransitionAction(name, description, toolName string) Action {
	return Action{
		Name:        name,
		Description: description,
		ToolName:    toolName,
		HTTPMethod:  "POST",
		RESTPath:    "{returnCaseGuid}/" + name,
		Args: []ArgDef{
			{Name: "returnCaseGuid", Description: "Return-case public GUID", Required: true, Type: "uuid"},
		},
		Flags: returnCaseTransitionFlags(),
	}
}

func init() {
	Register(&Domain{
		Name:        "return-case",
		Aliases:     []string{"return-cases", "returns"},
		Description: "Inspect and explicitly manage customer and vendor return lifecycles",
		APIPath:     "/api/returncases",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List tenant-scoped return cases with physical and commercial status",
				ToolName:    "UteamupReturnCaseList",
				Flags: []FlagDef{
					{Name: "page", BodyName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", BodyName: "pageSize", Description: "Results per page, maximum 200", Default: 50, Type: "int"},
					{Name: "status", BodyName: "status", Description: "Optional lifecycle status number", Type: "int"},
					{Name: "direction", BodyName: "direction", Description: "Optional direction: 0=customer return, 1=return to vendor", Type: "int"},
					{Name: "customer-guid", BodyName: "customerGuid", Description: "Optional customer public GUID", Type: "string"},
					{Name: "vendor-guid", BodyName: "vendorGuid", Description: "Optional vendor public GUID", Type: "string"},
					{Name: "requested-from-utc", BodyName: "requestedFromUtc", Description: "Optional requested-from UTC timestamp", Type: "string"},
					{Name: "requested-to-utc", BodyName: "requestedToUtc", Description: "Optional requested-to UTC timestamp", Type: "string"},
					{Name: "search", BodyName: "search", Description: "Case, shipment, carrier, or tracking search", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get one return with immutable events, ledger links, and resolution evidence",
				ToolName:    "UteamupReturnCaseGet",
				RESTPath:    "{returnCaseGuid}",
				Args: []ArgDef{
					{Name: "returnCaseGuid", Description: "Return-case public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create a reviewed idempotent customer or vendor return",
				ToolName:    "UteamupReturnCaseCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					returnCaseIdempotencyFlag(),
					{Name: "direction", BodyName: "direction", Description: "Direction: 0=customer return, 1=return to vendor", Required: true, Type: "int"},
					{Name: "reason", BodyName: "reason", Description: "Reviewed return reason", Required: true, Type: "string"},
					{Name: "customer-guid", BodyName: "customerGuid", Description: "Customer public GUID for a customer return", Type: "string"},
					{Name: "vendor-guid", BodyName: "vendorGuid", Description: "Vendor public GUID for a vendor return", Type: "string"},
					{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Optional source work-order public GUID", Type: "string"},
					{Name: "asset-guid", BodyName: "assetGuid", Description: "Optional affected asset public GUID", Type: "string"},
					{Name: "service-invoice-guid", BodyName: "serviceInvoiceGuid", Description: "Optional immutable service-invoice public GUID", Type: "string"},
					{Name: "purchase-order-guid", BodyName: "stockPurchaseOrderGuid", Description: "Optional purchase-order public GUID", Type: "string"},
					{Name: "lines-file", BodyName: "lines", Description: "JSON file containing stockItemGuid, optional stockItemUnitGuid, quantity, condition, and reason rows", Required: true, Type: "string", JSONFile: true},
				},
			},
			returnCaseTransitionAction(
				"approve",
				"Approve the exact reviewed return version",
				"UteamupReturnCaseApprove",
			),
			returnCaseTransitionAction(
				"reject",
				"Reject the exact reviewed return version with a reason",
				"UteamupReturnCaseReject",
			),
			returnCaseTransitionAction(
				"cancel",
				"Cancel the exact reviewed return version with a reason",
				"UteamupReturnCaseCancel",
			),
			returnCaseTransitionAction(
				"ship",
				"Record outbound shipment and evidence exactly once",
				"UteamupReturnCaseShip",
			),
			returnCaseTransitionAction(
				"receive",
				"Record a customer return into quarantine",
				"UteamupReturnCaseReceive",
			),
			returnCaseTransitionAction(
				"inspect",
				"Record reviewed physical dispositions from a JSON file",
				"UteamupReturnCaseInspect",
			),
			returnCaseTransitionAction(
				"credit",
				"Issue one immutable customer credit note or vendor credit memo",
				"UteamupReturnCaseCredit",
			),
			returnCaseTransitionAction(
				"replace",
				"Create exactly one linked replacement fulfillment",
				"UteamupReturnCaseReplace",
			),
			returnCaseTransitionAction(
				"repair",
				"Create exactly one linked repair work order",
				"UteamupReturnCaseRepair",
			),
			returnCaseTransitionAction(
				"close",
				"Close the exact resolved return version",
				"UteamupReturnCaseClose",
			),
		},
	})
}
