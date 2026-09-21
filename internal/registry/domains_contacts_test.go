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
		if found.RESTPath != "{externalGuid}" {
			t.Errorf("%s path = %q", actionName, found.RESTPath)
		}
		if len(found.Args) != 1 || found.Args[0].Name != "externalGuid" {
			t.Errorf("%s must use one public externalGuid arg", actionName)
		}
	}
}
