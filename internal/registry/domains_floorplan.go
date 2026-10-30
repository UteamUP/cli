package registry

func init() {
	Register(&Domain{
		Name:        "floorplan",
		Description: "Read and download location floorplan drawings using public GUIDs",
		APIPath:     "/api/floorplan",
		Actions: []Action{
			{Name: "list", Description: "List floorplan drawings and their public GUIDs", ToolName: "UteamupFloorplanList", HTTPMethod: "GET"},
			{Name: "get", Description: "Get a drawing by its public GUID", ToolName: "UteamupFloorplanGet", HTTPMethod: "GET", RESTPath: "{guid}",
				Args: []ArgDef{{Name: "guid", Description: "Floorplan GUID returned by list", Required: true, Type: "uuid"}}},
			{Name: "by-location", Description: "List drawings for a location GUID", ToolName: "UteamupFloorplanGetByLocation", HTTPMethod: "GET", RESTPath: "location/{locationGuid}",
				Args: []ArgDef{{Name: "locationGuid", Description: "Location GUID", Required: true, Type: "uuid"}}},
			{Name: "by-location-floor", Description: "Get a drawing for a location-floor GUID through MCP (API-key login required)", ToolName: "UteamupFloorplanGetByLocationFloor", MCPOnly: true,
				Args: []ArgDef{{Name: "locationFloorGuid", Description: "Location-floor GUID", Required: true, Type: "uuid"}}},
			{Name: "download", Description: "Download a drawing as base64 through MCP (API-key login required)", ToolName: "UteamupFloorplanDownload", MCPOnly: true,
				Args: []ArgDef{{Name: "guid", Description: "Floorplan GUID returned by list", Required: true, Type: "uuid"}}},
		},
	})
}
