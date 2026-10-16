package registry

import "testing"

func TestMcpConnectionsDomainRoutesToTheGrantsEndpoints(t *testing.T) {
	list := findDomainAction(t, "mcpconnections", "list")
	if list.RESTPath != "grants/tenant" || list.HTTPMethod != "" {
		t.Fatalf("list route = method %q path %q, want derived GET grants/tenant", list.HTTPMethod, list.RESTPath)
	}

	mine := findDomainAction(t, "mcpconnections", "mine")
	if mine.RESTPath != "grants/mine" || mine.HTTPMethod != "GET" {
		t.Fatalf("mine route = method %q path %q", mine.HTTPMethod, mine.RESTPath)
	}

	revoke := findDomainAction(t, "mcpconnections", "revoke")
	if revoke.RESTPath != "grants/{grantGuid}" || revoke.HTTPMethod != "DELETE" {
		t.Fatalf("revoke route = method %q path %q", revoke.HTTPMethod, revoke.RESTPath)
	}
	if len(revoke.Args) != 1 || !revoke.Args[0].Required || revoke.Args[0].BodyName != "grantGuid" {
		t.Fatalf("revoke must take exactly one required grant Guid arg, got %#v", revoke.Args)
	}
}

func TestMcpConnectionsDomainTakesNoTenantOrUserArguments(t *testing.T) {
	for _, name := range []string{"list", "mine", "revoke"} {
		action := findDomainAction(t, "mcpconnections", name)
		for _, flag := range action.Flags {
			t.Fatalf("%s must take no flags (tenant and user come from the session), got %q", name, flag.Name)
		}
	}
}
