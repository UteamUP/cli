package registry

func init() {
	Register(&Domain{
		Name:        "workpermit",
		Aliases:     []string{"work-permit"},
		Description: "Work permits: biosecurity zone entry declarations and the current user's zone status",
		APIPath:     "/api/workpermit",
		Actions: []Action{
			{
				Name:        "biosecurity-status",
				Description: "Show whether a work order is in a biosecurity zone and whether you hold a valid signed entry declaration for it",
				ToolName:    "UteamupBiosecurityStatus",
				RESTPath:    "biosecurity-status",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "workorder-guid", Description: "Work order GUID", Required: true, Type: "non-empty-uuid", QueryName: "workorderGuid"},
				},
			},
			{
				Name:        "biosecurity-entry",
				Description: "Record a biosecurity entry declaration (a Draft permit with the four disinfection lines) for a zone, or return your open draft",
				ToolName:    "UteamupBiosecurityEntry",
				RESTPath:    "biosecurity-entry",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "location-guid", Description: "Biosecurity zone location GUID", Required: true, Type: "non-empty-uuid"},
					{Name: "workorder-guid", Description: "Optional work order GUID the entry is made for", Type: "uuid"},
				},
			},
			{
				Name:        "links",
				Description: "List the work orders, assets, chemicals, knowledge articles, tools and certificates linked to a work permit",
				ToolName:    "UteamupWorkPermitLinksList",
				RESTPath:    "by-guid/{workPermitGuid}/links",
				HTTPMethod:  "GET",
				Flags:       []FlagDef{workPermitFlag()},
			},
			{
				Name:        "link-add",
				Description: "Link a work order, asset, chemical, knowledge article, tool or certificate to a work permit",
				ToolName:    "UteamupWorkPermitLinkAdd",
				RESTPath:    "by-guid/{workPermitGuid}/links",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					workPermitFlag(),
					{Name: "type", Description: "WorkOrder | Asset | Chemical | KnowledgeArticle | Tool | Certificate", Required: true, Type: "string", BodyName: "linkType", AllowedValues: workPermitLinkTypes},
					{Name: "target", Description: "GUID of the record to link", Required: true, Type: "non-empty-uuid", BodyName: "targetGuid"},
				},
			},
			{
				Name:        "link-remove",
				Description: "Remove a link from a work permit",
				ToolName:    "UteamupWorkPermitLinkRemove",
				RESTPath:    "links/by-guid/{linkGuid}",
				HTTPMethod:  "DELETE",
				Flags: []FlagDef{
					{Name: "link", Description: "Work permit link GUID", Required: true, Type: "non-empty-uuid", BodyName: "linkGuid"},
				},
			},
			{
				Name:        "linked-to",
				Description: "List the work permits linked to a work order, asset, chemical, knowledge article, tool or certificate",
				ToolName:    "UteamupWorkPermitsLinkedTo",
				RESTPath:    "linked/{linkType}/{targetGuid}",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "type", Description: "workorder | asset | chemical | knowledgearticle | tool | certificate", Required: true, Type: "string", BodyName: "linkType", AllowedValues: workPermitLinkedToTypes},
					{Name: "target", Description: "GUID of the linked record", Required: true, Type: "non-empty-uuid", BodyName: "targetGuid"},
				},
			},
			{
				Name:        "prerequisite-add",
				Description: "Make another work permit a prerequisite of this work permit",
				ToolName:    "UteamupWorkPermitPrerequisiteAdd",
				RESTPath:    "by-guid/{workPermitGuid}/prerequisites",
				HTTPMethod:  "POST",
				Flags:       []FlagDef{workPermitFlag(), prerequisiteFlag()},
			},
			{
				Name:        "prerequisite-remove",
				Description: "Remove a prerequisite work permit from this work permit",
				ToolName:    "UteamupWorkPermitPrerequisiteRemove",
				RESTPath:    "by-guid/{workPermitGuid}/prerequisites/{prerequisiteGuid}",
				HTTPMethod:  "DELETE",
				Flags:       []FlagDef{workPermitFlag(), prerequisiteFlag()},
			},
			{
				Name:        "approvers",
				Description: "List who decides each approval step of a work permit and where an emailed approval stands",
				ToolName:    "UteamupWorkPermitApproversList",
				RESTPath:    "by-guid/{workPermitGuid}/approvers",
				HTTPMethod:  "GET",
				Flags:       []FlagDef{workPermitFlag()},
			},
			{
				Name:        "approver-assign",
				Description: "Name a tenant user or a contact as the approver of one step; a contact on the current step is emailed a link and code",
				ToolName:    "UteamupWorkPermitApproverAssign",
				RESTPath:    "by-guid/{workPermitGuid}/approvers",
				HTTPMethod:  "PUT",
				Flags: []FlagDef{
					workPermitFlag(),
					{Name: "step", Description: "Supervisor | SafetyOfficer | SiteManager", Required: true, Type: "string", AllowedValues: workPermitApproverSteps},
					{Name: "kind", Description: "User | Contact", Required: true, Type: "string", AllowedValues: []string{"User", "Contact"}},
					{Name: "user", Description: "Tenant user GUID when --kind User", Type: "uuid", BodyName: "userGuid"},
					{Name: "contact", Description: "Contact GUID when --kind Contact", Type: "uuid", BodyName: "contactGuid"},
				},
			},
			{
				Name:        "approver-clear",
				Description: "Remove the named approver of one step; an open emailed link stops working",
				ToolName:    "UteamupWorkPermitApproverClear",
				RESTPath:    "by-guid/{workPermitGuid}/approvers/{step}",
				HTTPMethod:  "DELETE",
				Flags:       []FlagDef{workPermitFlag(), approverStepRouteFlag()},
			},
			{
				Name:        "approver-resend",
				Description: "Email the contact on the current step a fresh approval link and code",
				ToolName:    "UteamupWorkPermitApproverResend",
				RESTPath:    "by-guid/{workPermitGuid}/approvers/{step}/resend",
				HTTPMethod:  "POST",
				Flags:       []FlagDef{workPermitFlag(), approverStepRouteFlag()},
			},
			{
				Name:        "settings",
				Description: "Show the tenant's automatic expiry setting and how many open permits the next run would expire (needs Tenant.Update)",
				ToolName:    "UteamupWorkPermitSettingsGet",
				RESTPath:    "settings",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "preview-days", QueryName: "previewDays", Description: "Preview how many open permits would expire with this day count (1-365)", Type: "int"},
				},
			},
			{
				Name:        "settings-update",
				Description: "Turn automatic expiry of open work permits on or off and set the day count (needs Tenant.Update)",
				ToolName:    "UteamupWorkPermitSettingsUpdate",
				RESTPath:    "settings",
				HTTPMethod:  "PUT",
				Flags: []FlagDef{
					{Name: "enabled", Description: "Turn automatic expiry on", Default: false, Type: "bool", BodyName: "autoExpireEnabled"},
					{Name: "days", Description: "Days after creation at which an open permit expires (1-365)", Required: true, Type: "int", BodyName: "autoExpireAfterDays"},
				},
			},
		},
	})
}

