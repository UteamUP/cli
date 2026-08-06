package registry

// Document-review queue + acknowledgement.
// Threshold flip is SERIALIZABLE on the backend; re-ack by the same user is a no-op.

func init() {
	Register(&Domain{
		Name:        "document-review",
		Aliases:     []string{"docrev", "review"},
		Description: "Peer-review queue for imported documents",
		APIPath:     "/api/documentreview",
		Actions: []Action{
			{
				Name:        "queue",
				Description: "List documents awaiting review, paginated",
				ToolName:    "UteamupDocumentReviewQueue",
				HTTPMethod:  "GET",
				RESTPath:    "queue",
				Flags: []FlagDef{
					{Name: "page", Short: "p", QueryName: "page", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", QueryName: "pageSize", Description: "Page size (max 100)", Default: 25, Type: "int"},
					{Name: "batch-guid", Short: "b", QueryName: "batchGuid", Description: "Filter to a tenant-scoped import batch GUID", Type: "uuid"},
				},
			},
			{
				Name:        "acknowledge",
				Description: "Acknowledge a document as reviewed and return acknowledgement and reviewer GUIDs",
				ToolName:    "UteamupDocumentReviewAcknowledge",
				HTTPMethod:  "POST",
				RESTPath:    "{documentGuid}/acknowledge",
				Args: []ArgDef{
					{Name: "documentGuid", Description: "Tenant-scoped document GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "comment", Short: "c", BodyName: "comment", Description: "Optional reviewer comment", Type: "string"},
				},
			},
		},
	})
}
