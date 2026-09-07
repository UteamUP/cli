package registry

func init() {
	Register(&Domain{
		Name: "fleet-transport", Description: "Plan and dispatch internal passenger transport with workforce allocations", APIPath: "/api/fleet/transport",
		Actions: []Action{
			{Name: "list", Description: "List transport services in a bounded UTC range", ToolName: "UteamupFleetTransportList", HTTPMethod: "GET", Flags: []FlagDef{
				{Name: "from-utc", BodyName: "fromUtc", Description: "Inclusive UTC start", Required: true, Type: "string"},
				{Name: "to-utc", BodyName: "toUtc", Description: "Exclusive UTC end (at most 93 days)", Required: true, Type: "string"},
				{Name: "page", Description: "Page number", Type: "int", Default: 1},
				{Name: "page-size", BodyName: "pageSize", Description: "Results per page, 1-100", Type: "int", Default: 30},
			}},
			{Name: "get", Description: "Read a service and manifest", ToolName: "UteamupFleetTransportGet", HTTPMethod: "GET", RESTPath: "{assignmentGuid}", Args: []ArgDef{{Name: "assignmentGuid", Description: "Public service GUID", Required: true, Type: "uuid"}}},
			{Name: "create", Description: "Create a reviewed transport plan with an idempotency key", ToolName: "UteamupFleetTransportCreate", HTTPMethod: "POST", Flags: []FlagDef{{Name: "request-file", Short: "f", BodyName: "model", Description: "JSON ShiftInstanceVehicleCreateModel using public GUIDs and UTC dates", Required: true, Type: "string", JSONFile: true}}},
			{Name: "transition", Description: "Apply a reviewed lifecycle transition and optional final fare", ToolName: "UteamupFleetTransportTransition", HTTPMethod: "PUT", RESTPath: "{assignmentGuid}/status", Args: []ArgDef{{Name: "assignmentGuid", Description: "Public service GUID", Required: true, Type: "uuid"}}, Flags: []FlagDef{{Name: "request-file", Short: "f", BodyName: "model", Description: "JSON status, expectedUpdatedAtUtc and reviewed fare fields", Required: true, Type: "string", JSONFile: true}}},
			{Name: "manifest", Description: "Replace the reviewed passenger manifest within verified vehicle capacity", ToolName: "UteamupFleetTransportManifest", HTTPMethod: "PUT", RESTPath: "{assignmentGuid}/manifest", Args: []ArgDef{{Name: "assignmentGuid", Description: "Public service GUID", Required: true, Type: "uuid"}}, Flags: []FlagDef{{Name: "request-file", Short: "f", BodyName: "model", Description: "JSON passengers and expectedUpdatedAtUtc", Required: true, Type: "string", JSONFile: true}}},
		},
	})
}
