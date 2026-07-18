package registry

// Mirrors the MCP UteamupAttendanceStation* read tools backed by
// AttendanceStationController on the backend. Guid-first per the
// GUIDs-In/Integer-Ids-Out rule — the station identity is always the
// public station GUID. The runtime in registry.go calls
// apiClient.CallREST(...) against the declared APIPath, so action Name +
// RESTPath build the URL.
//
// REST surface (read-only slice exposed here):
//
//	GET /api/attendancestation                                  — list stations (?activeOnly=true)
//	GET /api/attendancestation/{stationGuid}                    — fetch one station
//	GET /api/attendancestation/attendance/corrections/pending   — corrections awaiting manager review
func init() {
	Register(&Domain{
		Name:        "attendance-station",
		Aliases:     []string{"attendance"},
		Description: "Read attendance stations and pending attendance corrections",
		APIPath:     "/api/attendancestation",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List attendance stations in the active tenant",
				ToolName:    "UteamupAttendanceStationList",
				Flags: []FlagDef{
					{Name: "active-only", Description: "Return only active stations (set --active-only=false to include inactive)", Default: true, Type: "bool"},
				},
			},
			{
				Name:        "get",
				Description: "Get one attendance station by its public GUID",
				ToolName:    "UteamupAttendanceStationGetByGuid",
				RESTPath:    "{stationGuid}",
				Args: []ArgDef{
					{Name: "stationGuid", Description: "Attendance station GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "corrections-pending",
				Description: "List pending attendance corrections awaiting manager review",
				ToolName:    "UteamupAttendanceCorrectionsPending",
				HTTPMethod:  "GET",
				RESTPath:    "attendance/corrections/pending",
			},
		},
	})
}
