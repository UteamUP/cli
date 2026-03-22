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

	Register(&Domain{Name: "document", Aliases: []string{"documents", "doc"}, Description: "Manage documents", Actions: crudActions("Document")})
	Register(&Domain{Name: "journal", Aliases: []string{"journals"}, Description: "Manage journal entries", Actions: crudActions("Journal")})
	Register(&Domain{Name: "comment", Aliases: []string{"comments"}, Description: "Manage comments", Actions: crudActions("Comment")})
}
