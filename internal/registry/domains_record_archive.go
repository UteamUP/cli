package registry

// withRecordArchive adds `archive` and `restore` to a domain whose records are archived, never deleted
// (contact, checklist, task list, tag, floorplan, failure code, warranty, UPMate agent, report template, knowledge
// space). Its GET list gains --archived, and its `delete` stays only as an alias of archive: the backend's old
// delete route archives too. An archived record keeps every link and history row that names it and stays readable
// by GUID. Both routes need the permission that guarded deleting the record.
func withRecordArchive(actions []Action, label, plural, permission string) []Action {
	for index := range actions {
		switch actions[index].Name {
		case "list":
			actions[index].Description += "; archived " + plural + " are hidden unless you pass --archived"
			actions[index].Flags = append(actions[index].Flags, FlagDef{
				Name: "archived", QueryName: "archived", Type: "bool",
				Description: "true lists only archived " + plural + "; omit it for the ones in use",
			})
		case "delete":
			actions[index].Description = "Archive the " + label + " (a " + label + " is never deleted; same as archive). Needs " + permission
		}
	}
	return append(actions, recordArchiveActions(label, permission)...)
}

// recordArchiveActions are the `archive` and `restore` actions on the domain's by-guid/{guid}/archive and /restore
// routes. Domains whose list cannot take the --archived query flag (automation, report, journal) add them directly.
func recordArchiveActions(label, permission string) []Action {
	return []Action{
		{
			Name: "archive",
			Description: "Archive the " + label + ": it leaves lists, searches and pickers, nothing linked to it is " +
				"removed, and it stays readable by GUID. Needs " + permission,
			ToolName:        "ArchiveRecord",
			HTTPMethod:      "POST",
			RESTPath:        "by-guid/{recordGuid}/archive",
			RejectExtraArgs: true,
			Args: []ArgDef{
				{Name: "recordGuid", Description: "GUID of the " + label + " to archive", Required: true, Type: "non-empty-uuid"},
			},
		},
		{
			Name:            "restore",
			Description:     "Restore an archived " + label + " so it shows in lists, searches and pickers again. Needs " + permission,
			ToolName:        "RestoreRecord",
			HTTPMethod:      "POST",
			RESTPath:        "by-guid/{recordGuid}/restore",
			RejectExtraArgs: true,
			Args: []ArgDef{
				{Name: "recordGuid", Description: "GUID of the archived " + label, Required: true, Type: "non-empty-uuid"},
			},
		},
	}
}
