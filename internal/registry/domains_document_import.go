package registry

// Document-import CLI surface — mirrors the MCP tools in
// UteamUP_API/MCP/Tools/DocumentImportTools.cs. Multipart upload and batch
// commit remain intentionally outside this read-only CLI surface.

func init() {
	Register(&Domain{
		Name:        "document-import",
		Aliases:     []string{"docimp", "import"},
		Description: "Inspect bulk image/document import batches produced by the web/mobile uploader",
		APIPath:     "/api/documentimport",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get a document-import batch with its items + AI suggestions",
				ToolName:    "UteamupDocumentImportGetBatch",
				HTTPMethod:  "GET",
				RESTPath:    "batch/{batchGuid}",
				Args: []ArgDef{
					{Name: "batchGuid", Description: "Tenant-scoped import batch GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
