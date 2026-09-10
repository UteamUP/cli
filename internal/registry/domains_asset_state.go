package registry

func init() {
	// Routes mirror AssetStateController (/api/assetstate/*). An asset carries two linked
	// axes: where it is in its life (the lifecycle stage) and whether it is running right
	// now (the operational state). Reading either is Asset.View; moving the operational
	// axis needs the dedicated Asset.SetOperationalState right, so a technician can take a
	// pump offline without holding approver rights. Every identifier is a GUID.
	assetGuid := ArgDef{Name: "assetGuid", Description: "Asset GUID", Required: true, Type: "non-empty-uuid"}

	Register(&Domain{
		Name:        "asset-state",
		Aliases:     []string{"astate", "asset-states"},
		Description: "Read and change whether an asset is running, and see what that costs it",
		APIPath:     "/api/assetstate",
		Actions: []Action{
			{
				Name:        "get",
				Description: "Both axes, the moves the asset may legally make next, and its lifespan position",
				ToolName:    "UteamupAssetStateGet",
				HTTPMethod:  "GET",
				RESTPath:    "asset/by-guid/{assetGuid}",
				Args:        []ArgDef{assetGuid},
			},
			{
				Name:        "history",
				Description: "Operational state history as closed intervals, newest first",
				ToolName:    "UteamupAssetStateHistory",
				HTTPMethod:  "GET",
				RESTPath:    "asset/by-guid/{assetGuid}/history",
				Args:        []ArgDef{assetGuid},
			},
			{
				Name:        "availability",
				Description: "Uptime and downtime over a window; Standby and Degraded do not count as downtime",
				ToolName:    "UteamupAssetStateAvailability",
				HTTPMethod:  "GET",
				RESTPath:    "asset/by-guid/{assetGuid}/availability",
				Args:        []ArgDef{assetGuid},
				Flags: []FlagDef{
					{Name: "from", QueryName: "from", Description: "Start of the window (UTC, ISO 8601)", Required: true, Type: "string"},
					{Name: "to", QueryName: "to", Description: "End of the window (UTC, ISO 8601)", Required: true, Type: "string"},
				},
			},
			{
				Name:        "set-operational",
				Description: "Move the operational axis; the lifecycle stage constrains what is reachable",
				ToolName:    "UteamupAssetStateSetOperational",
				HTTPMethod:  "PUT",
				RESTPath:    "asset/by-guid/{assetGuid}/operational",
				Args:        []ArgDef{assetGuid},
				Flags: []FlagDef{
					{Name: "to-state", BodyName: "toState", Description: "Online, Offline, Standby, UnderMaintenance, Degraded, Lost, Stolen, InTransit or InStorage. Lost and Stolen stop the asset accepting new work orders; Degraded opens a replacement plan", Required: true, Type: "string"},
					{Name: "reason", Description: "Why the asset is being put into this state", Type: "string"},
					{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Work order that prompted the change", Type: "non-empty-uuid"},
				},
			},
		},
	})
}
