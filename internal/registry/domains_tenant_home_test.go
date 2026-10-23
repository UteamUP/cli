package registry

import "testing"

// The UteamUP Home tenant commands mirror POST /api/tenant/{guid}/change-owner and
// PUT /api/tenant/users/{userGuid}/home-membership. Both identify people by user GUID.
func TestTenantHomeCommandsMatchTheBackendRoutes(t *testing.T) {
	d := findDomain("tenant")
	if d == nil {
		t.Fatal("expected tenant domain to be registered")
	}

	cases := []struct {
		action, method, path, arg string
		flags                     map[string]string
	}{
		{"transfer-ownership", "POST", "{tenantGuid}/change-owner", "tenantGuid",
			map[string]string{"new-owner-user-guid": "userGuid"}},
		{"home-membership-set", "PUT", "users/{userGuid}/home-membership", "userGuid",
			map[string]string{"access-ends-at": "accessEndsAt", "residence-location-guid": "residenceLocationGuid"}},
	}
	for _, c := range cases {
		action := findAction(d, c.action)
		if action == nil {
			t.Errorf("expected %s action on tenant", c.action)
			continue
		}
		if action.HTTPMethod != c.method || action.RESTPath != c.path {
			t.Errorf("%s routes to %s %s, want %s %s", c.action, action.HTTPMethod, action.RESTPath, c.method, c.path)
		}
		if len(action.Args) != 1 || action.Args[0].Name != c.arg || action.Args[0].Type != "non-empty-uuid" {
			t.Errorf("%s must take the %s GUID as its only positional argument", c.action, c.arg)
		}
		flags := flagsToMap(action.Flags)
		for name, body := range c.flags {
			if flag, ok := flags[name]; !ok || flag.BodyName != body {
				t.Errorf("--%s on %s must map to %s", name, c.action, body)
			}
		}
		for name := range flags {
			if name == "user-id" || name == "tenant-id" {
				t.Errorf("%s must not take an integer or identity-key id (--%s)", c.action, name)
			}
		}
	}

	if !flagsToMap(findAction(d, "transfer-ownership").Flags)["new-owner-user-guid"].Required {
		t.Error("--new-owner-user-guid must be required")
	}
}

// The household children commands mirror /api/tenant/home-children; a child is removed by GUID.
func TestTenantHomeChildrenCommandsMatchTheBackendRoutes(t *testing.T) {
	d := findDomain("tenant")
	if d == nil {
		t.Fatal("expected tenant domain to be registered")
	}
	routes := map[string][2]string{
		"home-children-list":   {"GET", "home-children"},
		"home-children-add":    {"POST", "home-children"},
		"home-children-remove": {"DELETE", "home-children/{childGuid}"},
	}
	for name, route := range routes {
		action := findAction(d, name)
		if action == nil {
			t.Errorf("expected %s action on tenant", name)
			continue
		}
		if action.HTTPMethod != route[0] || action.RESTPath != route[1] {
			t.Errorf("%s routes to %s %s, want %s %s", name, action.HTTPMethod, action.RESTPath, route[0], route[1])
		}
	}

	add := flagsToMap(findAction(d, "home-children-add").Flags)
	if !add["name"].Required || add["name"].BodyName != "name" {
		t.Error("--name must be required and map to name")
	}
	if add["birth-year"].Type != "int" || add["birth-year"].Default != nil {
		t.Error("--birth-year must be an optional int that is not sent unless given")
	}
	remove := findAction(d, "home-children-remove")
	if len(remove.Args) != 1 || remove.Args[0].Name != "childGuid" || remove.Args[0].Type != "non-empty-uuid" {
		t.Error("home-children-remove must take the child GUID as its only argument")
	}
}
