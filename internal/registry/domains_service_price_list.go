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
		},
	})
}
