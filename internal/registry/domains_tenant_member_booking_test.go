package registry

import "testing"

func TestTenantMemberBookingActionsMirrorBackendToolsAndRoutes(t *testing.T) {
	domain := findDomain("member-booking")
	if domain == nil {
		t.Fatal("member-booking domain is not registered")
	}
	if domain.APIPath != "/api/tenant" {
		t.Fatalf("APIPath = %q, want /api/tenant", domain.APIPath)
	}
	expected := map[string]struct {
		tool   string
		method string
		path   string
	}{
		"settings-set":       {"UteamupTenantMemberBookingSettingsUpdate", "PUT", "users/{userGuid}/booking-settings"},
		"my-preferences-get": {"UteamupMyBookingPreferencesGet", "GET", "me/booking-preferences"},
		"my-preferences-set": {"UteamupMyBookingPreferencesUpdate", "PUT", "me/booking-preferences"},
	}
	for actionName, want := range expected {
		action := findAction(domain, actionName)
		if action == nil {
			t.Fatalf("action %q is not registered", actionName)
		}
		if action.ToolName != want.tool || action.HTTPMethod != want.method || action.RESTPath != want.path || action.MCPOnly {
			t.Errorf("%s = tool %q method %q path %q mcpOnly %v, want REST %q %q %q",
				actionName, action.ToolName, action.HTTPMethod, action.RESTPath, action.MCPOnly, want.tool, want.method, want.path)
		}
	}

	settings := findAction(domain, "settings-set")
	if len(settings.Args) != 1 || settings.Args[0].Name != "userGuid" || settings.Args[0].Type != "uuid" || !settings.Args[0].Required {
		t.Fatalf("settings-set must take one required user GUID: %+v", settings.Args)
	}
	if len(findAction(domain, "my-preferences-get").Args) != 0 {
		t.Fatal("my-preferences-get reads the caller and takes no arguments")
	}
}

func TestTenantMemberBookingSettingsLeaveOmittedValuesUnchanged(t *testing.T) {
	settings := findAction(findDomain("member-booking"), "settings-set")
	expected := map[string]string{"is-bookable-resource": "isBookableResource", "allow-public-booking": "allowPublicBooking"}
	for _, flag := range settings.Flags {
		body, ok := expected[flag.Name]
		if !ok {
			t.Errorf("unexpected flag %+v (the admin cannot set the member's own opt-out)", flag)
			continue
		}
		if flag.BodyName != body || flag.Type != "bool" || flag.Required || flag.Default != nil {
			t.Errorf("%s must be an optional bool with no default mapped to %s: %+v", flag.Name, body, flag)
		}
		delete(expected, flag.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing booking settings flags: %v", expected)
	}

	preferences := findAction(findDomain("member-booking"), "my-preferences-set")
	if len(preferences.Flags) != 1 {
		t.Fatalf("my-preferences-set flags = %+v, want only --public-booking-opt-out", preferences.Flags)
	}
	optOut := preferences.Flags[0]
	if optOut.Name != "public-booking-opt-out" || optOut.BodyName != "publicBookingOptOut" || optOut.Type != "bool" || !optOut.Required {
		t.Fatalf("public-booking-opt-out must be a required bool: %+v", optOut)
	}
}
