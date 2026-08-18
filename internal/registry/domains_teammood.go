package registry

func init() {
	Register(&Domain{
		Name:        "teammood",
		Aliases:     []string{"team-mood", "workshop-mood"},
		Description: "Read anonymous Workshop Mood aggregates. Never dumps check-ins or named lists.",
		APIPath:     "/api/teammood",
		Actions: []Action{
			{
				Name:        "aggregates",
				Description: "Read anonymous daily team mood aggregates for one team GUID. k-anonymity applies.",
				ToolName:    "UteamupTeamMoodRead",
				HTTPMethod:  "GET",
				RESTPath:    "aggregates",
				Flags: []FlagDef{
					{Name: "team-guid", QueryName: "teamGuid", Description: "Public team GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
