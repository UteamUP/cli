package registry

import (
	"strings"
	"testing"
)

func TestCodingSystemDomainRegistered(t *testing.T) {
	var csDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "codingsystem" {
			csDomain = d
			break
		}
	}
	if csDomain == nil {
		t.Fatal("expected codingsystem domain to be registered")
	}

	// Verify aliases
	expectedAliases := map[string]bool{"cs": true, "coding": true}
	for _, alias := range csDomain.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias %q", alias)
		}
		delete(expectedAliases, alias)
	}
	if len(expectedAliases) > 0 {
		t.Errorf("missing aliases: %v", expectedAliases)
	}
}

func TestCodingSystemDomainActions(t *testing.T) {
	var csDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "codingsystem" {
			csDomain = d
			break
		}
	}
	if csDomain == nil {
		t.Fatal("expected codingsystem domain to be registered")
	}

	expectedActions := map[string]string{
		"list":                      "UteamupCodingsystemList",
		"tree":                      "UteamupCodingsystemTree",
		"search":                    "UteamupCodingsystemSearchAssets",
		"next-code":                 "UteamupCodingsystemNextCode",
		"assign":                    "UteamupCodingsystemAssignCode",
		"workorders":                "UteamupCodingsystemWorkorders",
		"create-workorder":          "UteamupCodingsystemCreateWorkorder",
		"materialize-register-code": "UteamupCodingsystemMaterializeRegisterCode",
		"set-designation-format":    "UteamupTenantcodingSetDesignationFormat",
		"profile-validate":          "UteamupTenantcodingProfileValidate",
		"profile-apply":             "UteamupTenantcodingProfileApply",
	}

	actionMap := make(map[string]string)
	for _, a := range csDomain.Actions {
		actionMap[a.Name] = a.ToolName
	}

	for name, tool := range expectedActions {
		if actual, ok := actionMap[name]; !ok {
			t.Errorf("missing action %q", name)
		} else if actual != tool {
			t.Errorf("action %q: expected tool %q, got %q", name, tool, actual)
		}
	}
}

func TestCodingSystemProfileActionsUseReviewedJSONFiles(t *testing.T) {
	want := map[string]struct {
		flagName string
		bodyName string
	}{
		"profile-validate": {flagName: "profile-file", bodyName: "profile"},
		"profile-apply":    {flagName: "review-file", bodyName: "review"},
	}

	for _, d := range DefaultRegistry.Domains() {
		if d.Name != "codingsystem" {
			continue
		}
		for _, action := range d.Actions {
			expected, ok := want[action.Name]
			if !ok {
				continue
			}
			if !action.MCPOnly {
				t.Errorf("%s must call the governed MCP tool", action.Name)
			}
			if len(action.Flags) != 1 {
				t.Fatalf("%s flags = %d, want 1", action.Name, len(action.Flags))
			}
			flag := action.Flags[0]
			if flag.Name != expected.flagName || flag.BodyName != expected.bodyName || !flag.Required || !flag.JSONFile {
				t.Errorf("%s flag = %#v", action.Name, flag)
			}
			delete(want, action.Name)
		}
	}
	if len(want) > 0 {
		t.Errorf("missing reviewed profile actions: %v", want)
	}
}

func TestCodingSystemSetDesignationFormatFlags(t *testing.T) {
	var action *Action
	for _, d := range DefaultRegistry.Domains() {
		if d.Name != "codingsystem" {
			continue
		}
		for i := range d.Actions {
			if d.Actions[i].Name == "set-designation-format" {
				action = &d.Actions[i]
			}
		}
	}
	if action == nil {
		t.Fatal("expected codingsystem set-designation-format action")
	}
	if !action.MCPOnly {
		t.Error("set-designation-format must call the MCP tool")
	}

	want := map[string]struct{ bodyName, flagType string }{
		"separator": {"separatorChar", "string"},
		"levels":    {"levelNames", "stringSlice"},
	}
	for _, flag := range action.Flags {
		expected, ok := want[flag.Name]
		if !ok {
			t.Errorf("unexpected flag %q", flag.Name)
			continue
		}
		if flag.BodyName != expected.bodyName || flag.Type != expected.flagType || !flag.Required {
			t.Errorf("flag %q: body %q type %q required %v", flag.Name, flag.BodyName, flag.Type, flag.Required)
		}
		delete(want, flag.Name)
	}
	if len(want) > 0 {
		t.Errorf("missing flags: %v", want)
	}
}

