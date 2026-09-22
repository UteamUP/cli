package registry

import "testing"

func TestSafetyDomainRegistered(t *testing.T) {
	d := findDomain("safety")
	if d == nil {
		t.Fatal("expected safety domain to be registered")
	}
	if d.APIPath != "/api/safetyincident" {
		t.Fatalf("safety APIPath = %q, want /api/safetyincident", d.APIPath)
	}
}

func TestSafetyDomainActions(t *testing.T) {
	d := findDomain("safety")
	if d == nil {
		t.Fatal("expected safety domain to be registered")
	}

	expected := map[string]string{
		"list":         "UteamupSafetyincidentList",
		"get":          "UteamupSafetyincidentGet",
		"create":       "UteamupSafetyincidentCreate",
		"group-get":    "UteamupSafetyincidentGroupGet",
		"group-create": "UteamupSafetyincidentGroupCreate",
		"classify":     "UteamupSafetyincidentClassify",
		"ita-export":   "UteamupOshaItaExport",

		// Added with the people-and-legal change. The MCP tool and the CLI action must move
		// together or the CLI silently drifts from what an agent can do.
		"contacts":           "UteamupSafetyincidentContactsList",
		"link-contact":       "UteamupSafetyincidentContactsLink",
		"unlink-contact":     "UteamupSafetyincidentContactsUnlink",
		"legal-reports":      "UteamupSafetyincidentLegalreportList",
		"legal-report":       "UteamupSafetyincidentLegalreportCreate",
		"draft-legal-report": "UteamupSafetyincidentLegalreportDraft",
	}

	actionMap := make(map[string]string)
	for _, a := range d.Actions {
		actionMap[a.Name] = a.ToolName
	}

	for name, tool := range expected {
		if actual, ok := actionMap[name]; !ok {
			t.Errorf("missing action %q", name)
		} else if actual != tool {
			t.Errorf("action %q: expected tool %q, got %q", name, tool, actual)
		}
	}
}

func TestSafetyGetUsesGuidPath(t *testing.T) {
	d := findDomain("safety")
	if d == nil {
		t.Fatal("expected safety domain to be registered")
	}
	action := findAction(d, "get")
	if action == nil {
		t.Fatal("expected get action on safety domain")
	}
	path, _ := buildRESTPath(d, *action, map[string]any{
		"guid": "11111111-1111-1111-1111-111111111111",
	})
	if path != "/api/safetyincident/by-guid/11111111-1111-1111-1111-111111111111" {
		t.Fatalf("safety get path = %q", path)
	}
}

func TestSafetyGroupActionsUseSharedEventGuidAndJsonModel(t *testing.T) {
	d := findDomain("safety")
	if d == nil {
		t.Fatal("expected safety domain")
	}
	get := findAction(d, "group-get")
	if get == nil {
		t.Fatal("expected group-get action")
	}
	path, _ := buildRESTPath(d, *get, map[string]any{
		"eventGroupGuid": "11111111-1111-1111-1111-111111111111",
	})
	if path != "/api/safetyincident/group/by-guid/11111111-1111-1111-1111-111111111111" {
		t.Fatalf("group-get path = %q", path)
	}
	create := findAction(d, "group-create")
	if create == nil || create.RESTPath != "group" || create.HTTPMethod != "POST" ||
		len(create.Flags) != 1 || create.Flags[0].BodyName != "model" {
		t.Fatalf("group-create action = %+v", create)
	}
}

func TestSafetyItaExportIncludeCasesFlag(t *testing.T) {
	d := findDomain("safety")
	if d == nil {
		t.Fatal("expected safety domain to be registered")
	}
	action := findAction(d, "ita-export")
	if action == nil {
		t.Fatal("expected ita-export action on safety domain")
	}

	var flag *FlagDef
	for i := range action.Flags {
		if action.Flags[i].Name == "include-cases" {
			flag = &action.Flags[i]
			break
		}
	}
	if flag == nil {
		t.Fatal("expected --include-cases on safety ita-export")
	}
	if flag.Type != "bool" || flag.BodyName != "includeCases" {
		t.Errorf("include-cases type/body = %q %q, want bool includeCases", flag.Type, flag.BodyName)
	}
	if v, ok := flag.Default.(bool); !ok || v {
		t.Errorf("include-cases default = %v (%T), want false", flag.Default, flag.Default)
	}
}

// The unlink route takes the link GUID and the incident GUID, in that order. Passing the contact
// GUID where the link GUID belongs would delete nothing and report success, so the arg names are
// worth pinning.
func TestSafetyUnlinkContactTakesBothGuids(t *testing.T) {
	d := findDomain("safety")
	if d == nil {
		t.Fatal("expected safety domain to be registered")
	}

	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "unlink-contact" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected unlink-contact action")
	}

	if len(action.Args) != 2 {
		t.Fatalf("unlink-contact args = %d, want 2", len(action.Args))
	}
	if action.Args[0].Name != "guid" || action.Args[1].Name != "linkGuid" {
		t.Errorf("unlink-contact args = %q/%q, want guid/linkGuid",
			action.Args[0].Name, action.Args[1].Name)
	}
	if action.RESTPath != "by-guid/{guid}/contacts/{linkGuid}" {
		t.Errorf("unlink-contact RESTPath = %q", action.RESTPath)
	}
}

// Drafting spends tenant AI credits, so it must be a POST that carries no body of its own — the
// incident is the entire input.
func TestSafetyDraftLegalReportIsAPostWithNoBody(t *testing.T) {
	d := findDomain("safety")
	if d == nil {
		t.Fatal("expected safety domain to be registered")
	}

	for _, a := range d.Actions {
		if a.Name != "draft-legal-report" {
			continue
		}
		if a.HTTPMethod != "POST" {
			t.Errorf("draft-legal-report HTTPMethod = %q, want POST", a.HTTPMethod)
		}
		if len(a.Flags) != 0 {
			t.Errorf("draft-legal-report should take no flags, got %d", len(a.Flags))
		}
		return
	}
	t.Fatal("expected draft-legal-report action")
}
