package registry

func init() {
	Register(&Domain{
		Name:        "floorplan-scene",
		Aliases:     []string{"scene3d"},
		Description: "Read, save and AI-generate the 3D scene of a location's floorplans",
		APIPath:     "/api/floorplanscene",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Get the 3D scene saved for a location",
				ToolName:    "UteamupFloorplanSceneGet",
				HTTPMethod:  "GET",
				RESTPath:    "location/{locationGuid}",
				Args: []ArgDef{
					{Name: "locationGuid", Description: "Location GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "save",
				Description: "Create or replace a location's 3D scene; a stale expected version is refused with 409",
				ToolName:    "UteamupFloorplanSceneSave",
				HTTPMethod:  "PUT",
				RESTPath:    "location/{locationGuid}",
				Args: []ArgDef{
					{Name: "locationGuid", Description: "Location GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "name", Description: "Scene name", Required: true, Type: "string"},
					{Name: "expected-version", BodyName: "expectedVersion", Description: "Version the scene was loaded at; 0 for a new scene", Type: "int", Default: 0},
					{
						Name:        "scene-file",
						BodyName:    "scene",
						Description: "JSON file containing the scene document (schemaVersion 1)",
						Required:    true,
						Type:        "string",
						JSONFile:    true,
					},
					{Name: "generated-by-ai", BodyName: "generatedByAi", Description: "Mark the scene as an accepted AI draft", Type: "bool"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a location's 3D scene; the floorplans are kept",
				ToolName:    "UteamupFloorplanSceneDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "location/{locationGuid}",
				Args: []ArgDef{
					{Name: "locationGuid", Description: "Location GUID", Required: true, Type: "uuid"},
				},
			},
			{
				// REST-only: there is no MCP estimate tool, so this ToolName is metadata.
				Name:        "estimate",
				Description: "Show the AI credits generating a scene from the location's floorplans would cost",
				ToolName:    "UteamupFloorplanSceneEstimate",
				HTTPMethod:  "GET",
				RESTPath:    "location/{locationGuid}/generate-estimate",
				Args: []ArgDef{
					{Name: "locationGuid", Description: "Location GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "generate",
				Description: "Generate a draft scene from the location's floorplans with AI (spends AI credits; not saved)",
				ToolName:    "UteamupFloorplanSceneGenerate",
				HTTPMethod:  "POST",
				RESTPath:    "location/{locationGuid}/generate",
				Args: []ArgDef{
					{Name: "locationGuid", Description: "Location GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
