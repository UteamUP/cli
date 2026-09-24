package registry

func init() {
	Register(&Domain{
		Name:        "asset",
		Aliases:     []string{"assets"},
		Description: "Manage assets and equipment inventory",
		Actions: []Action{
			{
				Name:        "ask",
				Description: "Ask a source-grounded question about an asset (2 AI credits; requires Asset.View)",
				ToolName:    "UteamupAssetAsk",
				HTTPMethod:  "POST",
				RESTPath:    "{assetGuid}/ask",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset ExternalGuid", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "question", Short: "q", Description: "Question to answer from records linked to the asset", Required: true, Type: "string"},
				},
			},
			{
				Name:        "list",
				Description: "List all assets with pagination and filtering",
				ToolName:    "UteamupAssetList",
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
					{Name: "filter", Short: "f", Description: "Filter by name", Type: "string"},
					{Name: "sort-by", Description: "Sort field (Name, CreatedAt, etc.)", Default: "Name", Type: "string"},
					{Name: "sort-order", Description: "Sort direction (asc or desc)", Default: "asc", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get asset details by ID",
				ToolName:    "UteamupAssetGet",
				Args:        []ArgDef{{Name: "id", Description: "Asset ID", Required: true, Type: "int"}},
			},
			{
				Name:        "get-by-guid",
				Description: "Get asset details by stable ExternalGuid (URL-safe, survives migrations)",
				ToolName:    "UteamupAssetGetByGuid",
				Args:        []ArgDef{{Name: "guid", Description: "Asset ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "duplicate",
				Description: "Duplicate a coded asset — deep-copies it and assigns the next code instance (e.g. GE0101 → GE0102). Requires Asset.Create permission.",
				ToolName:    "UteamupAssetDuplicate",
				HTTPMethod:  "POST",
				RESTPath:    "by-guid/{assetGuid}/duplicate",
				Args:        []ArgDef{{Name: "assetGuid", Description: "ExternalGuid of the source asset to duplicate (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "get-assigned-stock",
				Description: "Get parts/tools/chemicals assigned to an asset, with the stock locations each item sits in (quantity + low-stock state)",
				ToolName:    "UteamupAssetGetAssignedStock",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "get-documents-aggregated",
				Description: "Get ALL documents for an asset grouped by source — the asset's own documents plus its assigned parts', tools', chemicals', and industry code's documents (each tagged with sourceType/sourceGuid/sourceName)",
				ToolName:    "UteamupAssetGetAggregatedDocuments",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
			},
			{
				Name:        "create",
				Description: "Create an asset from a reviewed GUID-only JSON model",
				ToolName:    "UteamupAssetCreate",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "request-file", Short: "f", BodyName: "model", Description: "JSON file containing the GUID-only asset creation model", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "update",
				Description: "Update an existing asset by public GUID",
				ToolName:    "UteamupAssetUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{assetGuid}",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset public GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "name", Description: "New asset name", Type: "string"},
					{Name: "serial", Description: "New serial number", Type: "string"},
					{Name: "asset-type-guids", Description: "Replacement asset type GUIDs (repeatable or comma-separated)", Type: "stringSlice"},
					{Name: "primary-asset-type-guid", Description: "Primary asset type GUID (must be included in --asset-type-guids)", Type: "uuid"},
				},
			},
			{
				// Unlike update (a full replace), patch sends only the flags you pass.
				Name:        "patch",
				Description: "Change only the fields you pass on an asset by public GUID. An empty string clears a text field; 00000000-0000-0000-0000-000000000000 clears a category, location or floor.",
				ToolName:    "UteamupAssetPatch",
				HTTPMethod:  "PATCH",
				RESTPath:    "by-guid/{assetGuid}",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset public GUID", Required: true, Type: "non-empty-uuid"}},
				Flags: []FlagDef{
					{Name: "name", Description: "New asset name (2 to 512 characters)", Type: "string"},
					{Name: "serial-number", Description: "New serial number (empty clears it)", Type: "string"},
					{Name: "reference-number", Description: "New reference number (empty clears it)", Type: "string"},
					{Name: "model-number", Description: "New model number (empty clears it)", Type: "string"},
					{Name: "category-guid", Description: "Category GUID (the empty GUID clears it)", Type: "uuid"},
					{Name: "location-guid", Description: "Location GUID (the empty GUID clears it; without --location-floor-guid the floor is cleared)", Type: "uuid"},
					{Name: "location-floor-guid", Description: "Floor or room GUID within the location (the empty GUID clears it)", Type: "uuid"},
					// No Default: a bool flag with a default is always sent, which would flip the asset.
					{Name: "is-active", Description: "Set active (true) or inactive (false)", Type: "bool"},
					{Name: "expected-updated-at-utc", Description: "The asset's updatedAt as last read; a newer stored value fails with asset_changed", Type: "string"},
				},
			},
			{
				Name:        "get-specs",
				Description: "Get the effective attribute definitions (operating specs) for an asset, grouped per asset type",
				ToolName:    "UteamupAssetGetEffectiveAttributeDefinitions",
				RESTPath:    "by-guid/{assetGuid}/effective-attribute-definitions",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset public GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "delete",
				Description: "Delete an asset by ID",
				ToolName:    "UteamupAssetDelete",
				Args:        []ArgDef{{Name: "id", Description: "Asset ID", Required: true, Type: "int"}},
			},
			{
				Name:        "search",
				Description: "Search assets by name or serial number",
				ToolName:    "UteamupAssetSearch",
				Args:        []ArgDef{{Name: "query", Description: "Search term", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "set-responsible-owners",
				Description: "Set the responsible owners of an asset (replace-set; pass all owner user ids).",
				ToolName:    "UteamupAssetSetResponsibleOwners",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{assetGuid}/responsible-owners",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "user-ids", Description: "Responsible owner user ids — repeatable or comma-separated (replaces the current set)", Type: "stringSlice", BodyName: "userIds"},
				},
			},
			{
				Name:        "edit-code-assignment",
				Description: "Edit a coded asset's code assignment by ExternalGuid — rename the asset (--name), rename its code (--desired-code), reassign it to a different catalog entry, or move it under a different parent. Requires Asset.UpdateCodeAssignment permission.",
				ToolName:    "UteamupAssetEditCodeAssignment",
				HTTPMethod:  "PATCH",
				RESTPath:    "by-guid/{assetGuid}/codeassignment",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset ExternalGuid (format: 00000000-0000-0000-0000-000000000000)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "name", Description: "New asset name (unique per tenant, case-insensitive)", Type: "string"},
					{Name: "desired-code", Description: "Rename the asset's code to this exact segment (e.g. A01 → A001, 1LBA1 → 01LBA1)", Type: "string"},
					{Name: "code-catalog-entry-guid", Description: "Reassign the asset to a different catalog entry by its ExternalGuid", Type: "string"},
					{Name: "parent-asset-guid", Description: "Move the asset under a different parent asset by its ExternalGuid", Type: "string"},
					{Name: "demote-to-root", Description: "Promote the asset to root scope (clear its parent)", Default: false, Type: "bool"},
				},
			},
		},
	})
}