func TestCodingSystemTreeFlags(t *testing.T) {
	var csDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "codingsystem" {
			csDomain = d
			break
		}
	}
	if csDomain == nil {
		t.Fatal("expected codingsystem domain to be registered")
	}

	var treeAction *Action
	for i := range csDomain.Actions {
		if csDomain.Actions[i].Name == "tree" {
			treeAction = &csDomain.Actions[i]
			break
		}
	}
	if treeAction == nil {
		t.Fatal("expected tree action")
	}

	// Public identifiers are validated UUIDs and map to the registered MCP names.
	flagMap := make(map[string]FlagDef)
	for _, f := range treeAction.Flags {
		flagMap[f.Name] = f
	}

	csFlag, ok := flagMap["coding-system-guid"]
	if !ok {
		t.Fatal("missing coding-system-guid flag")
	}
	if !csFlag.Required {
		t.Error("coding-system-guid should be required")
	}
	if csFlag.Type != "uuid" || csFlag.BodyName != "codingSystemGuid" {
		t.Errorf("coding-system-guid = %#v, want uuid mapped to codingSystemGuid", csFlag)
	}

	parentFlag, ok := flagMap["parent-guid"]
	if !ok {
		t.Fatal("missing parent-guid flag")
	}
	if parentFlag.Required {
		t.Error("parent-guid should not be required")
	}
	if parentFlag.Type != "uuid" || parentFlag.BodyName != "parentGuid" {
		t.Errorf("parent-guid = %#v, want uuid mapped to parentGuid", parentFlag)
	}
}

func TestCodingSystemActionsUseRegisteredMCPGuidContracts(t *testing.T) {
	var csDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "codingsystem" {
			csDomain = d
			break
		}
	}
	if csDomain == nil {
		t.Fatal("expected codingsystem domain to be registered")
	}

	expectedGuidFlags := map[string]map[string]string{
		"next-code":        {"parent-guid": "parentEntryGuid"},
		"assign":           {"asset-guid": "assetGuid", "entry-guid": "codeCatalogEntryGuid"},
		"create-workorder": {"entry-guid": "codeCatalogEntryGuid"},
		// Register scaffolding rows have no GUID of their own, so the parent is
		// the only identifier on the wire — and it must stay a GUID.
		"materialize-register-code": {"coding-system-guid": "codingSystemGuid", "parent-guid": "parentEntryGuid"},
	}

	for _, action := range csDomain.Actions {
		if !action.MCPOnly {
			t.Errorf("action %q must use its registered MCP tool contract", action.Name)
		}
		for _, flag := range action.Flags {
			if strings.HasSuffix(flag.Name, "-id") {
				t.Errorf("action %q exposes forbidden integer identifier flag --%s", action.Name, flag.Name)
			}
		}

		for flagName, bodyName := range expectedGuidFlags[action.Name] {
			var found *FlagDef
			for i := range action.Flags {
				if action.Flags[i].Name == flagName {
					found = &action.Flags[i]
					break
				}
			}
			if found == nil {
				t.Errorf("action %q missing --%s", action.Name, flagName)
				continue
			}
			if found.Type != "uuid" || found.BodyName != bodyName {
				t.Errorf("action %q --%s = %#v, want uuid mapped to %s", action.Name, flagName, *found, bodyName)
			}
		}
	}
}

func TestWorkorderDomainHasByCodeAction(t *testing.T) {
	var woDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "workorder" {
			woDomain = d
			break
		}
	}
	if woDomain == nil {
		t.Fatal("expected workorder domain to be registered")
	}

	found := false
	for _, a := range woDomain.Actions {
		if a.Name == "by-code" {
			found = true
			if a.ToolName != "UteamupCodingsystemWorkorders" {
				t.Errorf("by-code action: expected tool UteamupCodingsystemWorkorders, got %q", a.ToolName)
			}
			break
		}
	}
	if !found {
		t.Error("expected by-code action on workorder domain")
	}
}
