package registry

// Public reporting links — the QR posters that let anyone submit an incident or
// hazard report without an account.
//
// Deliberately absent: an action to SUBMIT a report. That endpoint is
// unauthenticated, reCAPTCHA-gated and rate-limited by design. Exposing a
// scripted path through an authenticated CLI would be an abuse vector, and it
// could not satisfy the captcha anyway. Do not "fix" this omission.
func init() {
	Register(&Domain{
		Name:        "report-link",
		Aliases:     []string{"prl", "reporting-link"},
		Description: "Manage public anonymous reporting links (QR posters)",
		APIPath:     "/api/tenant/reportlinks",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List the tenant's public reporting links",
				ToolName:    "UteamupReportLinkList",
				RESTPath:    "",
				HTTPMethod:  "GET",
			},
			{
				Name:        "get",
				Description: "Get a single reporting link by its public GUID",
				ToolName:    "UteamupReportLinkGet",
				RESTPath:    "{guid}",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Public GUID of the reporting link", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "create",
				Description: "Create a reporting link and return its public token",
				ToolName:    "UteamupReportLinkCreate",
				RESTPath:    "",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "name", Short: "n", Description: "Where the poster will be deployed, e.g. \"Warehouse B - loading dock\"", Required: true, Type: "string"},
					{Name: "description", Short: "d", Description: "Internal note about the link's purpose or placement", Type: "string"},
					{Name: "require-contact-details", BodyName: "requireContactDetails", Description: "Require the reporter to identify themselves (defaults false — anonymity is the point)", Type: "bool"},
					{Name: "allow-photo-upload", BodyName: "allowPhotoUpload", Description: "Offer a photo upload on the form", Type: "bool", Default: true},
					{Name: "allow-category-selection", BodyName: "allowCategorySelection", Description: "Let the reporter choose a category", Type: "bool", Default: true},
					{Name: "allow-location-selection", BodyName: "allowLocationSelection", Description: "Let the reporter choose a location", Type: "bool", Default: true},
					{Name: "default-category-guid", BodyName: "defaultCategoryGuid", Description: "Category applied when the reporter does not choose one", Type: "uuid"},
					{Name: "default-location-guid", BodyName: "defaultLocationGuid", Description: "Location applied when the reporter does not choose one", Type: "uuid"},
					{Name: "default-asset-guid", BodyName: "defaultAssetGuid", Description: "Asset this poster is attached to, if machine-specific", Type: "uuid"},
					{Name: "notify-team-guid", BodyName: "notifyTeamGuid", Description: "Team notified when a report arrives", Type: "uuid"},
					{Name: "max-submissions-per-hour", BodyName: "maxSubmissionsPerHour", Description: "Per-link hourly abuse ceiling", Type: "int", Default: 20},
				},
			},
			{
				Name:        "update",
				Description: "Update a reporting link's configuration",
				ToolName:    "UteamupReportLinkUpdate",
				RESTPath:    "{guid}",
				HTTPMethod:  "PUT",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Public GUID of the reporting link", Required: true, Type: "uuid"},
					{Name: "name", Short: "n", Description: "Where the poster is deployed", Required: true, Type: "string"},
					{Name: "description", Short: "d", Description: "Internal note", Type: "string"},
					{Name: "is-active", BodyName: "isActive", Description: "Whether the link accepts reports", Type: "bool", Default: true},
					{Name: "require-contact-details", BodyName: "requireContactDetails", Description: "Require the reporter to identify themselves", Type: "bool"},
					{Name: "allow-photo-upload", BodyName: "allowPhotoUpload", Description: "Offer a photo upload on the form", Type: "bool", Default: true},
					{Name: "allow-category-selection", BodyName: "allowCategorySelection", Description: "Let the reporter choose a category", Type: "bool", Default: true},
					{Name: "allow-location-selection", BodyName: "allowLocationSelection", Description: "Let the reporter choose a location", Type: "bool", Default: true},
					{Name: "default-category-guid", BodyName: "defaultCategoryGuid", Description: "Category applied when the reporter does not choose one", Type: "uuid"},
					{Name: "default-location-guid", BodyName: "defaultLocationGuid", Description: "Location applied when the reporter does not choose one", Type: "uuid"},
					{Name: "default-asset-guid", BodyName: "defaultAssetGuid", Description: "Asset this poster is attached to", Type: "uuid"},
					{Name: "notify-team-guid", BodyName: "notifyTeamGuid", Description: "Team notified when a report arrives", Type: "uuid"},
					{Name: "max-submissions-per-hour", BodyName: "maxSubmissionsPerHour", Description: "Per-link hourly abuse ceiling", Type: "int", Default: 20},
				},
			},
			{
				Name:        "regenerate-token",
				Description: "Issue a new token, invalidating the printed URL immediately",
				ToolName:    "UteamupReportLinkRegenerateToken",
				RESTPath:    "{guid}/regenerate-token",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Public GUID of the reporting link", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete a reporting link (prefer deactivating if the poster is still on a wall)",
				ToolName:    "UteamupReportLinkDelete",
				RESTPath:    "{guid}",
				HTTPMethod:  "DELETE",
				Flags: []FlagDef{
					{Name: "guid", Short: "g", Description: "Public GUID of the reporting link", Required: true, Type: "uuid"},
				},
			},
		},
	})
}
