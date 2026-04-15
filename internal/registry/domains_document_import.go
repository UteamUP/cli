package registry

// Document-import CLI surface — mirrors the MCP tools in
// UteamUP_API/MCP/Tools/DocumentImportTools.cs. Covers review + ack flows;
// multipart upload / commit are intentionally HTTP-only.

func init() {
	Register(&Domain{
		Name:        "document-import",
		Aliases:     []string{"docimp", "import"},
		Description: "Inspect bulk image/document import batches produced by the web/mobile uploader",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get a document-import batch with its items + AI suggestions",
				ToolName:    "UteamupDocumentImportGetBatch",
				Flags: []FlagDef{
					{Name: "batch-id", Short: "b", Description: "Batch ID", Required: true, Type: "int"},
				},
			},
		},
	})
}
