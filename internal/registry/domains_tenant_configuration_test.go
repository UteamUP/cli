package registry

import "testing"

func TestTenantConfigurationActionsMirrorMCPTools(t *testing.T) {
	t.Parallel()

	var domain *Domain
	tenantDomainCount := 0
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "tenant" {
			tenantDomainCount++
			domain = candidate
		}
	}
	if domain == nil {
		t.Fatal("tenant domain is not registered")
	}
	if tenantDomainCount != 1 {
		t.Fatalf("tenant domain registrations = %d, want exactly 1", tenantDomainCount)
	}

	wantTools := map[string]string{
		"industry-profiles-get":     "UteamupTenantIndustryProfilesGet",
		"additional-industries-set": "UteamupTenantAdditionalIndustryProfilesSet",
		"modules-get":               "UteamupTenantModulesGet",
		"module-set":                "UteamupTenantModuleSet",
	}
	for _, action := range domain.Actions {
		toolName, wanted := wantTools[action.Name]
		if !wanted {
			continue
		}
		if action.ToolName != toolName {
			t.Errorf("action %q tool = %q, want %q", action.Name, action.ToolName, toolName)
		}
		if !action.MCPOnly {
			t.Errorf("action %q must use the MCP transport", action.Name)
		}
		delete(wantTools, action.Name)
	}
	if len(wantTools) != 0 {
		t.Fatalf("missing tenant configuration actions: %v", wantTools)
	}
}

func TestTenantConfigurationMutationArguments(t *testing.T) {
	t.Parallel()

	var additionalIndustries *Action
	var moduleSet *Action
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name != "tenant" {
			continue
		}
		for index := range domain.Actions {
			switch domain.Actions[index].Name {
			case "additional-industries-set":
				additionalIndustries = &domain.Actions[index]
			case "module-set":
				moduleSet = &domain.Actions[index]
			}
		}
	}
	if additionalIndustries == nil || moduleSet == nil {
		t.Fatal("tenant configuration mutation actions are missing")
	}

	if len(additionalIndustries.Flags) != 1 {
		t.Fatalf("additional industry flags = %d, want 1", len(additionalIndustries.Flags))
	}
	industryFlag := additionalIndustries.Flags[0]
	if industryFlag.Type != "stringSlice" || industryFlag.BodyName != "additionalIndustryProfileGuids" {
		t.Fatalf("additional industry flag = %+v", industryFlag)
	}
	defaultGuids, ok := industryFlag.Default.([]string)
	if !ok || len(defaultGuids) != 0 {
		t.Fatalf("additional industry default = %#v, want an explicit empty GUID list for clear", industryFlag.Default)
	}

	if len(moduleSet.Args) != 1 || moduleSet.Args[0].Name != "familyKey" {
		t.Fatalf("module family argument = %+v", moduleSet.Args)
	}
	if len(moduleSet.Flags) != 1 {
		t.Fatalf("module flags = %d, want 1", len(moduleSet.Flags))
	}
	enabledFlag := moduleSet.Flags[0]
	if enabledFlag.Type != "bool" || enabledFlag.BodyName != "isEnabled" || !enabledFlag.Required {
		t.Fatalf("module enabled flag = %+v", enabledFlag)
	}
}
