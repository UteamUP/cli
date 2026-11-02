package registry

func partCrudActions() []Action {
	actions := crudActions("Part")
	for index := range actions {
		if actions[index].Name == "create" || actions[index].Name == "update" {
			actions[index].Flags = append(actions[index].Flags, FlagDef{Name: "currency-guid", BodyName: "valueCurrencyGuid", Description: "Catalog value currency public GUID; empty GUID clears explicit metadata", Type: "uuid"})
		}
	}
	return actions
}

// Catalog file actions reuse the GUID REST routes; base64 image upload goes through
// the matching bounded MCP operation instead of inventing a second multipart encoder.
func catalogFileActions(kind, name string, documentByGuid bool) []Action {
	item := ArgDef{Name: name + "Guid", Description: kind + " public GUID", Required: true, Type: "non-empty-uuid"}
	document := ArgDef{Name: "documentGuid", Description: "Document public GUID", Required: true, Type: "non-empty-uuid"}
	documentPath := "by-guid/{" + item.Name + "}/documents/"
	if documentByGuid {
		documentPath += "by-guid/"
	}
	documentPath += "{documentGuid}"
	return []Action{
		{Name: "set-image-from-document", Description: "Make an already-linked image document the profile image", ToolName: "Uteamup" + kind + "SetImageFromDocument", HTTPMethod: "POST", RESTPath: "by-guid/{" + item.Name + "}/image/from-document/{documentGuid}", Args: []ArgDef{item, document}, RejectExtraArgs: true},
		{Name: "upload-image", Description: "Upload a PNG/JPEG profile image, up to 2 MB decoded; uses MCP", ToolName: "Uteamup" + kind + "UploadImage", MCPOnly: true, Args: []ArgDef{item}, RejectExtraArgs: true,
			Flags: []FlagDef{
				{Name: "file-name", BodyName: "fileName", Description: "Image file name without directory names", Required: true, Type: "string"},
				{Name: "content-type", BodyName: "contentType", Description: "image/png or image/jpeg", Required: true, Type: "string", AllowedValues: []string{"image/png", "image/jpeg"}},
				{Name: "base64-content", BodyName: "base64Content", Description: "Base64 bytes, without data-URL prefix, up to 2 MB decoded", Required: true, Type: "string", Sensitive: true},
			}},
		{Name: "add-document", Description: "Attach an existing tenant document; safe to repeat", ToolName: "Uteamup" + kind + "AddDocument", HTTPMethod: "POST", RESTPath: documentPath, Args: []ArgDef{item, document}, RejectExtraArgs: true},
		{Name: "remove-document", Description: "Detach a document, preserving the stored file", ToolName: "Uteamup" + kind + "RemoveDocument", HTTPMethod: "DELETE", RESTPath: documentPath, Args: []ArgDef{item, document}, RejectExtraArgs: true},
	}
}
