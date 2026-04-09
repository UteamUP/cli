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
		Actions: append(crudActions("Document"),
			Action{Name: "list-versions", Description: "List version history for a document", ToolName: "UteamupDocumentListVersions", Args: []ArgDef{{Name: "documentId", Description: "Document ID", Required: true}}},
			Action{Name: "upload-version", Description: "Upload a new version of a document", ToolName: "UteamupDocumentUploadVersion", Args: []ArgDef{{Name: "documentId", Description: "Document ID", Required: true}}, Flags: []FlagDef{{Name: "file", Description: "Path to file", Default: ""}, {Name: "notes", Description: "Change notes", Default: ""}}},
			Action{Name: "restore-version", Description: "Restore a previous version as current", ToolName: "UteamupDocumentRestoreVersion", Args: []ArgDef{{Name: "documentId", Description: "Document ID", Required: true}, {Name: "versionNumber", Description: "Version number to restore", Required: true}}},
			Action{Name: "archive", Description: "Archive (soft-delete) a document", ToolName: "UteamupDocumentArchive", RESTPath: "archive", Args: []ArgDef{{Name: "id", Description: "Document ID", Required: true}}},
			Action{Name: "unarchive", Description: "Restore a document from archive", ToolName: "UteamupDocumentUnarchive", RESTPath: "unarchive", Args: []ArgDef{{Name: "id", Description: "Document ID", Required: true}}},
			Action{Name: "list-archived", Description: "List archived documents", ToolName: "UteamupDocumentListArchived", RESTPath: "archived"},
		),
	})
	// journal domain moved to domains_journal.go (includes CRUD + by-code/by-asset actions)
	Register(&Domain{Name: "comment", Aliases: []string{"comments"}, Description: "Manage comments", Actions: crudActions("Comment")})
}
