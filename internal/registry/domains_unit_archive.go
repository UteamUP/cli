package registry

// unitArchiveActions adds archive and restore to the customer, vendor, team and location domains:
// POST {domain}/by-guid/{guid}/archive and .../restore. Each needs that type's delete permission.
func unitArchiveActions(guidArg, toolPrefix, archiveDescription, restoreDescription string) []Action {
	identity := []ArgDef{{Name: guidArg, Description: "Public GUID", Required: true, Type: "non-empty-uuid"}}
	return []Action{
		{
			Name:            "archive",
			Description:     archiveDescription,
			ToolName:        toolPrefix + "Archive",
			HTTPMethod:      "POST",
			RESTPath:        "by-guid/{" + guidArg + "}/archive",
			RejectExtraArgs: true,
			Args:            identity,
		},
		{
			Name:            "restore",
			Description:     restoreDescription,
			ToolName:        toolPrefix + "Restore",
			HTTPMethod:      "POST",
			RESTPath:        "by-guid/{" + guidArg + "}/restore",
			RejectExtraArgs: true,
			Args:            identity,
		},
	}
}
