package registry

func serviceSLATransitionFlags(reasonRequired bool) []FlagDef {
	return []FlagDef{
		{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped transition idempotency UUID", Required: true, Type: "string"},
		{Name: "expected-updated-at", BodyName: "expectedUpdatedAt", Description: "Exact reviewed UpdatedAt timestamp", Required: true, Type: "string"},
		{Name: "occurred-at", BodyName: "occurredAt", Description: "Optional ISO-8601 transition timestamp", Type: "string"},
		{Name: "reason", BodyName: "reason", Description: "Reviewed transition reason or note", Required: reasonRequired, Type: "string"},
	}
}

func init() {
	Register(&Domain{
		Name:        "service-sla",
		Aliases:     []string{"service-sla-milestones", "sla-milestone"},
		Description: "Inspect and explicitly manage service SLA milestone evidence",
		APIPath:     "/api/service-sla-milestones",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List SLA targets, status, remaining time, delta, and source evidence",
				ToolName:    "UteamupServiceSlaMilestoneList",
				Flags: []FlagDef{
					{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Optional workorder external GUID", Type: "string"},
					{Name: "agreement-guid", BodyName: "agreementGuid", Description: "Optional service agreement external GUID", Type: "string"},
					{Name: "status", BodyName: "status", Description: "Optional derived SLA status", Type: "string"},
					{Name: "search", BodyName: "search", Description: "Optional free-text filter over workorder name or ticket, customer, or agreement title", Type: "string"},
					{Name: "as-of", BodyName: "asOf", Description: "Optional ISO-8601 status evaluation timestamp", Type: "string"},
					{Name: "page", BodyName: "page", Description: "One-based page number", Default: 1, Type: "int"},
					{Name: "page-size", BodyName: "pageSize", Description: "Results per page, maximum 200", Default: 50, Type: "int"},
				},
			},
			{
				Name:        "get",
				Description: "Get one SLA milestone and its current, target, delta, source, and freshness evidence",
				ToolName:    "UteamupServiceSlaMilestoneGet",
				RESTPath:    "{milestoneGuid}",
				Args: []ArgDef{
					{Name: "milestoneGuid", Description: "SLA milestone external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "as-of", BodyName: "asOf", Description: "Optional ISO-8601 status evaluation timestamp", Type: "string"},
				},
			},
			{
				Name:        "initialize",
				Description: "Initialize response and resolution evidence for a covered workorder",
				ToolName:    "UteamupServiceSlaMilestoneInitialize",
				HTTPMethod:  "POST",
				RESTPath:    "workorders/{workorderGuid}/initialize",
				Args: []ArgDef{
					{Name: "workorderGuid", Description: "Covered workorder external GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped initialization idempotency UUID", Required: true, Type: "string"},
					{Name: "started-at", BodyName: "startedAt", Description: "Optional ISO-8601 milestone start timestamp", Type: "string"},
				},
			},
			{
				Name:        "pause",
				Description: "Pause an open SLA milestone with an explicit reviewed reason",
				ToolName:    "UteamupServiceSlaMilestonePause",
				HTTPMethod:  "POST",
				RESTPath:    "{milestoneGuid}/pause",
				Args: []ArgDef{
					{Name: "milestoneGuid", Description: "SLA milestone external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceSLATransitionFlags(true),
			},
			{
				Name:        "resume",
				Description: "Resume a paused SLA milestone and preserve exact paused duration",
				ToolName:    "UteamupServiceSlaMilestoneResume",
				HTTPMethod:  "POST",
				RESTPath:    "{milestoneGuid}/resume",
				Args: []ArgDef{
					{Name: "milestoneGuid", Description: "SLA milestone external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceSLATransitionFlags(false),
			},
			{
				Name:        "complete",
				Description: "Complete an SLA milestone with reviewed version evidence",
				ToolName:    "UteamupServiceSlaMilestoneComplete",
				HTTPMethod:  "POST",
				RESTPath:    "{milestoneGuid}/complete",
				Args: []ArgDef{
					{Name: "milestoneGuid", Description: "SLA milestone external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceSLATransitionFlags(false),
			},
			{
				Name:        "cancel",
				Description: "Cancel an open SLA milestone with an explicit reviewed reason",
				ToolName:    "UteamupServiceSlaMilestoneCancel",
				HTTPMethod:  "POST",
				RESTPath:    "{milestoneGuid}/cancel",
				Args: []ArgDef{
					{Name: "milestoneGuid", Description: "SLA milestone external GUID", Required: true, Type: "uuid"},
				},
				Flags: serviceSLATransitionFlags(true),
			},
			{
				Name:        "reconcile",
				Description: "Persist reviewed warning and breach thresholds for one workorder",
				ToolName:    "UteamupServiceSlaMilestoneReconcile",
				HTTPMethod:  "POST",
				RESTPath:    "reconcile",
				Flags: []FlagDef{
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Tenant-scoped reconciliation idempotency UUID", Required: true, Type: "string"},
					{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Workorder external GUID", Required: true, Type: "string"},
					{Name: "as-of", BodyName: "asOf", Description: "ISO-8601 reviewed reconciliation timestamp", Required: true, Type: "string"},
				},
			},
		},
	})
}
