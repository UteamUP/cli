package registry

func init() {
	Register(&Domain{
		Name:        "improvement-project",
		Aliases:     []string{"improvement", "imp-project"},
		Description: "Manage improvement projects",
		Actions:     crudActions("ImprovementProject"),
	})

	Register(&Domain{
		Name:        "kaizen-card",
		Aliases:     []string{"kaizen", "kc"},
		Description: "Manage kaizen cards",
		Actions:     crudActions("KaizenCard"),
	})

	Register(&Domain{
		Name:        "improvement-suggestion",
		Aliases:     []string{"suggestion", "imp-suggestion"},
		Description: "Manage improvement suggestions",
		Actions:     crudActions("ImprovementSuggestion"),
	})
}
