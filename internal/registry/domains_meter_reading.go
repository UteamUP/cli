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
		// Routes mirror MeterReadingController. Without APIPath the REST fallback derived
		// /api/meterreading — no such controller, so every verb 404'd.
		APIPath: "/api/assets",
		Actions: []Action{
			{
				Name:        "current",
				Description: "Get current (latest) meter values for an asset",
				ToolName:    "UteamupMeterreadingGetCurrent",
				HTTPMethod:  "GET",
				RESTPath:    "{assetGuid}/meter-readings/current",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "attributes",
				Description: "Get the full attribute snapshot (static + meter) for an asset",
				ToolName:    "UteamupMeterreadingGetAttributes",
				HTTPMethod:  "GET",
				RESTPath:    "{assetGuid}/attributes",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "history",
				Description: "Get paginated reading history for a specific meter attribute",
				ToolName:    "UteamupMeterreadingGetHistory",
				HTTPMethod:  "GET",
				RESTPath:    "{assetGuid}/meter-readings/{attributeDefinitionGuid}/history",
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
				HTTPMethod:  "POST",
				RESTPath:    "{assetGuid}/meter-readings",
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
				Name:        "ocr",
				Description: "Analyze a meter photo and return a review-only OCR reading suggestion",
				ToolName:    "UteamupMeterreadingPhotoOcr",
				RESTPath:    "{assetGuid}/meter-readings/{attributeDefinitionGuid}/ocr",
				HTTPMethod:  "POST",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
					{Name: "attribute-definition-guid", Description: "Attribute definition external Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "photo", Description: "Path to meter photo", Required: true, Type: "string", UploadFile: true},
				},
			},
			{
				Name:        "update-attributes",
				Description: "Upsert (create or update) attribute values for an asset",
				ToolName:    "UteamupMeterreadingUpdateAttributes",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "asset-guid", BodyName: "assetGuid", Description: "Asset external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "request-file", Short: "f", BodyName: "request", Description: "JSON file containing {\"values\":[{\"attributeDefinitionGuid\":\"…\",\"rawValue\":\"42\"}]}", Required: true, Type: "string", JSONFile: true},
				},
			},
		},
	})
}
