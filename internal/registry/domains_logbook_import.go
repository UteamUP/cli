package registry

// Logbook-import CLI surface — mirrors LogbookImport MCP tools.

func init() {
	Register(&Domain{
		Name:        "logbook-import",
		Aliases:     []string{"logimp"},
		Description: "Inspect parsed Word-logbook (.docx) imports before committing entries as Journals",
		APIPath:     "/api/logbookimport",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get a parsed logbook import with its entries",
				ToolName:    "UteamupLogbookImportGet",
				RESTPath:    "{externalGuid}",
				Args:        externalGUIDArg(),
			},
		},
	})
}
