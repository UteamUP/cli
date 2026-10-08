package registry

// Helpdesk cases and the tenant's intake mode. A Helpdesk-seat user submits and lists
// their own cases; a reviewer with WorkRequest.Review works the queue: start review,
// decline with a note, or upgrade to one work order, optionally from a template.
func init() {
	requestGUIDArg := []ArgDef{
		{Name: "requestGuid", Description: "Helpdesk case external GUID", Required: true, Type: "uuid"},
	}

	Register(&Domain{
		Name:        "helpdesk-request",
		Aliases:     []string{"helpdesk-requests", "hdr"},
		Description: "Submit helpdesk cases and review the helpdesk queue by GUID",
		APIPath:     "/api/helpdesk/requests",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List your own helpdesk cases, newest first",
				ToolName:    "UteamupHelpdeskRequestListOwn",
				Flags: []FlagDef{
					{Name: "page", QueryName: "page", Description: "Page number", Type: "int", Default: 1},
					{Name: "page-size", QueryName: "pageSize", Description: "Cases per page (1-100)", Type: "int", Default: 50},
				},
			},
			{
				Name:        "create",
				Description: "Submit a helpdesk case; the tenant's intake mode decides what follows",
				ToolName:    "UteamupHelpdeskRequestSubmit",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "title", Description: "Subject (max 200 characters)", Required: true, Type: "string"},
					{Name: "description", Description: "Details (max 4000 characters)", Required: true, Type: "string"},
					{Name: "idempotency-key", BodyName: "idempotencyKey", Description: "Optional key so a retried submit returns the first case", Type: "string"},
				},
			},
			{
				Name:        "queue",
				Description: "List the tenant's helpdesk cases for review (WorkRequest.Review)",
				ToolName:    "UteamupHelpdeskRequestQueue",
				HTTPMethod:  "GET",
				RESTPath:    "queue",
				Flags: []FlagDef{
					{Name: "status", QueryName: "status", Description: "submitted, inReview, approved, rejected or convertedToWorkorder", Type: "string"},
					{Name: "page", QueryName: "page", Description: "Page number", Type: "int", Default: 1},
					{Name: "page-size", QueryName: "pageSize", Description: "Cases per page (1-100)", Type: "int", Default: 50},
				},
			},
			{
				Name:        "review",
				Description: "Move a case to review or decline it with a note (WorkRequest.Review)",
				ToolName:    "UteamupHelpdeskRequestReview",
				HTTPMethod:  "PUT",
				RESTPath:    "{requestGuid}/status",
				Args:        requestGUIDArg,
				Flags: []FlagDef{
					{Name: "status", Description: "inReview or rejected", Required: true, Type: "string"},
					{Name: "review-notes", BodyName: "reviewNotes", Description: "Note to the requester (needed to decline, max 2000 characters)", Type: "string"},
				},
			},
			{
				Name:        "convert",
				Description: "Upgrade a case to one work order, optionally from a template (WorkRequest.Review)",
				ToolName:    "UteamupHelpdeskRequestConvert",
				HTTPMethod:  "POST",
				RESTPath:    "{requestGuid}/convert",
				Args:        requestGUIDArg,
				Flags: []FlagDef{
					{Name: "template-guid", BodyName: "templateGuid", Description: "Optional workorder template GUID of this tenant", Type: "uuid"},
				},
			},
		},
	})

	tenantGUIDArg := []ArgDef{
		{Name: "tenantGuid", Description: "Tenant GUID", Required: true, Type: "uuid"},
	}

	Register(&Domain{
		Name:        "helpdesk-intake",
		Description: "Read or set what happens to new helpdesk cases in a tenant",
		APIPath:     "/api/tenant",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Show the tenant's helpdesk settings, including the intake mode",
				ToolName:    "UteamupHelpdeskIntakeGet",
				HTTPMethod:  "GET",
				RESTPath:    "{tenantGuid}/helpdesk-licenses",
				Args:        tenantGUIDArg,
			},
			{
				Name:        "set",
				Description: "Set the intake mode: queue, direct, or template with --template-guid (Tenant.Update)",
				ToolName:    "UteamupHelpdeskIntakeSet",
				HTTPMethod:  "PATCH",
				RESTPath:    "{tenantGuid}/helpdesk-settings",
				Args:        tenantGUIDArg,
				Flags: []FlagDef{
					{Name: "mode", BodyName: "helpdeskIntakeMode", Description: "queue, direct or template", Required: true, Type: "string"},
					{Name: "template-guid", BodyName: "helpdeskDefaultTemplateGuid", Description: "Workorder template GUID, required for template mode", Type: "uuid"},
				},
			},
		},
	})
}
