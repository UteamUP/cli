package registry

import "testing"

func TestContactIceAssignmentsAction(t *testing.T) {
	d := findDomain("contact")
	if d == nil {
		t.Fatal("expected contact domain to be registered")
	}

	for _, action := range d.Actions {
		if action.Name != "ice-assignments" {
			continue
		}
		if action.ToolName != "UteamupEmergencycontactAssignmentsForContact" {
			t.Errorf("tool = %q", action.ToolName)
		}
		if action.RESTBasePath != "/api/emergencycontact" {
			t.Errorf("RESTBasePath = %q, want /api/emergencycontact", action.RESTBasePath)
		}
		if action.RESTPath != "by-contact/{guid}" {
			t.Errorf("RESTPath = %q", action.RESTPath)
		}
		if len(action.Args) != 1 || action.Args[0].Name != "guid" || action.Args[0].Type != "uuid" {
			t.Errorf("expected a single public guid arg")
		}
		return
	}

	t.Fatal("expected ice-assignments action on the contact domain")
}

func TestContactIceAssignmentsSyncAction(t *testing.T) {
	d := findDomain("contact")
	if d == nil {
		t.Fatal("expected contact domain to be registered")
	}

	for _, action := range d.Actions {
		if action.Name != "ice-assignments-sync" {
			continue
		}
		if action.ToolName != "UteamupEmergencycontactAssignmentsSyncForContact" {
			t.Errorf("tool = %q", action.ToolName)
		}
		if action.HTTPMethod != "PUT" || action.RESTPath != "by-contact/{guid}/assignments" {
			t.Errorf("route = %s %s", action.HTTPMethod, action.RESTPath)
		}
		if len(action.Flags) != 1 || !action.Flags[0].Required || !action.Flags[0].RootJSONObjectFile {
			t.Errorf("expected one required reviewed JSON body flag")
		}
		return
	}

	t.Fatal("expected ice-assignments-sync action on the contact domain")
}

func TestContactTypeDomainIsGuidFirstAndMatchesRestController(t *testing.T) {
	d := findDomain("contact-type")
	if d == nil {
		t.Fatal("expected contact-type domain to be registered")
	}
	if d.APIPath != "/api/contacttype" {
		t.Fatalf("APIPath = %q", d.APIPath)
	}

	for _, actionName := range []string{"get", "update", "delete"} {
		var found *Action
		for index := range d.Actions {
			if d.Actions[index].Name == actionName {
				found = &d.Actions[index]
				break
			}
		}
		if found == nil {
			t.Fatalf("missing %s action", actionName)
		}
		if found.RESTPath != "{contactTypeGuid}" {
			t.Errorf("%s path = %q", actionName, found.RESTPath)
		}
		if len(found.Args) != 1 || found.Args[0].Name != "contactTypeGuid" {
			t.Errorf("%s must use one public contactTypeGuid arg", actionName)
		}
	}
}

func TestContactTypeMutationsRequireOriginalReviewedNormalContract(t *testing.T) {
	d := findDomain("contact-type")
	for _, name := range []string{"create", "update", "delete"} {
		t.Run(name, func(t *testing.T) {
			var found *Action
			for index := range d.Actions {
				if d.Actions[index].Name == name {
					found = &d.Actions[index]
					break
				}
			}
			if found == nil || found.MCPOnly {
				t.Fatal("expected existing normal REST action")
			}
			wantTool := map[string]string{"create": "UteamupContacttypeCreate", "update": "UteamupContacttypeUpdate", "delete": "UteamupContacttypeDelete"}[name]
			if found.ToolName != wantTool {
				t.Fatal("tool mirror must match actual backend public name")
			}
			flags := map[string]FlagDef{}
			for _, flag := range found.Flags {
				flags[flag.Name] = flag
			}
			if !flags["idempotency-key"].Required || flags["idempotency-key"].Type != "uuid" || flags["mutation-outcome-version"].Default != 1 {
				t.Fatal("original operation identity and explicit receipt version required")
			}
			if !flags["confirm"].Required || !flags["confirm"].MustBeTrue || !flags["confirm"].LocalOnly {
				t.Fatal("reviewed mutation requires local explicit confirmation")
			}
			if name == "create" {
				if _, present := flags["expected-updated-at"]; present {
					t.Fatal("create must not invent a target revision")
				}
			} else if !flags["expected-updated-at"].Required || flags["expected-updated-at"].Type != "string" {
				t.Fatal("original seven-tick revision must remain a required wire string")
			}
		})
	}
}

