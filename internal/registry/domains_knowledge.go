package registry

// knowledgeEntityTypes lists the entity types accepted by the knowledge
// entity-link endpoints (KnowledgeArticleController /linked routes).
const knowledgeEntityTypes = "asset, part, tool, chemical, location, workorder, workordertemplate, industrycode"

func knowledgeSearchPageFlags() []FlagDef {
	return []FlagDef{
		{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int", BodyName: "pageNumber"},
		{Name: "page-size", Short: "s", Description: "Items per page", Default: 20, Type: "int", BodyName: "pageSize"},
	}
}

func documentGUIDArg() []ArgDef {
	return []ArgDef{{Name: "documentGuid", Description: "Document public GUID", Required: true, Type: "string"}}
}

func init() {
	Register(&Domain{
		Name:        "knowledge",
		Aliases:     []string{"kb", "knowledge-article"},
		Description: "Manage knowledge articles",
		// Backend controller is KnowledgeArticleController → api/knowledgearticle.
		// Without this override the auto-derived base path would be /api/knowledge,
		// which matches no backend route.
		APIPath: "/api/knowledgearticle",
		// GUID-first per CLIGuidelines.md: get/update/delete target the backend
		// by-guid/{guid} routes. The int {id} routes are [Obsolete] and do not
		// survive reseeds. list/create carry no identifier so they stay on the
		// base path.
		Actions: append([]Action{
			// The backend lists and searches articles only through POST search (GET answers 405).
			{Name: "list", Description: "List records", ToolName: "UteamupKnowledgeArticleList", HTTPMethod: "POST", RESTPath: "search",
				Flags: knowledgeSearchPageFlags()},
			{Name: "get", Description: "Get an article by GUID", ToolName: "UteamupKnowledgeArticleGet", RESTPath: "by-guid/{articleGuid}",
				Args: []ArgDef{{Name: "articleGuid", Description: "Knowledge article GUID", Required: true, Type: "string"}}},
			{Name: "create", Description: "Create a record", ToolName: "UteamupKnowledgeArticleCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update an article by GUID", ToolName: "UteamupKnowledgeArticleUpdate", RESTPath: "by-guid/{articleGuid}",
				Args: []ArgDef{{Name: "articleGuid", Description: "Knowledge article GUID", Required: true, Type: "string"}}, Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Delete an article by GUID", ToolName: "UteamupKnowledgeArticleDelete", RESTPath: "by-guid/{articleGuid}",
				Args: []ArgDef{{Name: "articleGuid", Description: "Knowledge article GUID", Required: true, Type: "string"}}},
			{Name: "move-relative", Description: "Place an article beside a target in its current parent; refresh before retrying an uncertain result",
				ToolName: "UteamupKnowledgeArticleMoveRelative", HTTPMethod: "POST", RejectExtraArgs: true, RESTPath: "by-guid/{articleGuid}/move/relative",
				Args: []ArgDef{{Name: "articleGuid", Description: "Article public GUID", Required: true, Type: "non-empty-uuid"}},
				Flags: []FlagDef{
					{Name: "target-article-guid", BodyName: "targetArticleGuid", Description: "Target article public GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "place-after", BodyName: "placeAfter", Description: "Place after target (false places before)", Required: true, Type: "bool"},
					{Name: "confirm", Description: "Confirm this article placement", Required: true, Type: "bool", MustBeTrue: true, LocalOnly: true},
				}},
			{Name: "publish", Description: "Publish a draft or in-review article so space viewers can read it",
				ToolName: "UteamupKnowledgeArticlePublish", HTTPMethod: "POST", RejectExtraArgs: true, RESTPath: "by-guid/{articleGuid}/publish",
				Args: []ArgDef{{Name: "articleGuid", Description: "Article public GUID", Required: true, Type: "non-empty-uuid"}}},
			{Name: "submit-review", Description: "Submit a draft article for review; it stays hidden from space viewers until published",
				ToolName: "UteamupKnowledgeArticleSubmitReview", HTTPMethod: "POST", RejectExtraArgs: true, RESTPath: "by-guid/{articleGuid}/submit-review",
				Args: []ArgDef{{Name: "articleGuid", Description: "Article public GUID", Required: true, Type: "non-empty-uuid"}}},
			{Name: "search", Description: "Search articles", ToolName: "UteamupKnowledgeArticleSearch", HTTPMethod: "POST", RESTPath: "search",
				Args: queryArg(), Flags: knowledgeSearchPageFlags()},
			{Name: "linked", Description: "List articles linked to an entity (valid entityTypes: " + knowledgeEntityTypes + ")",
				ToolName: "UteamupKnowledgeArticleGetLinked", HTTPMethod: "GET", RESTPath: "linked/{entityType}/{entityGuid}",
				Args: []ArgDef{
					{Name: "entityType", Description: "Entity type: " + knowledgeEntityTypes, Required: true, Type: "string"},
					{Name: "entityGuid", Description: "Entity GUID", Required: true, Type: "string"},
				}},
			{Name: "link", Description: "Link an article to an entity (valid entityTypes: " + knowledgeEntityTypes + ")",
				ToolName: "UteamupKnowledgeArticleLinkEntity", HTTPMethod: "POST", RESTPath: "linked/{entityType}/{entityGuid}/{articleGuid}",
				Args: []ArgDef{
					{Name: "entityType", Description: "Entity type: " + knowledgeEntityTypes, Required: true, Type: "string"},
					{Name: "entityGuid", Description: "Entity GUID", Required: true, Type: "string"},
					{Name: "articleGuid", Description: "Knowledge article GUID", Required: true, Type: "string"},
				}},
			{Name: "unlink", Description: "Unlink an article from an entity (valid entityTypes: " + knowledgeEntityTypes + ")",
				ToolName: "UteamupKnowledgeArticleUnlinkEntity", HTTPMethod: "DELETE", RESTPath: "linked/{entityType}/{entityGuid}/{articleGuid}",
				Args: []ArgDef{
					{Name: "entityType", Description: "Entity type: " + knowledgeEntityTypes, Required: true, Type: "string"},
					{Name: "entityGuid", Description: "Entity GUID", Required: true, Type: "string"},
					{Name: "articleGuid", Description: "Knowledge article GUID", Required: true, Type: "string"},
				}},
		}, locationKnowledgeActions()...),
	})

	Register(&Domain{
		Name:        "document",
		Aliases:     []string{"documents", "doc"},
		Description: "Manage documents with versioning and archiving",
		// Every document identity crossing the CLI boundary is the canonical public
		// GUID. Integer keys remain private to EF joins in the backend.
		Actions: append([]Action{
			{Name: "list", Description: "List records", ToolName: "UteamupDocumentList", Flags: paginationFlags()},
			{Name: "get", Description: "Get a document by GUID", ToolName: "UteamupDocumentGet", RESTPath: "by-guid/{documentGuid}", Args: documentGUIDArg()},
			// Content dedup: resolve a stored document by the SHA-256 of its bytes so
			// identical content is attached rather than uploaded a second time.
			{Name: "find-by-hash", Description: "Find a document by the SHA-256 of its bytes", ToolName: "UteamupDocumentFindByHash", HTTPMethod: "GET", RESTPath: "by-hash/{sha256}", Args: []ArgDef{{Name: "sha256", Description: "Lowercase 64-character hex SHA-256 of the file bytes", Required: true, Type: "string"}}},
			{Name: "create", Description: "Create a record", ToolName: "UteamupDocumentCreate", Flags: []FlagDef{jsonFlag()}},
			{Name: "update", Description: "Update a record by GUID", ToolName: "UteamupDocumentUpdate", RESTPath: "{documentGuid}", Args: documentGUIDArg(), Flags: []FlagDef{jsonFlag()}},
			{Name: "delete", Description: "Archive a record by GUID", ToolName: "UteamupDocumentDelete", RESTPath: "{documentGuid}", Args: documentGUIDArg()},
			{Name: "list-versions", Description: "List version history for a document by GUID", ToolName: "UteamupDocumentListVersionsByGuid", HTTPMethod: "GET", RESTPath: "{documentGuid}/versions", Args: documentGUIDArg()},
			{Name: "upload-version", Description: "Upload a new version of a document by GUID", ToolName: "UteamupDocumentUploadVersionByGuid", HTTPMethod: "POST", RESTPath: "{documentGuid}/versions", Args: documentGUIDArg(), Flags: []FlagDef{{Name: "file", Short: "f", BodyName: "file", Description: "Path to the version file", Required: true, Type: "string", UploadFile: true}}},
			{Name: "restore-version", Description: "Restore a previous version as current by GUID", ToolName: "UteamupDocumentRestoreVersionByGuid", HTTPMethod: "POST", RESTPath: "{documentGuid}/versions/{versionNumber}/restore", Args: []ArgDef{{Name: "documentGuid", Description: "Document GUID", Required: true, Type: "string"}, {Name: "versionNumber", Description: "Version number to restore", Required: true, Type: "int"}}},
		},
			Action{Name: "archive", Description: "Archive (soft-delete) a document by GUID", ToolName: "UteamupDocumentArchive", HTTPMethod: "DELETE", RESTPath: "{documentGuid}", Args: documentGUIDArg()},
			Action{Name: "unarchive", Description: "Restore a document from archive by GUID", ToolName: "UteamupDocumentUnarchive", HTTPMethod: "POST", RESTPath: "{documentGuid}/unarchive", Args: documentGUIDArg()},
			Action{Name: "list-archived", Description: "List archived documents", ToolName: "UteamupDocumentListArchived", RESTPath: "archived"},
			Action{Name: "get-metadata", Description: "Get extracted metadata (EXIF / PDF DocInfo / OOXML core / camera / GPS) for a document by GUID", ToolName: "UteamupDocumentGetMetadata", HTTPMethod: "GET", RESTPath: "{guid}/metadata", Args: []ArgDef{{Name: "guid", Description: "Document public GUID", Required: true}}},
			Action{Name: "get-timeline", Description: "Get the document timeline strip (ordered by CapturedAt ASC) for a date range", ToolName: "UteamupDocumentGetTimeline", HTTPMethod: "GET", RESTPath: "timeline", Flags: []FlagDef{
				{Name: "from", Description: "Range start (ISO 8601). Default: now - 90 days.", Default: ""},
				{Name: "to", Description: "Range end (ISO 8601). Default: now.", Default: ""},
				{Name: "types", Description: "Comma-separated content-type filter", Default: ""},
				{Name: "q", Description: "Case-insensitive text match over FileName + Title", Default: ""},
				{Name: "limit", Description: "Max rows (1-10000)", Default: 5000, Type: "int"},
			}},
			Action{Name: "get-library-history", Description: "Read library membership and exact versions at an instant, with explicit unknown history", ToolName: "UteamupDocumentGetLibraryHistory", HTTPMethod: "GET", RESTPath: "library-history", Flags: []FlagDef{
				{Name: "at", Description: "Historical ISO 8601 instant; omitted on the first page to use server time"},
				{Name: "cursor", Description: "Opaque continuation from the previous page; preserves instant and scope"},
				{Name: "page-size", QueryName: "pageSize", Description: "Current readable candidates per page (1-200)", Default: 100, Type: "int"},
			}},
			Action{Name: "get-archived-library-history", Description: "Read library history including currently archived documents; requires Document.View and Document.Archive.Read", ToolName: "UteamupDocumentGetArchivedLibraryHistory", HTTPMethod: "GET", RESTPath: "library-history/archived", Flags: []FlagDef{
				{Name: "at", Description: "Historical ISO 8601 instant; omitted on the first page to use server time"},
				{Name: "cursor", Description: "Opaque continuation from the previous page; preserves instant and archive scope"},
				{Name: "page-size", QueryName: "pageSize", Description: "Current readable candidates per page (1-200)", Default: 100, Type: "int"},
			}},
		),
	})
	// journal domain moved to domains_journal.go (includes CRUD + by-code/by-asset actions)
	Register(&Domain{Name: "comment", Aliases: []string{"comments"}, Description: "Manage tenant-scoped comments", Actions: append(commentActions(), commentAssetActions()...)})
}
