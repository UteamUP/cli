package registry

func init() {
	Register(&Domain{
		Name:        "workingtime",
		Aliases:     []string{"working-time", "wt"},
		Description: "Manage working-time compliance rule sets",
		APIPath:     "/api/workingtime",
		Actions: []Action{
			{
				Name:        "ruleset-list",
				Description: "List the tenant's working-time rule sets",
				ToolName:    "UteamupWorkingTimeRuleSetList",
				HTTPMethod:  "GET",
				RESTPath:    "rulesets",
			},
			{
				Name:        "ruleset-create",
				Description: "Create a working-time rule set",
				ToolName:    "UteamupWorkingTimeRuleSetCreate",
				HTTPMethod:  "POST",
				RESTPath:    "rulesets",
				Flags: []FlagDef{
					{Name: "name", Description: "Rule-set name", Type: "string", Required: true},
					{Name: "country", Description: "ISO country code (e.g. IS, US)", Type: "string", BodyName: "countryCode"},
				},
			},
			{
				Name:        "project-overtime",
				Description: "Pre-publish what-if overtime projection for rostered + worked hours",
				ToolName:    "UteamupWorkingTimeProjectOvertime",
				HTTPMethod:  "POST",
				RESTPath:    "project-overtime",
				Flags: []FlagDef{
					{Name: "rostered", Description: "Rostered hours", Type: "float", Required: true, BodyName: "rosteredHours"},
					{Name: "worked", Description: "Already-worked hours", Type: "float", Default: 0.0, BodyName: "workedHours"},
					{Name: "weekly-limit", Description: "Weekly hour limit (0 = engine default)", Type: "float", Default: 0.0, BodyName: "weeklyLimitHours"},
					{Name: "overtime-max-minutes", Description: "Cap on overtime minutes (0 = uncapped)", Type: "int", Default: 0, BodyName: "overtimeMaxMinutes"},
				},
			},
		},
	})
}
