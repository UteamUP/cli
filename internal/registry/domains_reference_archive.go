package registry

// withReferenceArchive adds `archive` and `restore` to a reference domain whose IsActive flag is its archive
// (skill, certificate, currency, shift, workorder-template, asset-maintenance-plan). Marking a record inactive
// keeps it readable by GUID and in every history row that names it; its list hides it unless --is-active=false.
// Both routes need the same permission as deleting the record.
func withReferenceArchive(actions []Action, label, plural, permission string) []Action {
	for index := range actions {
		if actions[index].Name == "list" {
			actions[index].Description += "; inactive " + plural + " are hidden unless you pass --is-active=false"
			actions[index].Flags = append(actions[index].Flags, FlagDef{
				Name: "is-active", QueryName: "isActive", Type: "bool",
				Description: "false lists only inactive " + plural + "; omit it for active ones",
			})
		}
	}
	return append(actions,
		Action{
			Name: "archive",
			Description: "Mark the " + label + " inactive: it leaves lists, searches and pickers and stays readable " +
				"in its history. Needs " + permission,
			ToolName:        "ArchiveReferenceRecord",
			HTTPMethod:      "POST",
			RESTPath:        "by-guid/{recordGuid}/archive",
			RejectExtraArgs: true,
			Args: []ArgDef{
				{Name: "recordGuid", Description: "GUID of the " + label + " to mark inactive", Required: true, Type: "non-empty-uuid"},
			},
		},
		Action{
			Name:            "restore",
			Description:     "Reactivate an inactive " + label + " so it shows in lists, searches and pickers again. Needs " + permission,
			ToolName:        "RestoreReferenceRecord",
			HTTPMethod:      "POST",
			RESTPath:        "by-guid/{recordGuid}/restore",
			RejectExtraArgs: true,
			Args: []ArgDef{
				{Name: "recordGuid", Description: "GUID of the inactive " + label, Required: true, Type: "non-empty-uuid"},
			},
		},
	)
}
