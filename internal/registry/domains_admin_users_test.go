package registry

import "testing"

func TestAdminUsersDomainRegistered(t *testing.T) {
	d := findAdminUsersDomain("admin-users")
	if d == nil {
		t.Fatal("expected admin-users domain to be registered")
	}
	if d.Description == "" {
		t.Error("admin-users domain must have a Description")
	}
}

func TestAdminUsersActionsWired(t *testing.T) {
	d := findAdminUsersDomain("admin-users")
	if d == nil {
		t.Fatal("expected admin-users domain to be registered")
	}
	expected := map[string]string{
		"list":           "UteamupAdminUserList",
		"get":            "UteamupAdminUserGet",
		"login-events":   "UteamupAdminUserLoginEvents",
		"disable":        "UteamupAdminUserDisable",
		"enable":         "UteamupAdminUserEnable",
		"reset-password": "UteamupAdminUserResetPassword",
	}
	got := map[string]string{}
	for _, a := range d.Actions {
		got[a.Name] = a.ToolName
	}
	for action, toolName := range expected {
		if got[action] != toolName {
			t.Errorf("expected admin-users action %q to map to %q, got %q", action, toolName, got[action])
		}
	}
}

func TestAdminUsersResetPasswordRequiresMode(t *testing.T) {
	d := findAdminUsersDomain("admin-users")
	if d == nil {
		t.Fatal("expected admin-users domain to be registered")
	}
	var rp *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "reset-password" {
			rp = &d.Actions[i]
			break
		}
	}
	if rp == nil {
		t.Fatal("expected admin-users reset-password action")
	}
	foundMode := false
	foundConfirmEmail := false
	for _, f := range rp.Flags {
		if f.Name == "mode" && f.Required {
			foundMode = true
		}
		if f.Name == "confirm-email" && f.Required {
			foundConfirmEmail = true
		}
	}
	if !foundMode {
		t.Error("reset-password must require --mode")
	}
	if !foundConfirmEmail {
		t.Error("reset-password must require --confirm-email")
	}
}

func findAdminUsersDomain(_ string) *Domain {
	const name = "admin-users"
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == name {
			return d
		}
	}
	return nil
}
