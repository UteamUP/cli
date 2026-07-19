package registry

import "testing"

func findReturnCaseDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "return-case" {
			return domain
		}
	}
	t.Fatal("return-case domain is not registered")
	return nil
}

func TestReturnCaseDomainMirrorsLifecycleTools(t *testing.T) {
	domain := findReturnCaseDomain(t)
	expected := map[string]string{
		"list":    "UteamupReturnCaseList",
		"get":     "UteamupReturnCaseGet",
		"create":  "UteamupReturnCaseCreate",
		"approve": "UteamupReturnCaseApprove",
		"reject":  "UteamupReturnCaseReject",
		"cancel":  "UteamupReturnCaseCancel",
		"ship":    "UteamupReturnCaseShip",
		"receive": "UteamupReturnCaseReceive",
		"inspect": "UteamupReturnCaseInspect",
		"credit":  "UteamupReturnCaseCredit",
		"replace": "UteamupReturnCaseReplace",
		"repair":  "UteamupReturnCaseRepair",
		"close":   "UteamupReturnCaseClose",
	}

	if domain.APIPath != "/api/returncases" {
		t.Fatalf("APIPath = %q, want /api/returncases", domain.APIPath)
	}
	for name, toolName := range expected {
		action := findReturnCaseAction(t, domain, name)
		if action.ToolName != toolName {
			t.Errorf("%s ToolName = %q, want %q", name, action.ToolName, toolName)
		}
	}
}

func TestReturnCaseWritesKeepOneRetryKeyInHeaderAndBody(t *testing.T) {
	domain := findReturnCaseDomain(t)
	for _, actionName := range []string{
		"create",
		"approve",
		"reject",
		"cancel",
		"ship",
		"receive",
		"inspect",
		"credit",
		"replace",
		"repair",
		"close",
	} {
		action := findReturnCaseAction(t, domain, actionName)
		var found bool
		for _, flag := range action.Flags {
			if flag.Name != "idempotency-key" {
				continue
			}
			found = flag.Required &&
				flag.HeaderName == "Idempotency-Key" &&
				flag.BodyName == "idempotencyKey" &&
				flag.MirrorHeaderInBody
		}
		if !found {
			t.Errorf("%s must mirror one required idempotency key", actionName)
		}
	}
}

func TestReturnCaseTransitionsRequireReviewedVersion(t *testing.T) {
	domain := findReturnCaseDomain(t)
	for _, action := range domain.Actions {
		if action.Name == "list" || action.Name == "get" || action.Name == "create" {
			continue
		}
		var found bool
		for _, flag := range action.Flags {
			if flag.Name == "expected-updated-at" && flag.Required {
				found = true
			}
		}
		if !found {
			t.Errorf("%s must require --expected-updated-at", action.Name)
		}
	}
}

func findReturnCaseAction(t *testing.T, domain *Domain, name string) Action {
	t.Helper()
	for _, action := range domain.Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("return-case action %q is not registered", name)
	return Action{}
}
