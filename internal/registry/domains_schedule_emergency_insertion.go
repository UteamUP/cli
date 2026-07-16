package registry

func init() {
	Register(&Domain{
		Name:        "schedule-emergency-insertion",
		Aliases:     []string{"emergency-insertion", "schedule-emergency"},
		Description: "Preview the least-disruptive emergency workorder insertion without changing the schedule",
		APIPath:     "/api/schedule/emergency-insertions",
		Actions: []Action{
			{
				Name:        "preview",
				Description: "Preview GUID-only displacement, travel, competency, and blocking evidence",
				ToolName:    "UteamupScheduleEmergencyInsertionPreview",
				HTTPMethod:  "POST",
				RESTPath:    "preview",
				Flags: []FlagDef{
					{Name: "workorder-guid", BodyName: "workorderGuid", Description: "Tenant-scoped urgent workorder GUID", Required: true, Type: "string"},
					{Name: "desired-start-utc", BodyName: "desiredStartUtc", Description: "Requested UTC start for the emergency work", Required: true, Type: "string"},
					{Name: "planning-window-end-utc", BodyName: "planningWindowEndUtc", Description: "Optional UTC end of the review window", Type: "string"},
					{Name: "team-guid", BodyName: "teamGuid", Description: "Optional tenant-scoped team GUID", Type: "string"},
					{Name: "technician-guids", BodyName: "technicianGuids", Description: "Optional tenant-member GUIDs to evaluate (maximum 30)", Type: "stringSlice"},
					{Name: "max-displaced-assignments", BodyName: "maxDisplacedAssignments", Description: "Maximum assignments the preview may move (0-10)", Type: "int", Default: 5},
				},
			},
		},
	})
}
