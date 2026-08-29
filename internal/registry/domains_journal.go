package registry

func init() {
	Register(&Domain{
		Name:        "journal",
		Aliases:     []string{"journals"},
		Description: "Manage journal entries, import documents, and query by code / asset / workorder",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List journal entries with pagination",
				ToolName:    "UteamupJournalList",
				HTTPMethod:  "POST",
				RESTPath:    "search",
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int", BodyName: "pageNumber"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 20, Type: "int", BodyName: "pageSize"},
				},
			},
			{
				Name:        "get",
				Description: "Get a journal entry by GUID",
				ToolName:    "UteamupJournalGet",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{journalGuid}",
				Args:        []ArgDef{{Name: "journalGuid", Description: "Journal GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "create",
				Description: "Create a journal or GUID-linked field note from a reviewed JSON file",
				ToolName:    "UteamupJournalCreate",
				MCPOnly:     true,
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "JSON file containing the JournalCreateModel request", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "update",
				Description: "Update a journal entry by GUID from a reviewed JSON file",
				ToolName:    "UteamupJournalUpdate",
				MCPOnly:     true,
				Args:        []ArgDef{{Name: "journalGuid", Description: "Journal GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "JSON file containing the JournalUpdateModel request", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a journal entry by GUID",
				ToolName:    "UteamupJournalDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "by-guid/{journalGuid}",
				Args:        []ArgDef{{Name: "journalGuid", Description: "Journal GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "by-code",
				Description: "List journal entries linked to a code catalog entry",
				ToolName:    "UteamupJournalGetByCode",
				HTTPMethod:  "GET",
				RESTPath:    "by-code/{codeCatalogEntryGuid}",
				Args:        []ArgDef{{Name: "codeCatalogEntryGuid", Description: "Code catalog entry GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int", QueryName: "pageNumber"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 20, Type: "int", QueryName: "pageSize"},
				},
			},
			{
				Name:        "by-asset",
				Description: "List journal entries for an asset",
				ToolName:    "UteamupJournalGetByAsset",
				HTTPMethod:  "GET",
				RESTPath:    "by-asset/{assetGuid}",
				Args:        []ArgDef{{Name: "assetGuid", Description: "Asset GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int", QueryName: "pageNumber"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 20, Type: "int", QueryName: "pageSize"},
				},
			},
			{
				Name:        "import",
				Description: "Import a .docx / .md / .txt file (base64) into a new journal; auto-tags KKS/Asset/Workorder tokens",
				ToolName:    "UteamupJournalImport",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "fileName", Description: "Source filename (extension drives MIME detection)", Required: true, Type: "string"},
					{Name: "fileContentBase64", Description: "Base64-encoded file content (max 10 MB)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "title", BodyName: "title", Description: "Journal title (defaults to filename)", Type: "string"},
					{Name: "summary", BodyName: "summary", Description: "Optional summary for the journal list surface", Type: "string"},
					{Name: "target-journal-guid", BodyName: "targetJournalGuid", Description: "Append to an existing journal instead of creating a new one", Type: "uuid"},
				},
			},
			{
				Name:        "create-from-image",
				Description: "Create a stub journal from one image (base64); re-encoded server-side to strip EXIF",
				ToolName:    "UteamupJournalCreateFromImage",
				MCPOnly:     true,
				Args: []ArgDef{
					{Name: "imageFileName", Description: "Source image filename", Required: true, Type: "string"},
					{Name: "imageContentBase64", Description: "Base64-encoded image bytes (max 25 MB, png/jpeg/webp/gif)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "title", BodyName: "title", Description: "Journal title (defaults to timestamp)", Type: "string"},
				},
			},
			{
				Name:         "search-assets",
				Description:  "Search assets for the $ mention trigger (tenant-scoped, active only)",
				ToolName:     "UteamupAssetMentionSearch",
				RESTBasePath: "/api/asset",
				RESTPath:     "mention-search",
				HTTPMethod:   "GET",
				Args:         []ArgDef{{Name: "query", Description: "Search query (min 1 char)", Required: true, Type: "string", QueryName: "query"}},
				Flags: []FlagDef{
					{Name: "limit", Short: "l", Description: "Max results (server caps at 20)", Default: 8, Type: "int", QueryName: "limit"},
				},
			},
			{
				Name:         "search-workorders",
				Description:  "Search workorders by TicketId for the % mention trigger (tenant-scoped)",
				ToolName:     "UteamupWorkorderMentionSearch",
				RESTBasePath: "/api/workorder",
				RESTPath:     "mention-search",
				HTTPMethod:   "GET",
				Args:         []ArgDef{{Name: "query", Description: "Search query against Workorder.TicketId", Required: true, Type: "string", QueryName: "query"}},
				Flags: []FlagDef{
					{Name: "limit", Short: "l", Description: "Max results (server caps at 20)", Default: 8, Type: "int", QueryName: "limit"},
				},
			},
		},
	})
}
