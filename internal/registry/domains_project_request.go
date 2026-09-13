package registry

// projectRequestFile keeps typed project REST DTOs at the JSON root. The
// registry does not switch transport based on login/apikey token provenance.
func projectRequestFile() FlagDef {
	return FlagDef{Name: "from-json", Description: "File containing the complete root JSON request; preserve reviewed identities, nulls and fingerprints",
		Type: "string", Required: true, RootJSONObjectFile: true}
}

// Project CRUD REST accepts multipart forms. Its named MCP model parameter
// therefore uses a parsed file under model instead of an unwrapped REST body.
func projectMCPModelFile() FlagDef {
	return FlagDef{Name: "from-json", Description: "File containing the complete project model (without a model wrapper)",
		Type: "string", Required: true, JSONFile: true, BodyName: "model"}
}
