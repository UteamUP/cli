package registry

func init() {
	commonTerms := []FlagDef{
		{Name: "name", BodyName: "name", Description: "Price-list display name", Required: true, Type: "string"},
		{Name: "currency", BodyName: "currency", Description: "Three-letter ISO currency code", Required: true, Type: "string"},
		{Name: "effective-from", BodyName: "effectiveFrom", Description: "Price-list start in ISO-8601 UTC", Required: true, Type: "string"},
		{Name: "effective-to", BodyName: "effectiveTo", Description: "Optional price-list end in ISO-8601 UTC", Type: "string"},
		{Name: "is-active", BodyName: "isActive", Description: "Whether the price-list version is active", Default: true, Type: "bool"},
		{Name: "notes", BodyName: "notes", Description: "Optional reviewed price-list notes", Type: "string"},
		{Name: "items-json", BodyName: "items", Description: "Path to a JSON array of GUID-only charge rules", Required: true, Type: "string", JSONFile: true},
	}

	createFlags := append([]FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped write idempotency UUID", Required: true, Type: "string"},
	}, commonTerms...)

	updateFlags := append([]FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped write idempotency UUID", Required: true, Type: "string"},
	}, commonTerms...)
	updateFlags = append(updateFlags,
		FlagDef{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Exact reviewed UpdatedAt timestamp", Required: true, Type: "string"})

	Register(&Domain{
		Name:        "service-price-list",
		Aliases:     []string{"service-price-lists", "price-list"},
		Description: "Manage versioned customer-facing field-service charge rules",
		APIPath:     "/api/service-price-lists",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List service price lists and agreement-usage evidence",
				ToolName:    "UteamupServicePriceListList",
				Flags: []FlagDef{
					{Name: "active-only", BodyName: "activeOnly", Description: "Return only versions effective at the requested time", Type: "bool"},
					{Name: "as-of", BodyName: "asOf", Description: "Optional ISO-8601 effective-state timestamp", Type: "string"},
					{Name: "include-archived", BodyName: "includeArchived", Description: "Include archived historical versions, which are hidden by default", Type: "bool"},
				},
			},
			{
				Name:        "get",
				Description: "Get a service price list and resolved charge rules by external GUID",
				ToolName:    "UteamupServicePriceListGet",
				RESTPath:    "{priceListGuid}",
				Args: []ArgDef{
					{Name: "priceListGuid", Description: "Service price-list external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "as-of", BodyName: "asOf", Description: "Optional ISO-8601 effective-state timestamp", Type: "string"},
				},
			},
			{
				Name:        "create",
				Description: "Create a reviewable versioned service price list",
				ToolName:    "UteamupServicePriceListCreate",
				HTTPMethod:  "POST",
				Flags:       createFlags,
			},
			{
				Name:        "update",
				Description: "Update an unlocked price-list version using reviewed evidence",
				ToolName:    "UteamupServicePriceListUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{priceListGuid}",
				Args: []ArgDef{
					{Name: "priceListGuid", Description: "Service price-list external GUID", Required: true, Type: "uuid"},
				},
				Flags: updateFlags,
			},
			{
				// Distinct from create: the server links the versions, retires the predecessor,
				// and moves its agreements onto the new rates in one transaction.
				Name:        "replacement",
				Description: "Create a future-effective replacement that retires the version it replaces",
				ToolName:    "UteamupServicePriceListCreateReplacement",
				HTTPMethod:  "POST",
				RESTPath:    "{priceListGuid}/replacement",
				Args: []ArgDef{
					{Name: "priceListGuid", Description: "External GUID of the version being replaced", Required: true, Type: "uuid"},
				},
				Flags: createFlags,
			},
			{
				// Refused with named reasons when the version still carries a linked agreement,
				// version lineage, or billing evidence.
				Name:        "delete",
				Description: "Permanently delete a price-list version that carries no evidence",
				ToolName:    "UteamupServicePriceListDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "{priceListGuid}",
				Args: []ArgDef{
					{Name: "priceListGuid", Description: "Service price-list external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "archive",
				Description: "Archive an inactive price-list version into history",
				ToolName:    "UteamupServicePriceListArchive",
				HTTPMethod:  "POST",
				RESTPath:    "{priceListGuid}/archive",
				Args: []ArgDef{
					{Name: "priceListGuid", Description: "Service price-list external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "restore",
				Description: "Return an archived price-list version to the working list",
				ToolName:    "UteamupServicePriceListRestore",
				HTTPMethod:  "POST",
				RESTPath:    "{priceListGuid}/restore",
				Args: []ArgDef{
					{Name: "priceListGuid", Description: "Service price-list external GUID", Required: true, Type: "uuid"},
				},
			},
			{
				// Read-only: runs the same rule selection a billing run uses, but records no run,
				// invoice, or billing evidence.
				Name:        "preview-rules",
				Description: "Show which price rules would apply to hypothetical evidence",
				ToolName:    "UteamupServicePriceListPreviewRules",
				HTTPMethod:  "POST",
				RESTPath:    "{priceListGuid}/rule-preview",
				Args: []ArgDef{
					{Name: "priceListGuid", Description: "Service price-list external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "work-type-guid", BodyName: "workTypeGuid", Description: "Optional work-type external GUID the labour was booked against", Type: "string"},
					{Name: "stock-item-guid", BodyName: "stockItemGuid", Description: "Optional stock-item external GUID the material came from", Type: "string"},
					{Name: "labour-hours", BodyName: "labourHours", Description: "Hypothetical approved labour hours", Default: 0.0, Type: "float"},
					{Name: "travel-hours", BodyName: "travelHours", Description: "Hypothetical approved travel hours", Default: 0.0, Type: "float"},
					{Name: "material-quantity", BodyName: "materialQuantity", Description: "Hypothetical material quantity", Default: 0.0, Type: "float"},
					{Name: "period-start", BodyName: "periodStart", Description: "Billing period start in ISO-8601 UTC", Required: true, Type: "string"},
					{Name: "period-end", BodyName: "periodEnd", Description: "Billing period end in ISO-8601 UTC", Required: true, Type: "string"},
				},
			},
		},
	})
}
