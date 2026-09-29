package registry

// Field checks use the same server-side permissions and actor scope as the
// mobile and web workflows. Complex form payloads are supplied as JSON files.
func init() {
	Register(&Domain{
		Name:        "field-check",
		Aliases:     []string{"checks", "readings-checks"},
		Description: "Configure checkpoints and record or review field visits",
		Actions: []Action{
			{Name: "settings", Description: "Get tenant evidence settings", ToolName: "UteamupFieldcheckSettingsGet", MCPOnly: true},
			{Name: "settings-save", Description: "Save tenant evidence settings", ToolName: "UteamupFieldcheckSettingsSave", MCPOnly: true,
				Flags: []FlagDef{{Name: "model-file", BodyName: "model", Description: "JSON file with isEnabled, requirePhoto, requireLocation", Required: true, Type: "string", JSONFile: true}}},
			{Name: "points", Description: "List active checkpoints", ToolName: "UteamupFieldcheckPointsList", MCPOnly: true},
			{Name: "point-save", Description: "Create or update a checkpoint", ToolName: "UteamupFieldcheckPointSave", MCPOnly: true,
				Flags: []FlagDef{
					{Name: "model-file", BodyName: "model", Description: "Checkpoint JSON file", Required: true, Type: "string", JSONFile: true},
					{Name: "checkpoint-guid", BodyName: "checkpointGuid", Description: "Existing checkpoint public GUID", Type: "uuid"},
				}},
			{Name: "due", Description: "List due route-stop checkpoints", ToolName: "UteamupFieldcheckDue", MCPOnly: true},
			{Name: "visits", Description: "List visits visible to the actor", ToolName: "UteamupFieldcheckVisitsList", MCPOnly: true,
				Flags: []FlagDef{{Name: "status", Description: "Optional visit status", Type: "string"}}},
			{Name: "start", Description: "Start an incomplete visit with an idempotent client GUID", ToolName: "UteamupFieldcheckVisitStart", MCPOnly: true,
				Flags: []FlagDef{{Name: "model-file", BodyName: "model", Description: "Visit start JSON file", Required: true, Type: "string", JSONFile: true}}},
			{Name: "submit", Description: "Submit answers and evidence for a visit", ToolName: "UteamupFieldcheckVisitSubmit", MCPOnly: true,
				Args:  []ArgDef{{Name: "visit-guid", BodyName: "visitGuid", Description: "Visit public GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{{Name: "model-file", BodyName: "model", Description: "Visit answers JSON file", Required: true, Type: "string", JSONFile: true}}},
			{Name: "review", Description: "Approve or reject a visit needing review", ToolName: "UteamupFieldcheckVisitReview", MCPOnly: true,
				Args:  []ArgDef{{Name: "visit-guid", BodyName: "visitGuid", Description: "Visit public GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{{Name: "model-file", BodyName: "model", Description: "Review decision JSON file", Required: true, Type: "string", JSONFile: true}}},
		},
	})
}
