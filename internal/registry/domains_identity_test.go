package registry

import "testing"

func TestIdentityDomainReadsOnlyTheSignedInUsersStatus(t *testing.T) {
	status := findDomainAction(t, "identity", "status")
	if status.HTTPMethod != "GET" || status.RESTPath != "status" {
		t.Fatalf("identity status route = method %q path %q, want GET status", status.HTTPMethod, status.RESTPath)
	}
	if status.ToolName != "UteamupIdentityVerificationStatus" {
		t.Fatalf("identity status tool = %q", status.ToolName)
	}
	if len(status.Flags) != 0 || len(status.Args) != 0 {
		t.Fatalf("identity status must take no user or identifier input: flags %#v args %#v", status.Flags, status.Args)
	}

	domain := findDomain("identity")
	if domain == nil || domain.APIPath != "/api/auth/eid" {
		t.Fatalf("identity domain path = %#v", domain)
	}
	for _, action := range domain.Actions {
		if action.HTTPMethod != "GET" {
			t.Fatalf("identity action %q is %s: the CLI must never start or change an electronic-ID flow", action.Name, action.HTTPMethod)
		}
	}
}
