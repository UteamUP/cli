package registry

func init() {
	Register(&Domain{
		Name:        "oncall",
		Aliases:     []string{"on-call"},
		Description: "Read the on-call rota",
		APIPath:     "/api/oncall",
		Actions: []Action{
			{
				Name:        "who",
				Description: "Who is on call for a schedule at an instant (defaults to now)",
				ToolName:    "UteamupOnCallWho",
				HTTPMethod:  "GET",
				RESTPath:    "{schedule-guid}/who",
				Args: []ArgDef{
					{Name: "schedule-guid", Description: "On-call schedule external Guid", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "at", Description: "Instant to resolve (ISO-8601 UTC). Defaults to now.", Type: "string"},
				},
			},
		},
	})
}
