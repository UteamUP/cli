package registry

func qualificationCatalogueActions(entity string) []Action {
	actions := crudActions(entity)
	for index := range actions {
		action := &actions[index]
		if action.Name == "get" || action.Name == "update" {
			action.Args = []ArgDef{{Name: "externalGuid", Description: "Catalog record public GUID", Required: true, Type: "non-empty-uuid"}}
			action.RESTPath = "by-guid/{externalGuid}"
		}
		if action.Name == "create" || action.Name == "update" {
			action.Flags = append(action.Flags, FlagDef{Name: "tag-guids", BodyName: "tagGuids", Type: "stringSlice",
				Description: "Complete set of active tenant tag GUIDs; omission preserves, empty clears"})
		}
	}
	return actions
}
