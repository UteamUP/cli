package registry

// Connector trust changes are deliberately absent from CLI and MCP.
func init() {
	Register(&Domain{
		Name: "edge-connectors", Description: "Read registered connectors and Radio module state", APIPath: "/api/edge-connectors",
		Actions: []Action{
			{Name: "list", Description: "List connector registrations in the active tenant", ToolName: "UteamupEdgeConnectorsList", HTTPMethod: "GET"},
			{Name: "radio-status", Description: "Read purchased Radio capacity and pending appliance acknowledgements", ToolName: "UteamupEdgeConnectorRadioGet", HTTPMethod: "GET", RESTPath: "{connectorGuid}/radio",
				Args: []ArgDef{{Name: "connectorGuid", Description: "Public connector GUID", Type: "non-empty-uuid", Required: true}}},
		},
	})
}
