package registry

// NFC, QR and barcode tags on locations, route stops and checkpoints. Read only:
// lookups never record a scan. The backend MCP tools require Asset.View.
func init() {
	Register(&Domain{
		Name:        "location-nfc-tag",
		Aliases:     []string{"location-nfc-tags", "location-tags"},
		Description: "List and look up tags registered on locations and checkpoints",
		Actions: []Action{
			{Name: "list", Description: "List active location tags", ToolName: "UteamupLocationNfcTagList", MCPOnly: true, RejectExtraArgs: true},
			{Name: "lookup", Description: "Find the location of a scanned tag value", ToolName: "UteamupLocationNfcTagGetByTagId", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{{Name: "tag-id", BodyName: "tagId", Description: "Value stored on the physical tag", Required: true, Type: "string"}}},
		},
	})
}
