package registry

func qualityAuditPublicGUIDArgument(name, description string) ArgDef {
	return ArgDef{
		Name:        name,
		Description: description,
		Required:    true,
		Type:        "non-empty-uuid",
	}
}

func qualityAuditPublicGUIDQueryFlag(name, queryName, description string) FlagDef {
	return FlagDef{
		Name:        name,
		QueryName:   queryName,
		Description: description,
		Type:        "non-empty-uuid",
	}
}

func qualityAuditCreateMutationFlags(recordLabel string) []FlagDef {
	return []FlagDef{
		qualityAuditRequestFileFlag(recordLabel),
		qualityAuditIdempotencyFlag(),
	}
}

func qualityAuditExistingMutationFlags(recordLabel string, requireConfirmation bool) []FlagDef {
	flags := []FlagDef{
		qualityAuditRequestFileFlag(recordLabel),
		qualityAuditIdempotencyFlag(),
		qualityAuditConcurrencyFlag(recordLabel),
	}
	if requireConfirmation {
		flags = append(flags, qualityAuditConfirmationFlag())
	}
	return flags
}

func qualityAuditRequestFileFlag(recordLabel string) FlagDef {
	return FlagDef{
		Name:               "request-file",
		Short:              "f",
		Description:        "Path to the exact " + recordLabel + " request DTO as one root JSON object",
		Required:           true,
		Type:               "string",
		RootJSONObjectFile: true,
	}
}

func qualityAuditIdempotencyFlag() FlagDef {
	return FlagDef{
		Name:        "idempotency-key",
		Description: "Opaque 8-128 byte caller key reused only for the exact same mutation",
		Required:    true,
		Type:        "string",
		HeaderName:  "Idempotency-Key",
	}
}

func qualityAuditConcurrencyFlag(recordLabel string) FlagDef {
	return FlagDef{
		Name:        "concurrency-token",
		Description: "Current opaque concurrency token from the " + recordLabel + " response",
		Required:    true,
		Type:        "string",
		Sensitive:   true,
		HeaderName:  "If-Match",
		StrongETag:  true,
	}
}

func qualityAuditConfirmationFlag() FlagDef {
	return FlagDef{
		Name:        "confirm",
		Description: "Explicitly confirm the reviewed retained mutation",
		Required:    true,
		Type:        "bool",
		MustBeTrue:  true,
		LocalOnly:   true,
	}
}
