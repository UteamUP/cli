package registry

// UPMate agents: event-started analysts that gather evidence, spend AI credits
// within a budget and end in a proposal a person approves. Mirrors
// UpmateAgentController (api/upmateagent) and the MCP automation tools.
func init() {
	agentGuid := ArgDef{Name: "agentGuid", Description: "Agent GUID", Required: true, Type: "non-empty-uuid"}
	runGuid := ArgDef{Name: "runGuid", Description: "Agent run GUID", Required: true, Type: "non-empty-uuid"}

	Register(&Domain{Name: "upmate-agent", Aliases: []string{"upmate-agents", "agent", "agents"}, Description: "Manage UPMate agents and follow their runs", APIPath: "/api/upmateagent", Actions: []Action{
		{Name: "list", HTTPMethod: "GET", Description: "List UPMate agents with their budgets and today's spend", ToolName: "UteamupUpmateAgentsList", RESTPath: ""},
		{Name: "get", HTTPMethod: "GET", Description: "Get one UPMate agent", ToolName: "UteamupUpmateAgentGet", RESTPath: "by-guid/{agentGuid}", Args: []ArgDef{agentGuid}},
		{Name: "capabilities", HTTPMethod: "GET", Description: "Capabilities an agent may be granted", ToolName: "UteamupUpmateAgentCapabilities", RESTPath: "capabilities"},
		{Name: "run", HTTPMethod: "POST", Description: "Start an agent now about one entity (spends AI credits within its budget)", ToolName: "UteamupUpmateAgentRun", RESTPath: "by-guid/{agentGuid}/run", Args: []ArgDef{
			agentGuid,
			{Name: "entityType", Description: "Entity type the run is about (e.g. Workorder)", Type: "string"},
			{Name: "entityGuid", Description: "GUID of that entity", Type: "uuid"},
			{Name: "requestGuid", Description: "Idempotency key; the same key never starts a second run", Type: "uuid"},
		}},
		{Name: "runs", HTTPMethod: "GET", Description: "Runs of one agent (paged, newest first)", ToolName: "UteamupUpmateAgentRunsList", RESTPath: "by-guid/{agentGuid}/runs", Args: []ArgDef{
			agentGuid,
			{Name: "page", Description: "Page number (default 1)", Type: "int", QueryName: "page"},
			{Name: "pageSize", Description: "Page size (default 20, max 100)", Type: "int", QueryName: "pageSize"},
		}},
		{Name: "run-get", HTTPMethod: "GET", Description: "One agent run with its evidence and proposal", ToolName: "UteamupUpmateAgentRunGet", RESTPath: "runs/{runGuid}", Args: []ArgDef{runGuid}},
		{Name: "run-cancel", HTTPMethod: "POST", Description: "Stop an agent run that is still working", ToolName: "UteamupUpmateAgentRunCancel", RESTPath: "runs/{runGuid}/cancel", Args: []ArgDef{runGuid}},
	}})
}
