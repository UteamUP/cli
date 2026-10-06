package registry

import "testing"

// The shutdown-rule domain mirrors AssetShutdownRuleController (/api/assetshutdownrule/*) plus
// the workorder shutdown override (PUT /api/workorder/by-guid/{guid}/shutdown-override).
// Every identifier is a GUID; the override is routed under /api/workorder, not the rule base.

func shutdownRuleDomain(t *testing.T) *Domain {
	t.Helper()
	d := findDomain("shutdown-rule")
	if d == nil {
		t.Fatal("expected shutdown-rule domain to be registered")
	}
	return d
}

func TestShutdownRuleDomainRoutesUnderAssetshutdownrule(t *testing.T) {
	d := shutdownRuleDomain(t)
	if d.APIPath != "/api/assetshutdownrule" {
		t.Errorf("APIPath = %q, want /api/assetshutdownrule", d.APIPath)
	}
}

func TestShutdownRuleActionsMatchBackendContract(t *testing.T) {
	d := shutdownRuleDomain(t)
	cases := []struct{ name, tool, method, base, path string }{
		{"list", "UteamupAssetShutdownRuleList", "", "", "asset/by-guid/{assetGuid}"},
		{"create", "UteamupAssetShutdownRuleCreate", "", "", ""},
		{"update", "UteamupAssetShutdownRuleUpdate", "", "", "by-guid/{ruleGuid}"},
		{"delete", "UteamupAssetShutdownRuleDelete", "", "", "by-guid/{ruleGuid}"},
		{"set-override", "UteamupWorkorderSetShutdownOverride", "PUT", "/api/workorder", "by-guid/{workorderGuid}/shutdown-override"},
	}

	for _, c := range cases {
		a := findAction(d, c.name)
		if a == nil {
			t.Errorf("expected %s action on shutdown-rule", c.name)
			continue
		}
		if a.ToolName != c.tool {
			t.Errorf("%s: tool = %q, want %q", c.name, a.ToolName, c.tool)
		}
		if a.HTTPMethod != c.method || a.RESTBasePath != c.base || a.RESTPath != c.path {
			t.Errorf("%s: route = %s %q%q, want %s %q%q", c.name, a.HTTPMethod, a.RESTBasePath, a.RESTPath, c.method, c.base, c.path)
		}
	}
	if len(d.Actions) != len(cases) {
		t.Errorf("shutdown-rule has %d actions, want %d", len(d.Actions), len(cases))
	}
}

func TestShutdownRuleRoutesExpandFromTheirGuidArgs(t *testing.T) {
	d := shutdownRuleDomain(t)
	cases := []struct {
		action string
		args   map[string]any
		want   string
	}{
		{"list", map[string]any{"assetGuid": "a-1"}, "/api/assetshutdownrule/asset/by-guid/a-1"},
		{"update", map[string]any{"ruleGuid": "r-1"}, "/api/assetshutdownrule/by-guid/r-1"},
		{"delete", map[string]any{"ruleGuid": "r-1"}, "/api/assetshutdownrule/by-guid/r-1"},
		{"set-override", map[string]any{"workorderGuid": "w-1"}, "/api/workorder/by-guid/w-1/shutdown-override"},
	}

	for _, c := range cases {
		action := findAction(d, c.action)
		if action == nil {
			t.Fatalf("missing shutdown-rule action %q", c.action)
		}
		got, consumed := buildRESTPath(d, *action, c.args)
		if got != c.want {
			t.Errorf("%s path = %q, want %q", c.action, got, c.want)
		}
		if len(consumed) != 1 {
			t.Errorf("%s consumed = %v, want one path arg", c.action, consumed)
		}
		for _, arg := range action.Args {
			if arg.Type != "non-empty-uuid" {
				t.Errorf("%s: positional %s has type %q, want non-empty-uuid", c.action, arg.Name, arg.Type)
			}
		}
	}

	create := findAction(d, "create")
	if create == nil {
		t.Fatal("missing shutdown-rule create")
	}
	if got, _ := buildRESTPath(d, *create, map[string]any{"assetGuid": "a-1"}); got != "/api/assetshutdownrule" {
		t.Errorf("create path = %q, want /api/assetshutdownrule", got)
	}
}

func TestShutdownRuleCreateAndUpdateCarryTheRuleFields(t *testing.T) {
	d := shutdownRuleDomain(t)
	for _, actionName := range []string{"create", "update"} {
		action := findAction(d, actionName)
		if action == nil {
			t.Fatalf("missing shutdown-rule %s", actionName)
		}
		flags := flagsToMap(action.Flags)
		for name, want := range map[string]struct {
			bodyName string
			typ      string
			required bool
		}{
			"shutdown-type":           {"shutdownType", "string", true},
			"workorder-template-guid": {"workorderTemplateGuid", "non-empty-uuid", false},
			"maintenance-type":        {"maintenanceType", "string", false},
			"instructions":            {"", "string", false},
			"is-active":               {"isActive", "bool", false},
		} {
			flag, ok := flags[name]
			if !ok {
				t.Errorf("%s: missing --%s", actionName, name)
				continue
			}
			if flag.BodyName != want.bodyName || flag.Type != want.typ || flag.Required != want.required {
				t.Errorf("%s --%s = {%q %q %v}, want {%q %q %v}", actionName, name,
					flag.BodyName, flag.Type, flag.Required, want.bodyName, want.typ, want.required)
			}
		}
		// A bool default must be a bool literal or the registry's type assertion panics.
		if def, ok := flags["is-active"].Default.(bool); !ok || !def {
			t.Errorf("%s --is-active default = %#v, want true", actionName, flags["is-active"].Default)
		}
	}

	create := flagsToMap(findAction(d, "create").Flags)
	if flag := create["asset-guid"]; flag.BodyName != "assetGuid" || flag.Type != "non-empty-uuid" || !flag.Required {
		t.Errorf("create --asset-guid = %+v, want required non-empty-uuid assetGuid", flag)
	}
	if _, ok := flagsToMap(findAction(d, "update").Flags)["asset-guid"]; ok {
		t.Error("update must not take --asset-guid; a rule never moves between assets")
	}
}

func TestShutdownRuleShutdownTypeIsConstrainedToTheBackendNames(t *testing.T) {
	d := shutdownRuleDomain(t)
	for _, actionName := range []string{"create", "update", "set-override"} {
		flag, ok := flagsToMap(findAction(d, actionName).Flags)["shutdown-type"]
		if !ok {
			t.Fatalf("%s: missing --shutdown-type", actionName)
		}
		want := []string{"none", "partial", "full", "lockoutTagout"}
		if len(flag.AllowedValues) != len(want) {
			t.Fatalf("%s: allowed = %v, want %v", actionName, flag.AllowedValues, want)
		}
		for i := range want {
			if flag.AllowedValues[i] != want[i] {
				t.Errorf("%s: allowed[%d] = %q, want %q", actionName, i, flag.AllowedValues[i], want[i])
			}
		}
	}
	// Leaving the type out clears the override, so it cannot be required there.
	if flagsToMap(findAction(d, "set-override").Flags)["shutdown-type"].Required {
		t.Error("set-override --shutdown-type must be optional so the override can be cleared")
	}
}

func TestShutdownRuleTakesNoIdentityArguments(t *testing.T) {
	d := shutdownRuleDomain(t)
	for _, action := range d.Actions {
		for _, flag := range action.Flags {
			switch flag.Name {
			case "user-id", "tenant-id", "tenant-guid", "id":
				t.Errorf("%s: --%s must not exist; identity and tenant come from the token", action.Name, flag.Name)
			}
		}
	}
}
