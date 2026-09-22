package registry

func init() {
	Register(&Domain{
		Name:        "safety",
		Aliases:     []string{"osha", "incident", "ita"},
		Description: "OSHA safety incidents and ITA 300A/300/301 CSV. UteamUP generates; the employer files.",
		APIPath:     "/api/safetyincident",
		Actions: []Action{
			{
				Name:        "list",
				Description: "List workplace injury/illness incidents",
				ToolName:    "UteamupSafetyincidentList",
				RESTPath:    "",
				HTTPMethod:  "GET",
				Flags: []FlagDef{
					{Name: "skip", Description: "Skip offset", Default: 0, Type: "int", QueryName: "skip"},
					{Name: "take", Description: "Page size (server caps at 200)", Default: 50, Type: "int", QueryName: "take"},
					{Name: "status", Description: "Optional status filter", Type: "string", QueryName: "status"},
				},
			},
			{
				Name:        "get",
				Description: "Get a safety incident by public GUID",
				ToolName:    "UteamupSafetyincidentGet",
				RESTPath:    "by-guid/{guid}",
				HTTPMethod:  "GET",
				Args:        []ArgDef{{Name: "guid", Description: "Incident GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "create",
				Description: "Create a draft safety incident. Recordability stays unset until classify.",
				ToolName:    "UteamupSafetyincidentCreate",
				RESTPath:    "",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "JSON file containing SafetyIncidentCreateModel", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "group-get",
				Description: "Get every injured person's case for one event",
				ToolName:    "UteamupSafetyincidentGroupGet",
				RESTPath:    "group/by-guid/{eventGroupGuid}",
				HTTPMethod:  "GET",
				Args:        []ArgDef{{Name: "eventGroupGuid", Description: "Shared event GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "group-create",
				Description: "Create one event with a separate case for each injured person",
				ToolName:    "UteamupSafetyincidentGroupCreate",
				RESTPath:    "group",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "JSON file containing SafetyIncidentGroupCreateModel", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "classify",
				Description: "Human-classify OSHA 1904 recordability. Never auto-decides.",
				ToolName:    "UteamupSafetyincidentClassify",
				RESTPath:    "by-guid/{guid}/classify",
				HTTPMethod:  "POST",
				Args:        []ArgDef{{Name: "guid", Description: "Incident GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "JSON file containing SafetyIncidentClassifyModel", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "ita-export",
				Description: "Prepare an OSHA ITA 300A CSV, and 300/301 case CSV when required or --include-cases. Employer files; UteamUP does not auto-file.",
				ToolName:    "UteamupOshaItaExport",
				RESTPath:    "ita/export",
				HTTPMethod:  "POST",
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "JSON file containing OshaItaExportRequestModel", Required: true, Type: "string", JSONFile: true},
					{Name: "include-cases", BodyName: "includeCases", Description: "Also prepare Form 300/301 case CSV. Appendix B 100+ always includes cases. Employer files; UteamUP does not auto-file.", Type: "bool", Default: false},
				},
			},
			{
				Name:        "contacts",
				Description: "List the contacts attached to an incident and the role each holds",
				ToolName:    "UteamupSafetyincidentContactsList",
				RESTPath:    "by-guid/{guid}/contacts",
				HTTPMethod:  "GET",
				Args:        []ArgDef{{Name: "guid", Description: "Incident GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "link-contact",
				Description: "Attach a tenant contact to an incident as legal, insurer, witness, physician or other",
				ToolName:    "UteamupSafetyincidentContactsLink",
				RESTPath:    "by-guid/{guid}/contacts",
				HTTPMethod:  "POST",
				Args:        []ArgDef{{Name: "guid", Description: "Incident GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "contact", BodyName: "contactGuid", Description: "Contact GUID in the same tenant", Required: true, Type: "uuid"},
					{Name: "role", BodyName: "role", Description: "legal, insurer, witness, physician or other", Required: true, Type: "string"},
					{Name: "note", BodyName: "note", Description: "Why this person is attached", Type: "string"},
				},
			},
			{
				Name:        "unlink-contact",
				Description: "Remove a contact from an incident. The contact itself is untouched.",
				ToolName:    "UteamupSafetyincidentContactsUnlink",
				RESTPath:    "by-guid/{guid}/contacts/{linkGuid}",
				HTTPMethod:  "DELETE",
				Args: []ArgDef{
					{Name: "guid", Description: "Incident GUID", Required: true, Type: "uuid"},
					{Name: "linkGuid", Description: "The link GUID, not the contact GUID", Required: true, Type: "uuid"},
				},
			},
			{
				Name:        "legal-reports",
				Description: "List the legal review reports written about an incident",
				ToolName:    "UteamupSafetyincidentLegalreportList",
				RESTPath:    "by-guid/{guid}/legal-report",
				HTTPMethod:  "GET",
				Args:        []ArgDef{{Name: "guid", Description: "Incident GUID", Required: true, Type: "uuid"}},
			},
			{
				Name:        "legal-report",
				Description: "Save a legal review report on an incident. Never changes OSHA recordability.",
				ToolName:    "UteamupSafetyincidentLegalreportCreate",
				RESTPath:    "by-guid/{guid}/legal-report",
				HTTPMethod:  "POST",
				Args:        []ArgDef{{Name: "guid", Description: "Incident GUID", Required: true, Type: "uuid"}},
				Flags: []FlagDef{
					{Name: "from-json", BodyName: "model", Description: "JSON file containing SafetyIncidentLegalReportCreateModel", Required: true, Type: "string", JSONFile: true},
				},
			},
			{
				Name:        "draft-legal-report",
				Description: "Draft a legal review with UPMate. Spends tenant AI credits and saves nothing; accept it with legal-report.",
				ToolName:    "UteamupSafetyincidentLegalreportDraft",
				RESTPath:    "by-guid/{guid}/legal-report/draft",
				HTTPMethod:  "POST",
				Args:        []ArgDef{{Name: "guid", Description: "Incident GUID", Required: true, Type: "uuid"}},
			},
		},
	})
}
