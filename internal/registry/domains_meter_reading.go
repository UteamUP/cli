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
					// BodyName is load-bearing: the backend binds ReadingValue/ReadingTimestamp,
					// and the camelCase default ("value"/"timestamp") silently recorded 0.
					{Name: "value", BodyName: "readingValue", Description: "Reading value (numeric)", Required: true, Type: "float"},
					{Name: "timestamp", BodyName: "readingTimestamp", Description: "Reading timestamp (ISO 8601, defaults to now)", Type: "string"},
					{Name: "notes", Description: "Optional notes", Type: "string"},
				},
			},
			{
				Name:        "record-statutory",
				Description: "Record a STATUTORY odometer reading (kilometragjald, Act 100/2025) with its legal trigger context",
				ToolName:    "UteamupMeterreadingRecordStatutory",
				HTTPMethod:  "POST",
				RESTPath:    "{assetGuid}/meter-readings/{attributeDefinitionGuid}/statutory",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
					{Name: "attribute-definition-guid", Description: "Odometer attribute definition external Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "value", BodyName: "readingValue", Description: "Reading value (numeric)", Required: true, Type: "float"},
					{Name: "context", Description: "Legal trigger: periodicCadence|rentalCheckout|rentalReturn|ownershipChange|deregistration|reregistration|temporaryExport|seasonalDeactivation", Required: true, Type: "string"},
					{Name: "observed-at", BodyName: "observedAt", Description: "Observation time (ISO 8601, defaults to now)", Type: "string"},
					{Name: "evidence-url", BodyName: "evidenceDocumentUrl", Description: "Optional evidence document URL", Type: "string"},
					{Name: "notes", Description: "Optional notes", Type: "string"},
				},
			},
			{
				Name:        "correct",
				Description: "Correct a reading via the append-only correction chain (original never mutated)",
				ToolName:    "UteamupMeterreadingCorrect",
				HTTPMethod:  "POST",
				RESTPath:    "{assetGuid}/meter-readings/by-guid/{readingGuid}/correct",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
					{Name: "reading-guid", Description: "Guid of the reading being corrected", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "value", BodyName: "correctedValue", Description: "Corrected reading value (numeric)", Required: true, Type: "float"},
					{Name: "notes", Description: "Optional correction notes", Type: "string"},
				},
			},
			{
				Name:        "replace-meter",
				Description: "Register a physical meter/odometer replacement (bumps the meter generation)",
				ToolName:    "UteamupMeterreadingReplaceMeter",
				HTTPMethod:  "POST",
				RESTPath:    "{assetGuid}/meter-readings/{attributeDefinitionGuid}/replace-meter",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset external Guid", Required: true, Type: "string"},
					{Name: "attribute-definition-guid", Description: "Meter attribute definition external Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "initial-value", BodyName: "initialValue", Description: "Reading shown on the replacement meter (defaults to 0)", Default: 0.0, Type: "float"},
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
