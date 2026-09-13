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
				Description: "Record peer review of an exact document version; this is not technical acceptance",
				ToolName:    "UteamupDocumentReviewAcknowledge",
				HTTPMethod:  "POST",
				RESTPath:    "{documentGuid}/acknowledge",
				Args: []ArgDef{
					{Name: "documentGuid", Description: "Tenant-scoped document GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "comment", Short: "c", BodyName: "comment", Description: "Optional reviewer comment", Type: "string"},
					{Name: "document-version-guid", BodyName: "documentVersionGuid", Description: "Exact observed version GUID, required for versioned documents", Type: "uuid"},
					{Name: "expected-document-updated-at", BodyName: "expectedDocumentUpdatedAt", Description: "Observed document timestamp, required with an exact version", Type: "string"},
				},
			},
			{
				Name: "receipt", Description: "Read the original active or archived exact peer-review receipt",
				ToolName: "UteamupDocumentReviewAcknowledgment", HTTPMethod: "GET",
				RESTPath: "{documentGuid}/acknowledgments/{acknowledgmentGuid}",
				Args: []ArgDef{
					{Name: "documentGuid", Description: "Tenant-scoped document GUID", Required: true, Type: "uuid"},
					{Name: "acknowledgmentGuid", Description: "Original acknowledgment GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
