package registry

import "testing"

func TestTenantLifecycleActionsMirrorBackendRoutes(t *testing.T) {
	domain := findDomain("tenant-lifecycle")
	if domain == nil {
		t.Fatal("tenant-lifecycle domain is not registered")
	}
	if domain.APIPath != "/api/billing/lifecycle" {
		t.Fatalf("APIPath = %q, want /api/billing/lifecycle", domain.APIPath)
	}
	expected := map[string]struct {
		tool   string
		method string
		base   string
		path   string
	}{
		"restore":           {"TenantLifecycleRestore", "POST", "", "restore"},
		"admin-get":         {"AdminGetTenantLifecycle", "GET", "/api/globaladmin/tenants", "{tenantGuid}/lifecycle"},
		"unarchive":         {"AdminUnarchiveTenant", "POST", "/api/globaladmin/tenants", "{tenantGuid}/lifecycle/unarchive"},
		"postpone-deletion": {"AdminPostponeTenantDeletion", "POST", "/api/globaladmin/tenants", "{tenantGuid}/lifecycle/postpone-deletion"},
		"legal-hold":        {"AdminSetTenantLegalHold", "PUT", "/api/globaladmin/tenants", "{tenantGuid}/lifecycle/legal-hold"},
	}
	for actionName, want := range expected {
		action := findAction(domain, actionName)
		if action == nil {
			t.Fatalf("action %q is not registered", actionName)
		}
		if action.ToolName != want.tool || action.HTTPMethod != want.method || action.RESTBasePath != want.base || action.RESTPath != want.path {
			t.Errorf("%s = tool %q method %q base %q path %q, want %q %q %q %q",
				actionName, action.ToolName, action.HTTPMethod, action.RESTBasePath, action.RESTPath,
				want.tool, want.method, want.base, want.path)
		}
	}

	get := findAction(domain, "get")
	if get == nil || !get.UseDomainBasePath || get.HTTPMethod != "GET" || len(get.Args) != 0 {
		t.Fatalf("get must read the current tenant at the domain root without arguments: %+v", get)
	}
}

func TestTenantLifecycleAdminActionsRequireATenantGuid(t *testing.T) {
	domain := findDomain("tenant-lifecycle")
	for _, name := range []string{"admin-get", "unarchive", "postpone-deletion", "legal-hold"} {
		action := findAction(domain, name)
		if len(action.Args) != 1 || action.Args[0].Name != "tenantGuid" || action.Args[0].Type != "non-empty-uuid" || !action.Args[0].Required {
			t.Errorf("%s must take one required non-empty tenant GUID: %+v", name, action.Args)
		}
	}
}

func TestTenantLifecycleFlagsSendTheBackendBodyNames(t *testing.T) {
	domain := findDomain("tenant-lifecycle")
	expected := map[string]map[string]string{
		"unarchive":         {"extend-days": "extendDays"},
		"postpone-deletion": {"deletion-date": "deletionDueAt"},
	}
	for actionName, flags := range expected {
		action := findAction(domain, actionName)
		for flagName, body := range flags {
			found := false
			for _, flag := range action.Flags {
				if flag.Name == flagName {
					found = true
					if flag.BodyName != body {
						t.Errorf("%s --%s body = %q, want %q", actionName, flagName, flag.BodyName, body)
					}
				}
			}
			if !found {
				t.Errorf("%s has no --%s flag", actionName, flagName)
			}
		}
	}
	hold := findAction(domain, "legal-hold")
	if len(hold.Flags) != 2 || hold.Flags[0].Name != "hold" || hold.Flags[0].Type != "bool" || !hold.Flags[0].Required {
		t.Fatalf("legal-hold needs a required bool --hold and an optional --reason: %+v", hold.Flags)
	}
}
