package registry

import "testing"

// There is no /api/user list or single-user route, so `ut user list` and `ut user get` 404'd.
// Both read the tenant's member routes, and get takes the member's public user GUID.
func TestUserListAndGetUseTheTenantMemberRoutes(t *testing.T) {
	const member = "3fa85f64-5717-4562-b3fc-2c963f66afa6"
	d := findDomain("user")
	if d == nil {
		t.Fatal("expected user domain to be registered")
	}

	cases := []struct {
		action string
		args   map[string]any
		want   string
	}{
		{"list", map[string]any{}, "/api/tenant/users"},
		{"get", map[string]any{"guid": member}, "/api/tenant/users/" + member},
	}
	for _, c := range cases {
		action := findDomainAction(t, "user", c.action)
		if got, _ := buildRESTPath(d, *action, c.args); got != c.want {
			t.Errorf("%s path = %q, want %q", c.action, got, c.want)
		}
		if HTTPMethod[action.Name] != "GET" || action.HTTPMethod != "" {
			t.Errorf("%s should be a plain GET", c.action)
		}
	}

	get := findDomainAction(t, "user", "get")
	if len(get.Args) != 1 || get.Args[0].Name != "guid" || get.Args[0].Type != "non-empty-uuid" {
		t.Errorf("get should take one required non-empty-uuid guid arg, got %+v", get.Args)
	}
	// The member route has no paging or filter, so list must not offer flags it would ignore.
	if list := findDomainAction(t, "user", "list"); len(list.Flags) != 0 {
		t.Errorf("list flags = %+v, want none", list.Flags)
	}
}

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
