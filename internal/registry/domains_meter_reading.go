package registry

// Mirrors the MCP UteamupMeterreading* tools backed by
// MeterReadingController on the backend. The endpoints are GUID-first —
// every command takes the asset's external Guid (and, where applicable,
// the attribute definition's external Guid) rather than the internal int ids.
func init() {
	Register(&Domain{
		Name:        "meter-reading",
		Aliases:     []string{"meter-readings", "mr"},
		Description: "Read and record meter values on assets",
		Actions: []Action{
			{
				Name:        "current",
				Description: "Get current (latest) meter values for an asset",
				ToolName:    "UteamupMeterreadingGetCurrent",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "attributes",
				Description: "Get the full attribute snapshot (static + meter) for an asset",
				ToolName:    "UteamupMeterreadingGetAttributes",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "history",
				Description: "Get paginated reading history for a specific meter attribute",
				ToolName:    "UteamupMeterreadingGetHistory",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
					{Name: "attribute-definition-guid", Description: "Attribute definition external Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page (max 1000)", Default: 50, Type: "int"},
					{Name: "from", Description: "Start date filter (ISO 8601)", Type: "string"},
					{Name: "to", Description: "End date filter (ISO 8601)", Type: "string"},
				},
			},
			{
				Name:        "record",
				Description: "Record a manual meter reading",
				ToolName:    "UteamupMeterreadingRecord",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "attribute-definition-guid", Description: "Attribute definition external Guid", Required: true, Type: "string"},
					{Name: "value", Description: "Reading value (numeric)", Required: true, Type: "float"},
					{Name: "timestamp", Description: "Reading timestamp (ISO 8601, defaults to now)", Type: "string"},
					{Name: "notes", Description: "Optional notes", Type: "string"},
				},
			},
			{
				Name:        "update-attributes",
				Description: "Upsert (create or update) attribute values for an asset",
				ToolName:    "UteamupMeterreadingUpdateAttributes",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "values-json", Description: "Attribute values as JSON array (e.g. '[{\"attributeDefinitionId\":1,\"rawValue\":\"42\"}]')", Required: true, Type: "string"},
				},
			},
		},
	})
}
