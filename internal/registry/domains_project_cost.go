package registry

func init() {
	Register(&Domain{
		Name: "project-cost", Description: "Record manual costs with stable retry identities", APIPath: "/api/projects",
		Actions: []Action{{
			Name: "create", Description: "Create one manual line; reuse the exact requestGuid and payload after a timeout",
			ToolName: "UteamupProjectCostRecordCreate", HTTPMethod: "POST", RESTPath: "{projectGuid}/cost-records",
			Args: projectGUIDArgument, Flags: []FlagDef{projectRequestFile()},
		}},
	})
}
