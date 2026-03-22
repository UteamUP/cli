package registry

func init() {
	Register(&Domain{Name: "workforce-group", Aliases: []string{"wg"}, Description: "Manage workforce groups", Actions: crudActions("WorkforceGroup")})
	Register(&Domain{Name: "workforce-training", Description: "Manage workforce group required training", Actions: crudActions("WorkforceGroupRequiredTraining")})
	Register(&Domain{Name: "workforce-planning", Aliases: []string{"wp"}, Description: "Manage workforce planning", Actions: crudActions("WorkforcePlanning")})
	Register(&Domain{Name: "skill", Aliases: []string{"skills"}, Description: "Manage skills", Actions: crudActions("Skill")})
	Register(&Domain{Name: "team", Aliases: []string{"teams"}, Description: "Manage teams", Actions: crudActions("Team")})
}
