package registry

func init() {
	Register(&Domain{Name: "workorder-template", Aliases: []string{"wot"}, Description: "Manage work order templates", Actions: crudActions("WorkorderTemplate")})
	Register(&Domain{Name: "workorder-signature", Description: "Manage work order signatures", Actions: crudActions("WorkorderSignature")})
	Register(&Domain{Name: "workorder-watchlist", Description: "Manage work order watchlists", Actions: crudActions("WorkorderWatchlist")})
	Register(&Domain{Name: "tasklist", Aliases: []string{"tasks"}, Description: "Manage task lists", Actions: crudActions("TaskList")})
	Register(&Domain{Name: "checklist", Aliases: []string{"checklists"}, Description: "Manage checklists", Actions: crudActions("CheckList")})
}
