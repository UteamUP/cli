package registry

func init() {
	Register(&Domain{
		Name:        "rotationtemplate",
		Aliases:     []string{"rotation-template", "rota-template"},
		Description: "Standard shift rotation templates (4-on-4-off, Pitman, …)",
		APIPath:     "/api/rotationtemplate",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the available standard rotation templates",
				ToolName:    "UteamupRotationTemplateList",
				HTTPMethod:  "GET",
			},
			{
				Name:        "build",
				Description: "Materialise a rotation template into a concrete pattern definition",
				ToolName:    "UteamupRotationTemplateBuild",
				HTTPMethod:  "POST",
				RESTPath:    "{key}/build",
				Args: []ArgDef{
					{Name: "key", Description: "Template key (e.g. four-on-four-off, pitman-223)", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "anchor", Description: "Anchor date (ISO-8601)", Type: "string", Required: true, BodyName: "anchorDate"},
					{Name: "day-shift", Description: "Day-shift external Guid", Type: "string", Required: true, BodyName: "dayShiftGuid"},
					{Name: "night-shift", Description: "Night-shift external Guid", Type: "string", Required: true, BodyName: "nightShiftGuid"},
				},
			},
		},
	})
}
