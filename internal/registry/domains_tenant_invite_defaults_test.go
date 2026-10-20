package registry

import "testing"

// invite-defaults-set mirrors PATCH /api/tenant/{guid}/invite-defaults, which replaces every
// field. Each flag must map to the backend's camelCase body name or the setting silently resets.
func TestTenantInviteDefaultsSetFlagsMatchBackendBody(t *testing.T) {
	d := findDomain("tenant")
	if d == nil {
		t.Fatal("expected tenant domain to be registered")
	}
	set := findAction(d, "invite-defaults-set")
	if set == nil {
		t.Fatal("expected invite-defaults-set action on tenant")
	}

	flags := flagsToMap(set.Flags)
	want := map[string]struct{ body, kind string }{
		"auto-license": {"autoAssignLicenseOnInvite", "bool"},
		"license-type": {"defaultLicenseType", "int"},
		"auto-role":    {"autoAssignRoleOnInvite", "bool"},
		"role-id":      {"defaultInviteRoleId", "string"},
		"auto-radio":   {"autoAssignRadioLicenseOnInvite", "bool"},
	}
	for name, expected := range want {
		flag, ok := flags[name]
		if !ok {
			t.Errorf("--%s is missing", name)
			continue
		}
		if flag.BodyName != expected.body || flag.Type != expected.kind {
			t.Errorf("--%s maps to %s (%s), want %s (%s)", name, flag.BodyName, flag.Type, expected.body, expected.kind)
		}
	}
}
