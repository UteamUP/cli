package registry

// Anchors are served by DigitalTwinAnchorController on its own route, so every anchor
// action overrides the domain base path rather than nesting under /api/digitaltwin.
const digitalTwinAnchorAPIPath = "/api/digitaltwinanchor"

func init() {
	Register(&Domain{
		Name:        "digitaltwin",
		Aliases:     []string{"twin"},
		Description: "Bind assets to 3D models and read a twin's live anchor state",
		APIPath:     "/api/digitaltwin",
		Actions: []Action{
			{
				Name:        "get-by-asset",
				Description: "Get the digital twin bound to an asset, with its anchors",
				ToolName:    "UteamupDigitalTwinGet",
				HTTPMethod:  "GET",
				RESTPath:    "asset/{assetGuid}",
				Args: []ArgDef{
					{Name: "assetGuid", Description: "Public asset GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "get",
				Description: "Get a digital twin by GUID, with its anchors",
				ToolName:    "UteamupDigitalTwinGetByGuid",
				HTTPMethod:  "GET",
				RESTPath:    "{twinGuid}",
				Args: []ArgDef{
					{Name: "twinGuid", Description: "Digital twin GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Bind an uploaded 3D model document to an asset",
				ToolName:    "UteamupDigitalTwinCreate",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "asset-guid", BodyName: "assetGuid", Description: "Asset the twin represents", Required: true, Type: "uuid"},
					{Name: "model-document-guid", BodyName: "modelDocumentGuid", Description: "Uploaded 3D model document GUID", Required: true, Type: "uuid"},
					{Name: "name", Description: "Twin name", Required: true, Type: "string"},
					{Name: "description", Description: "Optional twin description", Type: "string"},
					{Name: "dimension-unit", BodyName: "dimensionUnit", Description: "Model dimension unit, e.g. m or mm", Type: "string"},
					{Name: "scale-ratio", BodyName: "scaleRatio", Description: "Model-to-real-world scale ratio", Type: "float"},
					{Name: "up-axis", BodyName: "upAxis", Description: "Model up axis, Y or Z", Type: "string"},
					{Name: "transform-json", BodyName: "transformJson", Description: "Scene transform as JSON", Type: "string"},
					{Name: "triangle-count", BodyName: "triangleCount", Description: "Triangle count reported by the loader", Type: "int"},
				},
			},
			{
				Name:        "update",
				Description: "Update a twin's presentation metadata; the asset and model document are immutable",
				ToolName:    "UteamupDigitalTwinUpdate",
				HTTPMethod:  "PUT",
				RESTPath:    "{twinGuid}",
				Args: []ArgDef{
					{Name: "twinGuid", Description: "Digital twin GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{Name: "name", Description: "Twin name", Type: "string"},
					{Name: "description", Description: "Twin description", Type: "string"},
					{Name: "dimension-unit", BodyName: "dimensionUnit", Description: "Model dimension unit, e.g. m or mm", Type: "string"},
					{Name: "scale-ratio", BodyName: "scaleRatio", Description: "Model-to-real-world scale ratio", Type: "float"},
					{Name: "up-axis", BodyName: "upAxis", Description: "Model up axis, Y or Z", Type: "string"},
					{Name: "camera-bookmarks-json", BodyName: "cameraBookmarksJson", Description: "Saved camera bookmarks as JSON", Type: "string"},
					{Name: "transform-json", BodyName: "transformJson", Description: "Scene transform as JSON", Type: "string"},
					{Name: "triangle-count", BodyName: "triangleCount", Description: "Triangle count reported by the loader", Type: "int"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a twin and its anchors; the asset and model document are kept",
				ToolName:    "UteamupDigitalTwinDelete",
				HTTPMethod:  "DELETE",
				RESTPath:    "{twinGuid}",
				Args: []ArgDef{
					{Name: "twinGuid", Description: "Digital twin GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "live",
				Description: "Read latest value, status and open-workorder count for every anchor",
				ToolName:    "UteamupDigitalTwinLiveValues",
				HTTPMethod:  "GET",
				RESTPath:    "{twinGuid}/live",
				Args: []ArgDef{
					{Name: "twinGuid", Description: "Digital twin GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "suggest-anchors",
				Description: "Match model node names against coded assets and return proposed anchors",
				ToolName:    "UteamupDigitalTwinSuggestAnchors",
				HTTPMethod:  "POST",
				RESTPath:    "{twinGuid}/suggest-anchors",
				Args: []ArgDef{
					{Name: "twinGuid", Description: "Digital twin GUID", Required: true, Type: "uuid"},
				},
				Flags: []FlagDef{
					{
						Name:        "nodes-file",
						BodyName:    "nodes",
						Description: "JSON file containing the model nodes: [{nodeName, x, y, z}]",
						Required:    true,
						Type:        "string",
						JSONFile:    true,
					},
				},
			},
			{
				Name:         "anchors",
				Description:  "List the anchors placed on a twin",
				ToolName:     "UteamupDigitalTwinAnchorsList",
				RESTBasePath: digitalTwinAnchorAPIPath,
				HTTPMethod:   "GET",
				RESTPath:     "twin/{twinGuid}",
				Args: []ArgDef{
					{Name: "twinGuid", Description: "Digital twin GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:         "anchor-create",
				Description:  "Place an anchor on a twin",
				ToolName:     "UteamupDigitalTwinAnchorCreate",
				RESTBasePath: digitalTwinAnchorAPIPath,
				HTTPMethod:   "POST",
				Flags: append([]FlagDef{
					{Name: "twin-guid", BodyName: "digitalTwinGuid", Description: "Digital twin the anchor belongs to", Required: true, Type: "uuid"},
					{Name: "x", Description: "Anchor X position in model space", Type: "float", Default: 0.0},
					{Name: "y", Description: "Anchor Y position in model space", Type: "float", Default: 0.0},
					{Name: "z", Description: "Anchor Z position in model space", Type: "float", Default: 0.0},
					{Name: "marker-shape", BodyName: "markerShape", Description: "Marker shape", Type: "string", Default: "Sphere"},
					{Name: "is-visible", BodyName: "isVisible", Description: "Whether the anchor renders on the model", Type: "bool", Default: true},
					{Name: "display-order", BodyName: "displayOrder", Description: "Render order among anchors", Type: "int", Default: 0},
				}, digitalTwinAnchorTargetFlags()...),
			},
			{
				Name:         "anchor-bulk-create",
				Description:  "Create several anchors at once when applying reviewed suggestions",
				ToolName:     "UteamupDigitalTwinAnchorBulkCreate",
				RESTBasePath: digitalTwinAnchorAPIPath,
				HTTPMethod:   "POST",
				RESTPath:     "bulk",
				Flags: []FlagDef{
					{Name: "twin-guid", BodyName: "digitalTwinGuid", Description: "Digital twin the anchors belong to", Required: true, Type: "uuid"},
					{
						Name:        "anchors-file",
						BodyName:    "anchors",
						Description: "JSON file containing an array of anchor create models",
						Required:    true,
						Type:        "string",
						JSONFile:    true,
					},
				},
			},
			{
				Name:         "anchor-update",
				Description:  "Move, retarget or re-threshold an anchor",
				ToolName:     "UteamupDigitalTwinAnchorUpdate",
				RESTBasePath: digitalTwinAnchorAPIPath,
				HTTPMethod:   "PUT",
				RESTPath:     "{anchorGuid}",
				Args: []ArgDef{
					{Name: "anchorGuid", Description: "Anchor GUID", Required: true, Type: "uuid"},
				},
				// No defaults: an unchanged flag must never be sent, or a partial update
				// would silently overwrite position, visibility or marker shape.
				Flags: append([]FlagDef{
					{Name: "x", Description: "Anchor X position in model space", Type: "float"},
					{Name: "y", Description: "Anchor Y position in model space", Type: "float"},
					{Name: "z", Description: "Anchor Z position in model space", Type: "float"},
					{Name: "marker-shape", BodyName: "markerShape", Description: "Marker shape", Type: "string"},
					{Name: "is-visible", BodyName: "isVisible", Description: "Whether the anchor renders on the model", Type: "bool"},
					{Name: "display-order", BodyName: "displayOrder", Description: "Render order among anchors", Type: "int"},
				}, digitalTwinAnchorTargetFlags()...),
			},
			{
				Name:         "anchor-delete",
				Description:  "Remove an anchor; the target asset and document are left untouched",
				ToolName:     "UteamupDigitalTwinAnchorDelete",
				RESTBasePath: digitalTwinAnchorAPIPath,
				HTTPMethod:   "DELETE",
				RESTPath:     "{anchorGuid}",
				Args: []ArgDef{
					{Name: "anchorGuid", Description: "Anchor GUID", Required: true, Type: "uuid"},
				},
			},
		},
	})
}

// digitalTwinAnchorTargetFlags returns the anchor fields that are optional and identically
// shaped on create and update. Placement and visibility differ between the two (create
// defaults them, update must not) and stay declared on each action.
func digitalTwinAnchorTargetFlags() []FlagDef {
	return []FlagDef{
		{Name: "target-asset-guid", BodyName: "targetAssetGuid", Description: "Asset the anchor points at", Type: "uuid"},
		{Name: "target-document-guid", BodyName: "targetDocumentGuid", Description: "Document the anchor points at", Type: "uuid"},
		{Name: "node-name", BodyName: "nodeName", Description: "Model node name the anchor is bound to", Type: "string"},
		{Name: "label", Description: "Anchor label shown on the model", Type: "string"},
		{Name: "source-meter-guid", BodyName: "sourceMeterGuid", Description: "Meter supplying the anchor's live value", Type: "uuid"},
		{Name: "source-attribute-definition-guid", BodyName: "sourceAttributeDefinitionGuid", Description: "Telemetry attribute definition supplying the live value", Type: "uuid"},
		{Name: "unit", Description: "Unit of the anchor's value", Type: "string"},
		{Name: "warning-threshold", BodyName: "warningThreshold", Description: "Value at which the anchor turns warning", Type: "float"},
		{Name: "critical-threshold", BodyName: "criticalThreshold", Description: "Value at which the anchor turns critical", Type: "float"},
	}
}
