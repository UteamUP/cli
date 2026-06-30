package registry

// Knowledge AI (UPMate) — AI-assisted knowledge-page operations. Credit-charged
// against the tenant's AI quota. Tool names mirror MCP/Tools/KnowledgeAiTools.cs.
// Note: there is no generate-from-documents command here — file-based generation
// is HTTP-only (multipart cannot traverse the JSON-RPC MCP channel).
func init() {
	Register(&Domain{
		Name:        "knowledgeai",
		Aliases:     []string{"upmate", "kbai", "knowledge-ai"},
		Description: "AI-assisted knowledge page generation and translation (UPMate)",
		Actions: []Action{
			{Name: "translate", Description: "Translate a knowledge page (by GUID) into the given languages", ToolName: "UteamupKnowledgeAiTranslate",
				Args: []ArgDef{{Name: "articleGuid", Description: "Knowledge page (article) GUID to translate", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "languages", Short: "l", BodyName: "targetLanguages", Description: "Target language codes to fill (subset of is/pl/de/es); repeatable or comma-separated", Required: true, Type: "stringSlice"},
				}},
			{Name: "generate-from-text", Description: "Generate a new knowledge page from pasted text", ToolName: "UteamupKnowledgeAiGenerateFromText",
				Flags: []FlagDef{
					{Name: "space-guid", BodyName: "spaceGuid", Description: "GUID of the space to create the page in", Required: true, Type: "string"},
					{Name: "text", BodyName: "text", Description: "Source text to summarize into a page", Required: true, Type: "string"},
					{Name: "parent-article-guid", BodyName: "parentArticleGuid", Description: "Optional parent page GUID to nest the new page under", Type: "string"},
				}},
		},
	})
}
