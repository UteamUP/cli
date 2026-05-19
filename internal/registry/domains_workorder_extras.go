package registry

func init() {
	Register(&Domain{Name: "workorder-template", Aliases: []string{"wot"}, Description: "Manage work order templates", Actions: append(crudActions("WorkorderTemplate"), Action{
		Name:        "create-workorder",
		Description: "Create an open work order from a template, identified by the template's public GUID. No asset or resolution note required.",
		ToolName:    "UteamupWorkorderTemplateCreateFromTemplateByGuid",
		Flags: []FlagDef{
			{Name: "template", Description: "Public GUID of the workorder template (required)", Required: true, Type: "string"},
			{Name: "name", Description: "Optional override for the new work order name", Type: "string"},
			{Name: "description", Description: "Optional override for the new work order description", Type: "string"},
			{Name: "priority", Description: "Optional priority override (1=Low … 5=Critical)", Type: "int"},
			{Name: "notes", Description: "Optional notes override", Type: "string"},
		},
	})})
	Register(&Domain{Name: "workorder-signature", Description: "Manage work order signatures", Actions: crudActions("WorkorderSignature")})
	Register(&Domain{Name: "workorder-watchlist", Description: "Manage work order watchlists", Actions: crudActions("WorkorderWatchlist")})
	Register(&Domain{Name: "tasklist", Aliases: []string{"tasks"}, Description: "Manage task lists", Actions: crudActions("TaskList")})
	Register(&Domain{Name: "checklist", Aliases: []string{"checklists"}, Description: "Manage checklists", Actions: crudActions("CheckList")})
}
