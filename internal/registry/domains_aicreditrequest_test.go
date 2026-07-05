package registry

import "testing"

func findAiCreditRequestDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "aicreditrequest" {
			return d
		}
	}
	t.Fatal("expected aicreditrequest domain to be registered")
	return nil
}

func findAiCreditRequestAction(t *testing.T, name string) *Action {
	t.Helper()
	d := findAiCreditRequestDomain(t)
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected %q action on aicreditrequest domain", name)
	return nil
}

func TestAiCreditRequestDomainRegistered(t *testing.T) {
	d := findAiCreditRequestDomain(t)
	if d.Description == "" {
		t.Error("aicreditrequest domain must have a Description")
	}
	if d.APIPath != "/api/aicreditrequest" {
		t.Errorf("aicreditrequest APIPath = %q, want %q", d.APIPath, "/api/aicreditrequest")
	}
}

func TestAiCreditRequestActionsWired(t *testing.T) {
	expected := map[string]string{
		"submit":  "UteamupAiCreditRequestSubmit",
		"mine":    "UteamupAiCreditRequestMine",
		"pending": "UteamupAiCreditRequestPending",
		"fulfill": "UteamupAiCreditRequestFulfill",
		"reject":  "UteamupAiCreditRequestReject",
	}
	d := findAiCreditRequestDomain(t)
	if len(d.Actions) != len(expected) {
		t.Errorf("aicreditrequest has %d actions, want %d", len(d.Actions), len(expected))
	}
	for name, tool := range expected {
		a := findAiCreditRequestAction(t, name)
		if a.ToolName != tool {
			t.Errorf("action %q ToolName = %q, want %q", name, a.ToolName, tool)
		}
	}
}

func TestAiCreditRequestSubmitWiring(t *testing.T) {
	submit := findAiCreditRequestAction(t, "submit")
	if submit.HTTPMethod != "POST" {
		t.Errorf("submit HTTPMethod = %q, want POST", submit.HTTPMethod)
	}
	if submit.RESTPath != "" {
		t.Errorf("submit RESTPath = %q, want empty (posts to base path)", submit.RESTPath)
	}
	var foundAmount bool
	for _, f := range submit.Flags {
		if f.Name == "requested-monthly-credits" {
			foundAmount = true
			if !f.Required {
				t.Error("submit --requested-monthly-credits must be required")
			}
			if f.Type != "int" {
				t.Errorf("submit --requested-monthly-credits Type = %q, want int", f.Type)
			}
			if f.BodyName != "requestedMonthlyCredits" {
				t.Errorf("submit --requested-monthly-credits BodyName = %q, want requestedMonthlyCredits", f.BodyName)
			}
		}
	}
	if !foundAmount {
		t.Error("submit action must declare a --requested-monthly-credits flag")
	}
}

func TestAiCreditRequestGuidRoutesAreGuidFirst(t *testing.T) {
	cases := map[string]string{
		"fulfill": "{guid}/fulfill",
		"reject":  "{guid}/reject",
	}
	for name, wantPath := range cases {
		a := findAiCreditRequestAction(t, name)
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

	// fulfill requires the package-plan-guid body flag (GUID-keyed at the boundary).
	fulfill := findAiCreditRequestAction(t, "fulfill")
	var foundPlan bool
	for _, f := range fulfill.Flags {
		if f.Name == "package-plan-guid" {
			foundPlan = true
			if !f.Required {
				t.Error("fulfill --package-plan-guid must be required")
			}
			if f.BodyName != "packagePlanGuid" {
				t.Errorf("fulfill --package-plan-guid BodyName = %q, want packagePlanGuid", f.BodyName)
			}
		}
	}
	if !foundPlan {
		t.Error("fulfill action must declare a --package-plan-guid flag")
	}
}

func TestAiCreditRequestPendingWiring(t *testing.T) {
	pending := findAiCreditRequestAction(t, "pending")
	if pending.RESTPath != "pending" {
		t.Errorf("pending RESTPath = %q, want %q", pending.RESTPath, "pending")
	}
	if len(pending.Args) != 0 {
		t.Errorf("pending must take no positional args, got %+v", pending.Args)
	}
}
