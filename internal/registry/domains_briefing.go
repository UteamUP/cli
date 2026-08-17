package registry

func init() {
	Register(&Domain{
		Name:        "briefing",
		Aliases:     []string{"daily-briefing"},
		Description: "Read the scheduled daily role-pack briefing and confirm one HITL decision",
		APIPath:     "/api/dailybriefing",
		Actions: []Action{
			{
				Name:        "today",
				Description: "Read today's daily role-pack briefing (transcript, never SMS dispatch)",
				ToolName:    "UteamupBriefingDailyRead",
				HTTPMethod:  "GET",
				RESTPath:    "today",
			},
			{
				Name:        "get",
				Description: "Read one daily briefing by public GUID",
				ToolName:    "UteamupBriefingDailyRead",
				HTTPMethod:  "GET",
				RESTPath:    "{briefingGuid}",
				Args: []ArgDef{
					{Name: "briefingGuid", Description: "Public briefing GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "confirm",
				Description: "Confirm one HITL briefing decision after WorkOrder.Create recheck. SMS is not dispatched.",
				ToolName:    "UteamupBriefingDecisionPropose",
				HTTPMethod:  "POST",
				RESTPath:    "{briefingGuid}/confirm",
				Args: []ArgDef{
					{Name: "briefingGuid", Description: "Public briefing GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "option-key", BodyName: "optionKey", Description: "createWorkorder or scheduleInspection", Required: true, Type: "string"},
				},
			},
		},
	})
}
