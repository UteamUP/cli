package registry

import "testing"

func findWorkingTimeDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "workingtime" {
			return d
		}
	}
	t.Fatal("expected workingtime domain to be registered")
	return nil
}

func TestWorkingTimeDomainRegistered(t *testing.T) {
	d := findWorkingTimeDomain(t)
	if d.APIPath != "/api/workingtime" {
		t.Errorf("workingtime APIPath = %q, want %q", d.APIPath, "/api/workingtime")
	}
	wantAliases := map[string]bool{"working-time": true, "wt": true}
	for _, a := range d.Aliases {
		delete(wantAliases, a)
	}
	if len(wantAliases) != 0 {
		t.Errorf("workingtime missing aliases %v (got %v)", wantAliases, d.Aliases)
	}
}

func TestWorkingTimeRuleSetActionsWired(t *testing.T) {
	d := findWorkingTimeDomain(t)
	byName := map[string]*Action{}
	for i := range d.Actions {
		byName[d.Actions[i].Name] = &d.Actions[i]
	}

	list, ok := byName["ruleset-list"]
	if !ok || list.HTTPMethod != "GET" || list.RESTPath != "rulesets" {
		t.Errorf("ruleset-list must be GET \"rulesets\", got %+v", list)
	}

	create, ok := byName["ruleset-create"]
	if !ok || create.HTTPMethod != "POST" || create.RESTPath != "rulesets" {
		t.Fatalf("ruleset-create must be POST \"rulesets\", got %+v", create)
	}
	var nameFlag, countryFlag *FlagDef
	for i := range create.Flags {
		switch create.Flags[i].Name {
		case "name":
			nameFlag = &create.Flags[i]
		case "country":
			countryFlag = &create.Flags[i]
		}
	}
	if nameFlag == nil || !nameFlag.Required {
		t.Errorf("ruleset-create needs a required 'name' flag, got %+v", create.Flags)
	}
	if countryFlag == nil || countryFlag.BodyName != "countryCode" {
		t.Errorf("ruleset-create 'country' must map to countryCode, got %+v", countryFlag)
	}
}

func TestWorkingTimeProjectOvertimeActionWired(t *testing.T) {
	d := findWorkingTimeDomain(t)
	var po *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "project-overtime" {
			po = &d.Actions[i]
		}
	}
	if po == nil {
		t.Fatal("expected 'project-overtime' action")
	}
	if po.HTTPMethod != "POST" || po.RESTPath != "project-overtime" {
		t.Errorf("project-overtime = %s %q, want POST \"project-overtime\"", po.HTTPMethod, po.RESTPath)
	}
	byFlag := map[string]*FlagDef{}
	for i := range po.Flags {
		byFlag[po.Flags[i].Name] = &po.Flags[i]
	}
	// Float flag defaults MUST be float literals (an untyped int default panics the registry).
	for _, name := range []string{"worked", "weekly-limit"} {
		f, ok := byFlag[name]
		if !ok {
			t.Fatalf("project-overtime missing %q flag", name)
		}
		if _, isFloat := f.Default.(float64); !isFloat {
			t.Errorf("float flag %q default must be a float literal, got %T (%v)", name, f.Default, f.Default)
		}
	}
	if r, ok := byFlag["rostered"]; !ok || r.BodyName != "rosteredHours" || !r.Required {
		t.Errorf("project-overtime 'rostered' must be a required float → rosteredHours, got %+v", r)
	}
}

func TestWorkingTimeHolidaysActionWired(t *testing.T) {
	d := findWorkingTimeDomain(t)
	var h *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "holidays" {
			h = &d.Actions[i]
		}
	}
	if h == nil {
		t.Fatal("expected 'holidays' action")
	}
	if h.HTTPMethod != "GET" || h.RESTPath != "holidays/{year}" {
		t.Errorf("holidays = %s %q, want GET \"holidays/{year}\"", h.HTTPMethod, h.RESTPath)
	}
	if len(h.Args) != 1 || h.Args[0].Name != "year" || !h.Args[0].Required || h.Args[0].Type != "int" {
		t.Errorf("holidays must take a required int 'year' arg, got %+v", h.Args)
	}
}