func TestContactMutationsAreGuidFirstReviewedNormalRoutes(t *testing.T) {
	d := findDomain("contact")
	for _, name := range []string{"create", "update", "delete"} {
		t.Run(name, func(t *testing.T) {
			var found *Action
			for i := range d.Actions {
				if d.Actions[i].Name == name {
					found = &d.Actions[i]
					break
				}
			}
			if found == nil {
				t.Fatal("missing action")
			}
			want := map[string]string{"create": "UteamupContactCreate", "update": "UteamupContactUpdate", "delete": "UteamupContactDelete"}[name]
			if found.ToolName != want || found.MCPOnly {
				t.Fatal("must use existing normal REST with canonical tool mirror")
			}
			if name != "create" && (len(found.Args) != 1 || found.Args[0].Name != "contactGuid" || found.Args[0].Type != "uuid") {
				t.Fatal("public GUID required")
			}
			flags := map[string]FlagDef{}
			for _, flag := range found.Flags {
				flags[flag.Name] = flag
			}
			if !flags["idempotency-key"].Required || flags["idempotency-key"].Type != "uuid" {
				t.Fatal("original operation key required")
			}
			if flags["mutation-outcome-version"].Default != 1 {
				t.Fatal("explicit v1 receipt negotiation required")
			}
			if !flags["confirm"].Required || !flags["confirm"].MustBeTrue || !flags["confirm"].LocalOnly {
				t.Fatal("review confirmation required")
			}
			if name != "create" && (!flags["expected-updated-at"].Required || flags["expected-updated-at"].Type != "string") {
				t.Fatal("exact original wire revision required")
			}
		})
	}
}

func TestReviewedIceActionsUseNormalGuidRoutesAndExplicitNegotiation(t *testing.T) {
	d := findDomain("contact")
	for _, name := range []string{"ice-assignments-review", "ice-assignments-replace-reviewed"} {
		t.Run(name, func(t *testing.T) {
			var action *Action
			for index := range d.Actions {
				if d.Actions[index].Name == name {
					action = &d.Actions[index]
					break
				}
			}
			if action == nil || action.MCPOnly || action.RESTBasePath != "/api/emergencycontact" {
				t.Fatal("reviewed ICE must reuse the normal REST owner")
			}
			if len(action.Args) != 1 || action.Args[0].Name != "guid" || action.Args[0].Type != "uuid" {
				t.Fatal("only public Contact GUID is caller-controlled scope")
			}
			flags := map[string]FlagDef{}
			for _, flag := range action.Flags {
				flags[flag.Name] = flag
			}
			if name == "ice-assignments-review" {
				if action.ToolName != "UteamupEmergencycontactAssignmentsReview" || action.HTTPMethod != "GET" || action.RESTPath != "by-contact/{guid}" {
					t.Fatal("reviewed read tool/route mismatch")
				}
				version := flags["reviewed-version"]
				if version.Default != 1 || version.QueryName != "reviewedVersion" || version.Type != "int" {
					t.Fatal("normal GET must explicitly negotiate version1")
				}
				return
			}
			if action.ToolName != "UteamupEmergencycontactAssignmentsReplaceReviewed" || action.HTTPMethod != "PUT" || action.RESTPath != "by-contact/{guid}/assignments" {
				t.Fatal("reviewed replacement tool/route mismatch")
			}
			if !flags["from-json"].Required || !flags["from-json"].RootJSONObjectFile {
				t.Fatal("complete original request must be supplied unchanged")
			}
			key := flags["idempotency-key"]
			if !key.Required || key.Type != "uuid" || key.HeaderName != "Idempotency-Key" || key.MirrorHeaderInBody {
				t.Fatal("header must bind original key without overwriting a conflicting body identity")
			}
			confirmation := flags["confirm"]
			if !confirmation.Required || !confirmation.MustBeTrue || !confirmation.LocalOnly {
				t.Fatal("local explicit confirmation must not add a fifth JSON field")
			}
		})
	}
}
