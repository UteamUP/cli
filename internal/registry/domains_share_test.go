package registry

import "testing"

// The share domain mirrors EntityShareController (/api/share/*). Two things are
// load-bearing: --entity-type fills a route segment, so its value set IS the escaping;
// and --section carries the scope grid, so losing it silently widens every CLI-created
// share to the default grant.

func shareDomain(t *testing.T) *Domain {
	t.Helper()
	d := findDomain("share")
	if d == nil {
		t.Fatal("expected share domain to be registered")
	}
	return d
}

func TestShareDomainRoutesUnderShare(t *testing.T) {
	d := shareDomain(t)
	if d.APIPath != "/api/share" {
		t.Errorf("APIPath = %q, want /api/share", d.APIPath)
	}
}

func TestShareActionsMatchBackendContract(t *testing.T) {
	d := shareDomain(t)
	cases := []struct{ name, tool, method, path string }{
		{"create", "UteamupShareCreate", "POST", "{entityType}"},
		{"list", "UteamupShareListByTarget", "", "{entityType}/by-target/{targetGuid}"},
		{"update", "UteamupShareUpdate", "PATCH", "{shareGuid}"},
		{"revoke", "UteamupShareRevoke", "DELETE", "{shareGuid}"},
		{"shared-with-me", "UteamupShareSharedWithMe", "", "shared-with-me"},
	}

	for _, c := range cases {
		a := findAction(d, c.name)
		if a == nil {
			t.Errorf("expected %s action on share", c.name)
			continue
		}
		if a.ToolName != c.tool {
			t.Errorf("%s: tool = %q, want %q", c.name, a.ToolName, c.tool)
		}
		if a.HTTPMethod != c.method || a.RESTPath != c.path {
			t.Errorf("%s: route = %s %q, want %s %q", c.name, a.HTTPMethod, a.RESTPath, c.method, c.path)
		}
	}
	if len(d.Actions) != len(cases) {
		t.Errorf("share has %d actions, want %d", len(d.Actions), len(cases))
	}
}

// The kind picks the route AND the permission pair the backend requires, so every
// management action has to state it. Only shared-with-me is exempt: its REST route
// returns both kinds and accepts no filter, so a flag there would be a silent no-op.
func TestShareManagementActionsAllRequireTheEntityType(t *testing.T) {
	d := shareDomain(t)
	for _, name := range []string{"create", "list", "update", "revoke"} {
		action := findAction(d, name)
		if action == nil {
			t.Fatalf("missing share action %q", name)
		}
		flag, ok := flagsToMap(action.Flags)["entity-type"]
		if !ok {
			t.Errorf("%s: missing --entity-type", name)
			continue
		}
		if !flag.Required {
			t.Errorf("%s: --entity-type must be required", name)
		}
		if len(flag.AllowedValues) != 2 ||
			!containsExact(flag.AllowedValues, "asset") ||
			!containsExact(flag.AllowedValues, "assetgroup") {
			t.Errorf("%s: --entity-type AllowedValues = %v, want [asset assetgroup]", name, flag.AllowedValues)
		}
	}

	sharedWithMe := findAction(d, "shared-with-me")
	if sharedWithMe == nil {
		t.Fatal("missing share action shared-with-me")
	}
	if _, ok := flagsToMap(sharedWithMe.Flags)["entity-type"]; ok {
		t.Error("shared-with-me must not offer --entity-type: the REST route takes no filter")
	}
}

// Path expansion does not escape, so a route-filling flag without a closed value set
// would let a caller reshape the URL.
func TestShareEntityTypeIsConstrainedWhereverItFillsTheRoute(t *testing.T) {
	d := shareDomain(t)
	for _, name := range []string{"create", "list"} {
		action := findAction(d, name)
		if action == nil {
			t.Fatalf("missing share action %q", name)
		}
		flag := flagsToMap(action.Flags)["entity-type"]
		if flag.BodyName != "entityType" {
			t.Errorf("%s: --entity-type BodyName = %q, want entityType (it fills the route)", name, flag.BodyName)
		}
		if err := validateActionDefinition(*action); err != nil {
			t.Errorf("%s definition is invalid: %v", name, err)
		}
	}

	// The two GUID-addressed routes are kind-agnostic on the wire, so the kind travels
	// on the query string and never lands in a body the update model does not declare.
	for _, name := range []string{"update", "revoke"} {
		action := findAction(d, name)
		if action == nil {
			t.Fatalf("missing share action %q", name)
		}
		flag := flagsToMap(action.Flags)["entity-type"]
		if flag.QueryName != "entityType" || flag.BodyName != "" {
			t.Errorf("%s: --entity-type must be a query flag, got QueryName=%q BodyName=%q", name, flag.QueryName, flag.BodyName)
		}
	}
}

func TestShareRoutesExpandFromTheirFlagsAndArgs(t *testing.T) {
	d := shareDomain(t)
	cases := []struct {
		action string
		args   map[string]any
		want   string
	}{
		{"create", map[string]any{"entityType": "assetgroup"}, "/api/share/assetgroup"},
		{"list", map[string]any{"entityType": "asset", "targetGuid": "a-1"}, "/api/share/asset/by-target/a-1"},
		{"update", map[string]any{"shareGuid": "s-1"}, "/api/share/s-1"},
		{"revoke", map[string]any{"shareGuid": "s-1"}, "/api/share/s-1"},
		{"shared-with-me", map[string]any{}, "/api/share/shared-with-me"},
	}

	for _, c := range cases {
		action := findAction(d, c.action)
		if action == nil {
			t.Fatalf("missing share action %q", c.action)
		}
		got, _ := buildRESTPath(d, *action, c.args)
		if got != c.want {
			t.Errorf("%s path = %q, want %q", c.action, got, c.want)
		}
	}
}

// Read is the safe default everywhere else in the product; the CLI must not be the one
// surface that hands out more than the operator typed.
func TestShareCreateCarriesTheScopeGridAndDefaultsToRead(t *testing.T) {
	d := shareDomain(t)
	create := findAction(d, "create")
	if create == nil {
		t.Fatal("expected create action on share")
	}

	flags := flagsToMap(create.Flags)
	target, ok := flags["target-guid"]
	if !ok || !target.Required || target.BodyName != "targetGuid" || target.Type != "non-empty-uuid" {
		t.Errorf("--target-guid must be a required non-empty GUID mapped to targetGuid, got %+v", target)
	}
	section, ok := flags["section"]
	if !ok || section.BodyName != "sections" || section.Type != "stringSlice" {
		t.Errorf("--section must be a repeatable slice mapped to sections, got %+v", section)
	}
	if flags["access-level"].Default != "read" {
		t.Errorf("access-level default = %v, want read", flags["access-level"].Default)
	}
}

// An update replaces the grid wholesale, so the flag has to exist here too — otherwise
// the only way to narrow a share from the CLI would be to revoke and re-create it.
func TestShareUpdateCanReplaceTheScopeGrid(t *testing.T) {
	d := shareDomain(t)
	update := findAction(d, "update")
	if update == nil {
		t.Fatal("expected update action on share")
	}

	flags := flagsToMap(update.Flags)
	section, ok := flags["section"]
	if !ok || section.BodyName != "sections" || section.Type != "stringSlice" {
		t.Errorf("--section must be a repeatable slice mapped to sections, got %+v", section)
	}
	expiry, ok := flags["expires-in-days"]
	if !ok || expiry.BodyName != "expiresInDays" || expiry.Type != "int" {
		t.Errorf("--expires-in-days must be an int mapped to expiresInDays, got %+v", expiry)
	}
}
