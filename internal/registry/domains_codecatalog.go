package registry

func init() {
	Register(&Domain{
		Name:        "codecatalog",
		Aliases:     []string{"cc", "codes"},
		Description: "Search, update, and deactivate industry code catalog entries",
		APIPath:     "/api/codecatalog",
		Actions: []Action{
			{
				Name:        "search",
				Description: "Search industry codes by text (mention search)",
				ToolName:    "UteamupCodeCatalogSearch",
				RESTPath:    "mention-search",
				Args:        []ArgDef{{Name: "query", Description: "Search term (e.g., 'pump', '1-HLA')", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "limit", Short: "l", Description: "Maximum results to return", Default: 10, Type: "int"},
				},
			},
			{
				Name:        "update-by-guid",
				Description: "Update a code catalog entry by stable ExternalGuid",
				ToolName:    "UteamupCodingsystemUpdateEntryByGuid",
				Args: []ArgDef{
					{Name: "guid", Description: "Entry ExternalGuid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "name", Description: "New display name", Type: "string"},
					{Name: "description", Description: "New description", Type: "string"},
					{Name: "component-type-code", Description: "New component type code", Type: "string"},
					{Name: "component-type-description", Description: "New component type description", Type: "string"},
					{Name: "drawing-reference", Description: "New drawing reference", Type: "string"},
				},
			},
			{
				Name:        "deactivate-by-guid",
				Description: "Deactivate a code catalog entry by ExternalGuid (cascades to descendants)",
				ToolName:    "UteamupCodingsystemDeactivateEntryByGuid",
				Args: []ArgDef{
					{Name: "guid", Description: "Entry ExternalGuid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "remove-asset-assignment",
				Description: "Remove the code assignment from an asset by its ExternalGuid — preserves audit log",
				ToolName:    "UteamupCodingsystemRemoveAssetAssignment",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset ExternalGuid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "assign-asset",
				Description: "Assign a code-catalog entry to an asset by their ExternalGuids — GUID-first; preserves audit log. Folded in from prod bug 81e76313.",
				ToolName:    "UteamupCodingsystemAssignAsset",
				Args: []ArgDef{
					{Name: "asset-guid", Description: "Asset ExternalGuid", Required: true, Type: "string"},
					{Name: "code-catalog-entry-guid", Description: "Target code-catalog entry ExternalGuid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "history",
				Description: "Fetch the full chronological history timeline for one industry code — images, documents, work orders, asset edits, journals, and inventory additions. Cursor-paginated.",
				ToolName:    "UteamupCodecatalogHistory",
				Args: []ArgDef{
					{Name: "code-guid", Description: "External Guid of the code whose timeline to fetch", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "types", Description: "Comma-separated EntityType filter (Image,Document,Workorder,Asset,Journal,AssetPart,AssetTool,AssetChemical)", Type: "string"},
					{Name: "actor-guid", Description: "Optional actor external Guid filter", Type: "string"},
					{Name: "from-utc", Description: "Inclusive lower bound on event timestamp (ISO-8601 UTC)", Type: "string"},
					{Name: "to-utc", Description: "Inclusive upper bound on event timestamp (ISO-8601 UTC)", Type: "string"},
					{Name: "q", Description: "Free-text fragment matched against actor name + entity name + journal preview", Type: "string"},
					{Name: "cursor", Description: "Opaque cursor from a prior response's nextCursor; omit for the first page", Type: "string"},
					{Name: "page-size", Description: "Page size (clamped 1..100 server-side)", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "designation",
				Description: "Assemble the IEC 81346-2 / RDS-PP full designation (e.g. =MAA10+UNIT01-CT001:5) for a code catalog entry by walking its parent chain. Returns the assembled string plus a per-segment breakdown showing each ancestor's aspect (Function / Location / Product / Terminal) and prefix character.",
				ToolName:    "UteamupCodecatalogDesignation",
				HTTPMethod:  "GET",
				RESTPath:    "entries/by-guid/{guid}/designation",
				Args: []ArgDef{
					{Name: "guid", Description: "Code catalog entry ExternalGuid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "identification-letters",
				Description: "List the canonical IEC 81346-2 single-character identification letters (A B C E F G K P Q R S T U V W X) with their English labels. Pure constant lookup — used by the frontend useIec81346Letters composable for parity validation.",
				ToolName:    "UteamupCodecatalogIdentificationLetters",
				HTTPMethod:  "GET",
				RESTPath:    "identification-letters",
			},
		},
	})
}
