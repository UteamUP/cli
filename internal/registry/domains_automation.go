package registry

func init() {
	// Routes mirror AutomationController (/api/automation/...) and
	// AutomationRuntimeController (/api/automation/{runs,catalog,settings}). Every
	// identifier is a GUID; the int-keyed twins are not exposed here.
	automationGuid := ArgDef{Name: "externalGuid", Description: "Automation GUID", Required: true, Type: "non-empty-uuid"}
	runGuid := ArgDef{Name: "runGuid", Description: "Run GUID", Required: true, Type: "non-empty-uuid"}
	approvalGuid := ArgDef{Name: "requestGuid", Description: "Approval request GUID", Required: true, Type: "non-empty-uuid"}

	Register(&Domain{Name: "automation", Aliases: []string{"automations", "workflow", "workflows"}, Description: "Manage automations, publish workflows and follow their runs", APIPath: "/api/automation", Actions: []Action{
		{Name: "list", HTTPMethod: "POST", Description: "List automations (paged)", ToolName: "UteamupAutomationSearch", RESTPath: "search", Args: []ArgDef{
			{Name: "nameFilter", Description: "Partial, case-insensitive name filter", Type: "string"},
			{Name: "pageNumber", Description: "Page number (default 1)", Type: "int"},
			{Name: "pageSize", Description: "Page size (default 10, max 100)", Type: "int"},
		}},
		{Name: "get", HTTPMethod: "GET", Description: "Get one automation", ToolName: "UteamupAutomationGet", RESTPath: "by-guid/{externalGuid}", Args: []ArgDef{automationGuid}},
		{Name: "workflow-get", HTTPMethod: "GET", Description: "Get the draft workflow graph", ToolName: "UteamupAutomationWorkflowGet", RESTPath: "by-guid/{externalGuid}/workflow", Args: []ArgDef{automationGuid}},
		{Name: "workflow-save", HTTPMethod: "PUT", Description: "Save the draft workflow graph", ToolName: "UteamupAutomationWorkflowSave", RESTPath: "by-guid/{externalGuid}/workflow", Args: []ArgDef{
			automationGuid,
			{Name: "workflowDefinitionJson", Description: "Workflow graph JSON", Required: true, Type: "string"},
		}},
		{Name: "workflow-validate", HTTPMethod: "POST", Description: "Validate the draft without publishing", ToolName: "UteamupAutomationWorkflowValidate", RESTPath: "by-guid/{externalGuid}/validate", Args: []ArgDef{
			automationGuid,
			{Name: "workflowDefinitionJson", Description: "Graph JSON to validate instead of the saved draft", Type: "string"},
		}},
		{Name: "workflow-publish", HTTPMethod: "POST", Description: "Publish the draft so the runtime executes it", ToolName: "UteamupAutomationWorkflowPublish", RESTPath: "by-guid/{externalGuid}/publish", Args: []ArgDef{
			automationGuid,
			{Name: "note", Description: "Version note", Type: "string"},
		}},
		{Name: "state", HTTPMethod: "GET", Description: "Published version, failures and secrets state", ToolName: "UteamupAutomationState", RESTPath: "by-guid/{externalGuid}/state", Args: []ArgDef{automationGuid}},
		{Name: "versions", HTTPMethod: "GET", Description: "Published version history", ToolName: "UteamupAutomationVersions", RESTPath: "by-guid/{externalGuid}/versions", Args: []ArgDef{automationGuid}},
		{Name: "trigger", HTTPMethod: "POST", Description: "Start one automation by hand", ToolName: "UteamupAutomationTrigger", RESTPath: "by-guid/{externalGuid}/trigger", Args: []ArgDef{
			automationGuid,
			{Name: "entityType", Description: "Entity type the run is about (e.g. Workorder)", Type: "string"},
			{Name: "entityGuid", Description: "GUID of that entity", Type: "uuid"},
		}},
		{Name: "runs", HTTPMethod: "POST", Description: "List workflow runs (paged, newest first)", ToolName: "UteamupAutomationRunsList", RESTPath: "runs/search", Args: []ArgDef{
			{Name: "status", Description: "Run status filter (running, waiting, succeeded, failed, ...)", Type: "string"},
			{Name: "automationGuid", Description: "Only runs of this automation", Type: "uuid"},
			{Name: "pageNumber", Description: "Page number (default 1)", Type: "int"},
			{Name: "pageSize", Description: "Page size (default 25, max 100)", Type: "int"},
		}},
		{Name: "run-get", HTTPMethod: "GET", Description: "One run with its step timeline", ToolName: "UteamupAutomationRunGet", RESTPath: "runs/{runGuid}", Args: []ArgDef{runGuid}},
		{Name: "run-cancel", HTTPMethod: "POST", Description: "Stop a run that is still running or waiting", ToolName: "UteamupAutomationRunCancel", RESTPath: "runs/{runGuid}/cancel", Args: []ArgDef{runGuid}},
		{Name: "run-stats", HTTPMethod: "GET", Description: "Per-day outcome counts and live counters", ToolName: "UteamupAutomationRunStats", RESTPath: "runs/stats", Args: []ArgDef{
			{Name: "days", Description: "Window in days (default 14, max 90)", Type: "int", QueryName: "days"},
		}},
		{Name: "catalog", HTTPMethod: "GET", Description: "Triggers, actions and node types the builder may use", ToolName: "UteamupAutomationCatalog", RESTPath: "catalog"},
		{Name: "settings", HTTPMethod: "GET", Description: "Tenant automation settings and limits", ToolName: "UteamupAutomationSettings", RESTPath: "settings"},
		{Name: "pause", HTTPMethod: "POST", Description: "Pause every automation for the tenant", ToolName: "UteamupAutomationPause", RESTPath: "settings/pause", Args: []ArgDef{
			{Name: "reason", Description: "Why automations are paused", Type: "string"},
		}},
		{Name: "resume", HTTPMethod: "POST", Description: "Resume automations for the tenant", ToolName: "UteamupAutomationResume", RESTPath: "settings/resume"},
		{Name: "approvals", HTTPMethod: "GET", Description: "Workflow approvals the caller may decide (paged, newest first)", ToolName: "UteamupAutomationApprovalsList", RESTPath: "approvals", Args: []ArgDef{
			{Name: "status", Description: "Status filter (pending, approved, rejected, expired)", Type: "string", QueryName: "status"},
			{Name: "page", Description: "Page number (default 1)", Type: "int", QueryName: "page"},
			{Name: "pageSize", Description: "Page size (default 20, max 100)", Type: "int", QueryName: "pageSize"},
		}},
		{Name: "approval-get", HTTPMethod: "GET", Description: "One approval request", ToolName: "UteamupAutomationApprovalGet", RESTPath: "approvals/{requestGuid}", Args: []ArgDef{approvalGuid}},
		{Name: "approval-decide", HTTPMethod: "POST", Description: "Approve or reject a waiting workflow step", ToolName: "UteamupAutomationApprovalDecide", RESTPath: "approvals/{requestGuid}/decide", Args: []ArgDef{
			approvalGuid,
			{Name: "approve", Description: "true to approve, false to reject", Required: true, Type: "bool"},
			{Name: "comment", Description: "Comment recorded with the decision", Type: "string"},
		}},
	}})
}
