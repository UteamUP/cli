package registry

func init() {
	Register(&Domain{
		Name:        "geofence-zone",
		Aliases:     []string{"geofence-zones", "geofence", "zone"},
		Description: "Manage geofence zones for GPS-based time tracking",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List all geofence zones with pagination and filtering",
				ToolName:    "UteamupGeofenceZoneList",
				Flags: []FlagDef{
					{Name: "page", Short: "p", Description: "Page number", Default: 1, Type: "int"},
					{Name: "page-size", Short: "s", Description: "Items per page", Default: 25, Type: "int"},
					{Name: "filter", Short: "f", Description: "Filter by name", Type: "string"},
					{Name: "sort-by", Description: "Sort field (Name, CreatedAt, etc.)", Default: "Name", Type: "string"},
					{Name: "sort-order", Description: "Sort direction (asc or desc)", Default: "asc", Type: "string"},
				},
			},
			{
				Name:        "get",
				Description: "Get geofence zone details by ID",
				ToolName:    "UteamupGeofenceZoneGet",
				Args:        []ArgDef{{Name: "id", Description: "Geofence zone ID", Required: true, Type: "int"}},
			},
			{
				Name:        "create",
				Description: "Create a new geofence zone",
				ToolName:    "UteamupGeofenceZoneCreate",
				Flags: []FlagDef{
					{Name: "name", Description: "Zone name", Required: true, Type: "string"},
					{Name: "center-latitude", Description: "Center latitude coordinate", Required: true, Type: "float"},
					{Name: "center-longitude", Description: "Center longitude coordinate", Required: true, Type: "float"},
					{Name: "radius-meters", Description: "Zone radius in meters", Default: 200, Type: "float"},
					{Name: "from-json", Description: "JSON file with zone data", Type: "string"},
				},
			},
			{
				Name:        "update",
				Description: "Update an existing geofence zone",
				ToolName:    "UteamupGeofenceZoneUpdate",
				Args:        []ArgDef{{Name: "id", Description: "Geofence zone ID", Required: true, Type: "int"}},
				Flags: []FlagDef{
					{Name: "name", Description: "New zone name", Type: "string"},
					{Name: "center-latitude", Description: "New center latitude", Type: "float"},
					{Name: "center-longitude", Description: "New center longitude", Type: "float"},
					{Name: "radius-meters", Description: "New radius in meters", Type: "float"},
					{Name: "from-json", Description: "JSON file with update data", Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a geofence zone by ID",
				ToolName:    "UteamupGeofenceZoneDelete",
				Args:        []ArgDef{{Name: "id", Description: "Geofence zone ID", Required: true, Type: "int"}},
			},
		},
	})
}
