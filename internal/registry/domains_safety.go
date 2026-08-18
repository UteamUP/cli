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
		},
	})
}
