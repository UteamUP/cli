package registry

import "testing"

func findHandoverAttestationDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "handoverattestation" {
			return d
		}
	}
	t.Fatal("expected handoverattestation domain to be registered")
	return nil
}

func TestHandoverAttestationDomainRegistered(t *testing.T) {
	d := findHandoverAttestationDomain(t)
	if d.APIPath != "/api/handoverattestation" {
		t.Errorf("handoverattestation APIPath = %q, want %q", d.APIPath, "/api/handoverattestation")
	}
}

func TestHandoverAttestationActionsWired(t *testing.T) {
	d := findHandoverAttestationDomain(t)
	byName := map[string]*Action{}
	for i := range d.Actions {
		byName[d.Actions[i].Name] = &d.Actions[i]
	}

	issue, ok := byName["issue"]
	if !ok || issue.HTTPMethod != "POST" || issue.RESTPath != "{handover-guid}/issue" {
		t.Fatalf("issue must be POST \"{handover-guid}/issue\", got %+v", issue)
	}
	if len(issue.Args) != 1 || issue.Args[0].Name != "handover-guid" || issue.Args[0].Type != "uuid" {
		t.Errorf("issue must take a uuid 'handover-guid' arg, got %+v", issue.Args)
	}

	verify, ok := byName["verify"]
	if !ok || verify.HTTPMethod != "POST" || verify.RESTPath != "verify" {
		t.Fatalf("verify must be POST \"verify\", got %+v", verify)
	}
	var tokenFlag *FlagDef
	for i := range verify.Flags {
		if verify.Flags[i].Name == "token" {
			tokenFlag = &verify.Flags[i]
		}
	}
	if tokenFlag == nil || !tokenFlag.Required {
		t.Errorf("verify must have a required 'token' flag, got %+v", verify.Flags)
	}
}
