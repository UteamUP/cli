package registry

func init() {
	Register(&Domain{
		Name:        "quality",
		Aliases:     []string{"qms", "quality-management"},
		Description: "Browse enterprise Quality records and immutable governance ledgers",
		APIPath:     "/api/quality/governance",
		Actions: []Action{
			{
				Name:        "list",
				Description: "Search the permission-filtered tenant Quality index",
				ToolName:    "UteamupQualityRecordsSearch",
				HTTPMethod:  "GET",
				RESTPath:    "records",
				Flags: []FlagDef{
					{Name: "record-kind", QueryName: "recordKind", Description: "Optional Quality record kind name", Type: "string"},
					{Name: "site-guid", QueryName: "siteGuid", Description: "Optional owning-site public GUID", Type: "string"},
					{Name: "retention-status", QueryName: "retentionDispositionStatus", Description: "Optional Active, ReviewDue, or Archived state", Type: "string"},
					{Name: "integrity-status", QueryName: "historyIntegrityStatus", Description: "Optional NotVerified, Verified, or Failed state", Type: "string"},
					{Name: "search", QueryName: "search", Description: "Display-number prefix or exact record GUID", Type: "string"},
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page, maximum 100", Default: 25, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Read one immutable Quality activity, decision, evidence, and escalation ledger",
				ToolName:    "UteamupQualityRecordLedgerGet",
				HTTPMethod:  "GET",
				RESTPath:    "records/{recordKind}/{domainRecordGuid}",
				Args: []ArgDef{
					{Name: "recordKind", Description: "Quality record kind name", Required: true, Type: "string"},
					{Name: "domainRecordGuid", Description: "Quality domain-record public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:         "cleanup-deadletters",
				Description:  "List tenant-scoped document-storage cleanup dead letters",
				ToolName:     "UteamupQualityCleanupDeadLettersList",
				HTTPMethod:   "GET",
				RESTBasePath: "/api/quality/operations/storage-cleanup/dead-letters",
				Flags: []FlagDef{
					{Name: "page", QueryName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", QueryName: "pageSize", Description: "Results per page", Default: 25, Type: "int"},
				},
			},
			{
				Name:         "cleanup-deadletter-get",
				Description:  "Inspect one document-storage cleanup dead letter and its redrive preview",
				ToolName:     "UteamupQualityCleanupDeadLetterGet",
				HTTPMethod:   "GET",
				RESTBasePath: "/api/quality/operations/storage-cleanup/dead-letters",
				RESTPath:     "{cleanupJobGuid}",
				Args: []ArgDef{
					{Name: "cleanupJobGuid", Description: "Cleanup-job public GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:         "cleanup-deadletter-redrive",
				Description:  "Explicitly redrive one server-approved Azure cleanup dead letter",
				ToolName:     "UteamupQualityCleanupDeadLetterRedrive",
				HTTPMethod:   "POST",
				RESTBasePath: "/api/quality/operations/storage-cleanup/dead-letters",
				RESTPath:     "{cleanupJobGuid}/redrive",
				Args: []ArgDef{
					{Name: "cleanupJobGuid", Description: "Cleanup-job public GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "preview-token", BodyName: "previewToken", Description: "Unexpired preview token from cleanup-deadletter-get", Required: true, Sensitive: true, Type: "string"},
					{Name: "reason", BodyName: "reason", Description: "Auditable operator reason for the redrive", Required: true, Type: "string"},
					{Name: "confirm", BodyName: "confirmed", Description: "Explicitly confirm the reviewed redrive", Required: true, Type: "bool", MustBeTrue: true},
					{Name: "idempotency-key", HeaderName: "Idempotency-Key", Description: "Caller-generated GUID reused only for the same redrive", Required: true, Type: "uuid"},
					{Name: "concurrency-token", HeaderName: "If-Match", Description: "Current concurrency token from cleanup-deadletter-get", Required: true, Sensitive: true, Type: "string"},
				},
			},
		},
	})
}
