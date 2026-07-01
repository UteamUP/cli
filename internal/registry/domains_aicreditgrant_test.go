package registry

import "testing"

func findAiCreditGrantDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "aicreditgrant" {
			return d
		}
	}
	t.Fatal("expected aicreditgrant domain to be registered")
	return nil
}

func findAiCreditGrantAction(t *testing.T, name string) *Action {
	t.Helper()
	d := findAiCreditGrantDomain(t)
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected %q action on aicreditgrant domain", name)
	return nil
}

func TestAiCreditGrantDomainRegistered(t *testing.T) {
	d := findAiCreditGrantDomain(t)
	if d.Description == "" {
		t.Error("aicreditgrant domain must have a Description")
	}
	if d.APIPath != "/api/aicreditgrant" {
		t.Errorf("aicreditgrant APIPath = %q, want %q", d.APIPath, "/api/aicreditgrant")
	}
	wantAliases := map[string]bool{"aicreditgrants": true, "aicredit": true, "aicredits": true}
	for _, a := range d.Aliases {
		delete(wantAliases, a)
	}
	if len(wantAliases) != 0 {
		t.Errorf("aicreditgrant domain missing aliases: %v (got %v)", wantAliases, d.Aliases)
	}
}

func TestAiCreditGrantActionsWired(t *testing.T) {
	expected := map[string]string{
		"issue":  "UteamupAiCreditGrantIssue",
		"mine":   "UteamupAiCreditGrantMine",
		"claim":  "UteamupAiCreditGrantClaim",
		"revoke": "UteamupAiCreditGrantRevoke",
	}
	d := findAiCreditGrantDomain(t)
	if len(d.Actions) != len(expected) {
		t.Errorf("aicreditgrant has %d actions, want %d", len(d.Actions), len(expected))
	}
	for name, tool := range expected {
		a := findAiCreditGrantAction(t, name)
		if a.ToolName != tool {
			t.Errorf("action %q ToolName = %q, want %q", name, a.ToolName, tool)
		}
	}
}

func TestAiCreditGrantIssueWiring(t *testing.T) {
	issue := findAiCreditGrantAction(t, "issue")
	if issue.HTTPMethod != "POST" {
		t.Errorf("issue HTTPMethod = %q, want POST", issue.HTTPMethod)
	}
	if issue.RESTPath != "" {
		t.Errorf("issue RESTPath = %q, want empty (posts to base path)", issue.RESTPath)
	}

	// tenant-guid and amount must be required flags (camelCase to tenantGuid / amount in the body).
	wantRequired := map[string]bool{"tenant-guid": false, "amount": false}
	wantType := map[string]string{"tenant-guid": "string", "amount": "int"}
	for _, f := range issue.Flags {
		if _, tracked := wantRequired[f.Name]; !tracked {
			continue
		}
		if !f.Required {
			t.Errorf("issue --%s must be required", f.Name)
		}
		if f.Type != wantType[f.Name] {
			t.Errorf("issue --%s Type = %q, want %q", f.Name, f.Type, wantType[f.Name])
		}
		wantRequired[f.Name] = true
	}
	for name, found := range wantRequired {
		if !found {
			t.Errorf("issue action must declare a --%s flag", name)
		}
	}

	// requires-claim is the claim-gate toggle (bool).
	var foundBool bool
	for _, f := range issue.Flags {
		if f.Name == "requires-claim" {
			foundBool = true
			if f.Type != "bool" {
				t.Errorf("issue --requires-claim Type = %q, want bool", f.Type)
			}
		}
	}
	if !foundBool {
		t.Error("issue action must declare a --requires-claim bool flag")
	}
}

func TestAiCreditGrantMineWiring(t *testing.T) {
	mine := findAiCreditGrantAction(t, "mine")
	if mine.RESTPath != "mine" {
		t.Errorf("mine RESTPath = %q, want %q", mine.RESTPath, "mine")
	}
	if len(mine.Args) != 0 {
		t.Errorf("mine must take no positional args, got %+v", mine.Args)
	}
}

func TestAiCreditGrantGuidRoutesAreGuidFirst(t *testing.T) {
	cases := map[string]string{
		"claim":  "{guid}/claim",
		"revoke": "{guid}/revoke",
	}
	for name, wantPath := range cases {
		a := findAiCreditGrantAction(t, name)
		if a.HTTPMethod != "POST" {
			t.Errorf("%s HTTPMethod = %q, want POST", name, a.HTTPMethod)
		}
		if a.RESTPath != wantPath {
			t.Errorf("%s RESTPath = %q, want %q", name, a.RESTPath, wantPath)
		}
		if len(a.Args) != 1 || a.Args[0].Name != "guid" || a.Args[0].Type != "string" || !a.Args[0].Required {
			t.Errorf("%s must take a single required string `guid` arg, got %+v", name, a.Args)
		}
	}
}
