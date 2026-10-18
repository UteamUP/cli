package registry

func init() {
	Register(&Domain{
		Name:        "workorder-template-prerequisite",
		Aliases:     []string{"wot-prerequisite", "wot-prereq"},
		Description: "Templates that must be done before work created from a template",
		APIPath:     "/api/workordertemplate",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the templates a template requires first",
				ToolName:    "UteamupWorkorderTemplatePrerequisiteList",
				HTTPMethod:  "GET",
				RESTPath:    "by-guid/{templateGuid}/prerequisites",
				Args: []ArgDef{
					{Name: "templateGuid", Description: "Public GUID of the workorder template", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "set",
				Description: "Replace the templates a template requires first (max 10, no loops); omit the flag to clear",
				ToolName:    "UteamupWorkorderTemplatePrerequisiteSet",
				HTTPMethod:  "PUT",
				RESTPath:    "by-guid/{templateGuid}/prerequisites",
				Args: []ArgDef{
					{Name: "templateGuid", Description: "Public GUID of the workorder template", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "prerequisite-template-guids", BodyName: "prerequisiteTemplateGuids", Description: "Ordered prerequisite template GUIDs (comma-separated)", Type: "stringSlice"},
				},
			},
		},
	})
}
