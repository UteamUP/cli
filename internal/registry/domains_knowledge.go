package registry

func init() {
	Register(&Domain{
		Name:        "knowledge",
		Aliases:     []string{"kb", "knowledge-article"},
		Description: "Manage knowledge articles",
		Actions: append(crudActions("KnowledgeArticle"),
			Action{Name: "search", Description: "Search articles", ToolName: "UteamupKnowledgeArticleSearch", Args: queryArg(), Flags: paginationFlags()},
		),
	})

	Register(&Domain{
		Name:        "document",
		Aliases:     []string{"documents", "doc"},
		Description: "Manage documents with versioning and archiving",
		// list/get/create stay as-is (get keeps the legacy {id:int} route per the
		// document GUID-first contract). The lifecycle verbs below are GUID-keyed
		// against the new /api/document/{externalGuid}/... routes; the int routes
		// remain as [Obsolete] shims on the backend.
		Actions: append([]Action{
			{Name: "list", Description: "List records", ToolName: "UteamupDocumentList", Flags: paginationFlags()},
			{Name: "get", Description: "Get by ID", ToolName: "UteamupDocumentGet", Args: idArg()},
			{Name: "create", Description: "Create a record", ToolName: "UteamupDocumentCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a record by GUID", ToolName: "UteamupDocumentUpdate", Args: externalGuidArg(), Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete a record by GUID", ToolName: "UteamupDocumentDelete", Args: externalGuidArg()},
			{Name: "list-versions", Description: "List version history for a document by GUID", ToolName: "UteamupDocumentListVersions", HTTPMethod: "GET", RESTPath: "{externalGuid}/versions", Args: externalGuidArg()},
			{Name: "upload-version", Description: "Upload a new version of a document by GUID", ToolName: "UteamupDocumentUploadVersion", HTTPMethod: "POST", RESTPath: "{externalGuid}/versions", Args: externalGuidArg(), Flags: []FlagDef{{Name: "file", Description: "Path to file", Default: ""}, {Name: "notes", Description: "Change notes", Default: ""}}},
			{Name: "restore-version", Description: "Restore a previous version as current by GUID", ToolName: "UteamupDocumentRestoreVersion", HTTPMethod: "POST", RESTPath: "{externalGuid}/versions/{versionNumber}/restore", Args: []ArgDef{{Name: "externalGuid", Description: "Document GUID", Required: true, Type: "string"}, {Name: "versionNumber", Description: "Version number to restore", Required: true, Type: "int"}}},
		},
			Action{Name: "archive", Description: "Archive (soft-delete) a document", ToolName: "UteamupDocumentArchive", RESTPath: "archive", Args: []ArgDef{{Name: "id", Description: "Document ID", Required: true}}},
			Action{Name: "unarchive", Description: "Restore a document from archive", ToolName: "UteamupDocumentUnarchive", RESTPath: "unarchive", Args: []ArgDef{{Name: "id", Description: "Document ID", Required: true}}},
			Action{Name: "list-archived", Description: "List archived documents", ToolName: "UteamupDocumentListArchived", RESTPath: "archived"},
			Action{Name: "get-metadata", Description: "Get extracted metadata (EXIF / PDF DocInfo / OOXML core / camera / GPS) for a document by GUID", ToolName: "UteamupDocumentGetMetadata", Args: []ArgDef{{Name: "guid", Description: "Document public GUID", Required: true}}},
			Action{Name: "get-timeline", Description: "Get the document timeline strip (ordered by CapturedAt ASC) for a date range", ToolName: "UteamupDocumentGetTimeline", Flags: []FlagDef{
				{Name: "from", Description: "Range start (ISO 8601). Default: now - 90 days.", Default: ""},
				{Name: "to", Description: "Range end (ISO 8601). Default: now.", Default: ""},
				{Name: "types", Description: "Comma-separated content-type filter", Default: ""},
				{Name: "q", Description: "Case-insensitive text match over FileName + Title", Default: ""},
				{Name: "limit", Description: "Max rows (1-10000)", Default: 5000, Type: "int"},
			}},
		),
	})
	// journal domain moved to domains_journal.go (includes CRUD + by-code/by-asset actions)
	Register(&Domain{Name: "comment", Aliases: []string{"comments"}, Description: "Manage comments", Actions: crudActions("Comment")})
}
