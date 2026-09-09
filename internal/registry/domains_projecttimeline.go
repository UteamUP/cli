package registry

func init() {
	for _, item := range []struct{ kind, guid, tool string }{
		{"workorders", "workorderGuid", "UteamupProjectTimelineWorkorderSchedule"},
		{"stages", "stageGuid", "UteamupProjectTimelineStageSchedule"},
	} {
		Register(&Domain{Name: "project-timeline-" + item.kind, Description: "Change project schedule dates without overwriting other fields", APIPath: "/api/projects", Actions: []Action{
			{Name: "update", HTTPMethod: "PUT", ToolName: item.tool, Description: "Save reviewed dates using the latest updatedAt revision", RESTPath: "{projectGuid}/timeline/" + item.kind + "/{" + item.guid + "}", Args: []ArgDef{
				{Name: "projectGuid", Description: "Project GUID", Required: true, Type: "non-empty-uuid"},
				{Name: item.guid, Description: "Linked item GUID", Required: true, Type: "non-empty-uuid"},
			}, Flags: []FlagDef{
				{Name: "start-date", Description: "ISO 8601 start; omit for a stage milestone", Required: item.kind == "workorders", Type: "string"},
				{Name: "due-date", Description: "ISO 8601 finish", Required: true, Type: "string"},
				{Name: "expected-updated-at", Description: "Latest updatedAt returned by the server", Required: true, Type: "string"},
			}},
		}})
	}
}
