package registry

import "strings"

// archiveKinds are the record kinds shown in the Tenant Management archive (backend ArchiveOverviewSources).
var archiveKinds = []string{
	"contact", "checklist", "tasklist", "tag", "floorplan", "failurecode", "warranty", "workforcegroup", "holiday",
	"asset", "part", "tool", "chemical", "customer", "team", "location", "vendor", "contract", "project", "document",
	"category", "assettype",
}

// The company archive, read only: every archived record the caller may read, who archived it and when, and what
// links to it. Nothing archived is deleted; each record kind restores with its own `restore` action.
func init() {
	Register(&Domain{
		Name:        "archive",
		Description: "Read the company archive (archived records of every kind, who archived them, their links)",
		APIPath:     "/api/archive",
		Actions: []Action{
			{
				Name: "list",
				Description: "List archived records, newest archive first, with the visible kinds and their counts. " +
					"Needs Tenant.View or Tenant.Update; kinds you cannot view are left out",
				ToolName:   "ListArchive",
				HTTPMethod: "GET",
				Flags: []FlagDef{
					{Name: "kind", Description: "Only this kind: " + strings.Join(archiveKinds, ", "), Type: "string", QueryName: "kind"},
					{Name: "search", Description: "Text the record name contains", Type: "string", QueryName: "search"},
					{Name: "archived-by", Description: "Only records archived by this user GUID", Type: "uuid", QueryName: "archivedByUserGuid"},
					{Name: "from", Description: "Archived on or after this UTC date-time (ISO 8601)", Type: "string", QueryName: "from"},
					{Name: "to", Description: "Archived on or before this UTC date-time (ISO 8601)", Type: "string", QueryName: "to"},
					{Name: "page", Description: "Page number, from 1", Type: "int", Default: 1, QueryName: "page"},
					{Name: "page-size", Description: "Rows per page, 1 to 100", Type: "int", Default: 25, QueryName: "pageSize"},
				},
			},
			{
				Name:            "get",
				Description:     "Show one archived record with who archived it and every link. Needs Tenant.View or Tenant.Update and the kind's view permission",
				ToolName:        "GetArchivedRecord",
				HTTPMethod:      "GET",
				RESTPath:        "{kind}/{recordGuid}",
				RejectExtraArgs: true,
				Args: []ArgDef{
					{Name: "kind", Description: "The record kind: " + strings.Join(archiveKinds, ", "), Required: true, Type: "string", AllowedValues: archiveKinds},
					{Name: "recordGuid", Description: "GUID of the archived record", Required: true, Type: "non-empty-uuid"},
				},
			},
		},
	})
}
