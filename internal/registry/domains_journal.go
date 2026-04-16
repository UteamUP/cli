package registry

func init() {
	Register(&Domain{
		Name:        "journal",
		Aliases:     []string{"journals"},
		Description: "Manage journal entries, import documents, and query by code / asset / workorder",
		Actions: append(crudActions("Journal"),
			Action{
				Name:        "by-code",
				Description: "List journal entries linked to a code catalog entry",
				ToolName:    "UteamupJournalByCode",
				RESTPath:    "by-code",
				Args:        []ArgDef{{Name: "code-catalog-entry-id", Description: "Code catalog entry ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
				},
			},
			Action{
				Name:        "by-asset",
				Description: "List journal entries for an asset",
				ToolName:    "UteamupJournalByAsset",
				RESTPath:    "by-asset",
				Args:        []ArgDef{{Name: "asset-id", Description: "Asset ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
				},
			},
			// --- Import pipeline (journal-document-import-and-inline-tags) ---
			// These wrap the POST /api/journal/import and /{guid}/images endpoints
			// via base64 payloads so MCP/CLI callers don't need multipart plumbing.
			// The underlying MCP tool converts the base64 body into the same
			// multipart shape the HTTP endpoint accepts.
			Action{
				Name:        "import",
				Description: "Import a .docx / .md / .txt file (base64) into a new journal; auto-tags KKS/Asset/Workorder tokens",
				ToolName:    "UteamupJournalImport",
				RESTPath:    "import",
				Args: []ArgDef{
					{Name: "file-name", Description: "Source filename (extension drives MIME detection)", Required: true, Type: "string"},
					{Name: "file-content-base64", Description: "Base64-encoded file content (max 10 MB)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "title", Description: "Journal title (defaults to filename)", Type: "string"},
					{Name: "summary", Description: "Optional summary for the journal list surface", Type: "string"},
					{Name: "target-journal-guid", Description: "Append to an existing journal instead of creating a new one", Type: "uuid"},
				},
			},
			Action{
				Name:        "create-from-image",
				Description: "Create a stub journal from one image (base64); re-encoded server-side to strip EXIF",
				ToolName:    "UteamupJournalCreateFromImage",
				RESTPath:    "import",
				Args: []ArgDef{
					{Name: "image-file-name", Description: "Source image filename", Required: true, Type: "string"},
					{Name: "image-content-base64", Description: "Base64-encoded image bytes (max 25 MB, png/jpeg/webp/gif)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "title", Description: "Journal title (defaults to timestamp)", Type: "string"},
				},
			},
			// --- Mention search helpers ---
			// Read-only autocomplete endpoints matching what the web journal
			// editor hits for the `#`, `$`, and `%` triggers. Lets scripts
			// pre-resolve mention IDs before calling the update endpoint.
			Action{
				Name:        "search-assets",
				Description: "Search assets for the $ mention trigger (tenant-scoped, active only)",
				ToolName:    "UteamupAssetMentionSearch",
				RESTPath:    "",
				Args:        []ArgDef{{Name: "query", Description: "Search query (min 1 char)", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "limit", Short: "l", Description: "Max results (server caps at 20)", Default: 8, Type: "int"},
				},
			},
			Action{
				Name:        "search-workorders",
				Description: "Search workorders by TicketId for the % mention trigger (tenant-scoped)",
				ToolName:    "UteamupWorkorderMentionSearch",
				RESTPath:    "",
				Args:        []ArgDef{{Name: "query", Description: "Search query against Workorder.TicketId", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "limit", Short: "l", Description: "Max results (server caps at 20)", Default: 8, Type: "int"},
				},
			},
		),
	})
}