// workPermitLinkTypes is the backend link-type enum sent in the link-add body;
// workPermitLinkedToTypes is the same set as the lowercase linked/{linkType} route segment.
var (
	workPermitLinkTypes     = []string{"WorkOrder", "Asset", "Chemical", "KnowledgeArticle", "Tool", "Certificate"}
	workPermitLinkedToTypes = []string{"workorder", "asset", "chemical", "knowledgearticle", "tool", "certificate"}
)

// workPermitApproverSteps is the step enum sent in the approver-assign body;
// workPermitApproverStepSegments is the same set as the lowercase approvers/{step} route segment.
var (
	workPermitApproverSteps        = []string{"Supervisor", "SafetyOfficer", "SiteManager"}
	workPermitApproverStepSegments = []string{"supervisor", "safetyofficer", "sitemanager"}
)

func approverStepRouteFlag() FlagDef {
	return FlagDef{Name: "step", Description: "supervisor | safetyofficer | sitemanager", Required: true, Type: "string", BodyName: "step", AllowedValues: workPermitApproverStepSegments}
}

func workPermitFlag() FlagDef {
	return FlagDef{Name: "permit", Description: "Work permit GUID", Required: true, Type: "non-empty-uuid", BodyName: "workPermitGuid"}
}

func prerequisiteFlag() FlagDef {
	return FlagDef{Name: "prerequisite", Description: "Prerequisite work permit GUID", Required: true, Type: "non-empty-uuid", BodyName: "prerequisiteGuid"}
}
