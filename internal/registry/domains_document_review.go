package registry

// Document-review queue + acknowledgement.
// Threshold flip is SERIALIZABLE on the backend; re-ack by the same user is a no-op.

func init() {
	Register(&Domain{
		Name:        "document-review",
		Aliases:     []string{"docrev", "review"},
		Description: "Peer-review queue for imported documents",
		Actions: []Action{
			{
				Name:        "queue",
				Description: "List documents awaiting review, paginated",
				ToolName:    "UteamupDocumentReviewQueue",
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Page size (max 100)", Default: 25, Type: "int"},
					{Name: "batch-id", Short: "b", Description: "Filter to a single import batch", Type: "int"},
				},
			},
			{
				Name:        "acknowledge",
				Description: "Acknowledge a document as reviewed (idempotent)",
				ToolName:    "UteamupDocumentReviewAcknowledge",
				Flags: []FlagDef{
					{Name: "document-id", Short: "d", Description: "Document ID", Required: true, Type: "int"},
					{Name: "comment", Short: "c", Description: "Optional reviewer comment", Type: "string"},
				},
			},
		},
	})
}
