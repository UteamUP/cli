package registry

// Fleet inspection templates are read through the backend MCP tools, which apply
// InspectionTemplate.Read and the Fleet Inspections feature like the REST routes.
func init() {
	Register(&Domain{
		Name:        "inspection-template",
		Aliases:     []string{"inspection-templates"},
		Description: "List and read fleet inspection checklist templates",
		Actions: []Action{
			{Name: "list", Description: "List inspection templates, including system templates", ToolName: "UteamupInspectionTemplateList", MCPOnly: true, RejectExtraArgs: true},
			{Name: "get", Description: "Get an inspection template and its items", ToolName: "UteamupInspectionTemplateGet", MCPOnly: true, RejectExtraArgs: true,
				Args: []ArgDef{{Name: "template-guid", BodyName: "templateGuid", Description: "Inspection template public GUID", Required: true, Type: "non-empty-uuid"}}},
		},
	})
}
