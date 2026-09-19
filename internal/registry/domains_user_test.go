package registry

import "testing"

// Reading a colleague's next of kin is a cross-domain action: it lives on the user domain for
// discoverability but is served by /api/emergencycontact, so the base-path override is the thing
// that makes it work at all.
func TestUserEmergencyContactsAction(t *testing.T) {
	d := findDomain("user")
	if d == nil {
		t.Fatal("expected user domain to be registered")
	}

	for _, a := range d.Actions {
		if a.Name != "emergency-contacts" {
			continue
		}
		if a.ToolName != "UteamupEmergencycontactListForUser" {
			t.Errorf("tool = %q", a.ToolName)
		}
		if a.RESTBasePath != "/api/emergencycontact" {
			t.Errorf("RESTBasePath = %q, want /api/emergencycontact", a.RESTBasePath)
		}
		if a.RESTPath != "by-user/{guid}" {
			t.Errorf("RESTPath = %q", a.RESTPath)
		}
		if len(a.Args) != 1 || a.Args[0].Name != "guid" {
			t.Errorf("expected a single guid arg")
		}
		// The reason is what the subject sees beside the reader's name.
		var hasReason bool
		for _, f := range a.Flags {
			if f.Name == "reason" && f.QueryName == "reason" {
				hasReason = true
			}
		}
		if !hasReason {
			t.Error("expected a reason flag mapped to the reason query parameter")
		}
		return
	}
	t.Fatal("expected emergency-contacts action on the user domain")
}
