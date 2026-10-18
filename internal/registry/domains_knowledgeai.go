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
				HTTPMethod: "POST", RESTBasePath: "/api/knowledgearticle", RESTPath: "by-guid/{articleGuid}/upmate-translate",
				Args: []ArgDef{{Name: "articleGuid", Description: "Knowledge page (article) GUID to translate", Required: true, Type: "string"}},
				Flags: []FlagDef{
					{Name: "languages", Short: "l", BodyName: "targetLanguages", Description: "Target language codes to fill (subset of is/pl/de/es); repeatable or comma-separated", Required: true, Type: "stringSlice"},
				}},
			{Name: "generate-from-text", Description: "Generate a new knowledge page from pasted text", ToolName: "UteamupKnowledgeAiGenerateFromText",
				// No REST endpoint generates a knowledge page from text; the MCP tool takes one request object.
				MCPOnly: true,
				Flags: []FlagDef{
					{Name: "request-file", Short: "f", BodyName: "request", Description: "JSON file with spaceGuid, text, and optional parentArticleGuid", Required: true, Type: "string", JSONFile: true},
				}},
		},
	})
}
