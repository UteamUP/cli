package registry

// CLI domain mirroring the MCP `CodePdfHotspotTools` — manages positional
// hotspots on PDF / drawing pages. All identifiers are GUIDs at the boundary;
// integer FKs never cross the CLI surface.
//
// Tool names match the MCP `[McpServerTool]` method names exactly so the
// runtime can route the action to the correct backend handler.
func init() {
	Register(&Domain{
		Name:        "pdfhotspot",
		Aliases:     []string{"hotspot", "pdfhotspots"},
		Description: "Manage positional hotspots that anchor industrial codes to drawing/PDF pages",
		Actions: []Action{
			{
				Name:        "list-for-drawing",
				Description: "List every hotspot on a drawing/PDF (codes referenced + their normalized rectangles)",
				ToolName:    "UteamupCodepdfhotspotListForDrawing",
				Args: []ArgDef{
					{Name: "documentGuid", Description: "Drawing/document Guid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "list-drawings-for-code",
				Description: "List every drawing where a given industrial code has a hotspot (or a CodePdfLink without an anchor yet)",
				ToolName:    "UteamupCodepdfhotspotListDrawingsForCode",
				Args: []ArgDef{
					{Name: "codeGuid", Description: "Code catalog entry Guid", Required: true, Type: "string"},
				},
			},
			{
				Name:        "create",
				Description: "Create a positional hotspot on a CodePdfLink (rectangle in normalized 0..1 coords)",
				ToolName:    "UteamupCodepdfhotspotCreate",
				Args: []ArgDef{
					{Name: "linkGuid", Description: "Parent CodePdfLink Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "from-json", Description: "JSON file with the CodePdfHotspotCreateModel payload (page, shape, x, y, w, h, kind, target Guids, optional label)", Type: "string"},
					{Name: "page", Description: "1-based page number on the parent drawing", Default: 1, Type: "int"},
					{Name: "x", Description: "Normalized [0,1] horizontal origin (left edge for rectangles)", Default: 0.0, Type: "float"},
					{Name: "y", Description: "Normalized [0,1] vertical origin (top edge for rectangles)", Default: 0.0, Type: "float"},
					{Name: "w", Description: "Normalized [0,1] width", Default: 0.0, Type: "float"},
					{Name: "h", Description: "Normalized [0,1] height", Default: 0.0, Type: "float"},
					{Name: "kind", Description: "Hotspot kind: 'code' (default) or 'xref'", Default: "code", Type: "string"},
					{Name: "target-code-guid", Description: "Code catalog entry Guid (required when kind=code)", Type: "string"},
					{Name: "target-document-guid", Description: "Document Guid the hotspot cross-references (required when kind=xref)", Type: "string"},
					{Name: "label", Description: "Optional label (max 200 chars)", Type: "string"},
				},
			},
			{
				Name:        "update",
				Description: "Update an existing hotspot — any of page, shape, coords, kind, target Guids, or label",
				ToolName:    "UteamupCodepdfhotspotUpdate",
				Args: []ArgDef{
					{Name: "linkGuid", Description: "Parent CodePdfLink Guid", Required: true, Type: "string"},
					{Name: "hotspotGuid", Description: "Hotspot Guid", Required: true, Type: "string"},
				},
				Flags: []FlagDef{
					{Name: "from-json", Description: "JSON file with the CodePdfHotspotUpdateModel payload", Type: "string"},
					{Name: "page", Description: "New 1-based page number", Type: "int"},
					{Name: "x", Description: "New normalized [0,1] x", Type: "float"},
					{Name: "y", Description: "New normalized [0,1] y", Type: "float"},
					{Name: "w", Description: "New normalized [0,1] width", Type: "float"},
					{Name: "h", Description: "New normalized [0,1] height", Type: "float"},
					{Name: "kind", Description: "New kind ('code' or 'xref')", Type: "string"},
					{Name: "target-code-guid", Description: "New target code Guid", Type: "string"},
					{Name: "target-document-guid", Description: "New target document Guid", Type: "string"},
					{Name: "label", Description: "New label (max 200 chars)", Type: "string"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a hotspot from its parent CodePdfLink",
				ToolName:    "UteamupCodepdfhotspotDelete",
				Args: []ArgDef{
					{Name: "linkGuid", Description: "Parent CodePdfLink Guid", Required: true, Type: "string"},
					{Name: "hotspotGuid", Description: "Hotspot Guid", Required: true, Type: "string"},
				},
			},
		},
	})
}
