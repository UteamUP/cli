package registry

// Upload tickets give a single-use, 15-minute URL for one file; the CLI only
// issues and reads tickets, the bytes are sent with an HTTP PUT to uploadUrl.
func init() {
	Register(&Domain{
		Name:        "document-upload",
		Aliases:     []string{"upload-ticket"},
		Description: "Get single-use upload URLs for large documents",
		Actions: []Action{
			{Name: "ticket", Description: "Get an upload URL for one file (PUT the bytes to uploadUrl)", ToolName: "UteamupDocumentCreateUploadTicket", MCPOnly: true, RejectExtraArgs: true,
				Flags: []FlagDef{
					{Name: "file-name", BodyName: "fileName", Description: "File name with extension", Required: true, Type: "string"},
					{Name: "content-type", BodyName: "contentType", Description: "MIME type, e.g. application/pdf", Required: true, Type: "string"},
					{Name: "size-bytes", BodyName: "sizeBytes", Description: "Exact file size in bytes", Required: true, Type: "int"},
					{Name: "description", Description: "Optional description", Type: "string"},
				}},
			{Name: "status", Description: "Check an upload ticket and get the document GUID", ToolName: "UteamupDocumentUploadTicketGet", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{{Name: "ticketGuid", Description: "Ticket GUID", Required: true, Type: "uuid"}}},
		},
	})
}
