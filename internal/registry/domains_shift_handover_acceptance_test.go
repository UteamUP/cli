package registry

import "testing"

func TestShiftHandoverAcceptanceActionsUseGuidContracts(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")

	assertShiftHandoverAction(t, domain, "pending-acceptances", "GET", "acceptances/pending", false)
	assertShiftHandoverAction(t, domain, "submit", "PUT", "by-guid/{handoverGuid}/submit", true)
	assertShiftHandoverAction(t, domain, "start-review", "PUT", "by-guid/{handoverGuid}/start-review", true)
	assertShiftHandoverAction(t, domain, "accept", "PUT", "by-guid/{handoverGuid}/accept", true)
	assertShiftHandoverAction(t, domain, "complete", "PUT", "by-guid/{handoverGuid}/complete", true)
	assertShiftHandoverAction(
		t,
		domain,
		"decline-acceptance",
		"PUT",
		"by-guid/{handoverGuid}/decline-acceptance",
		true,
	)
}

func findRegisteredDomain(t *testing.T, name string) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == name {
			return domain
		}
	}
	t.Fatalf("%s domain not registered", name)
	return nil
}

func assertShiftHandoverAction(
	t *testing.T,
	domain *Domain,
	name string,
	httpMethod string,
	restPath string,
	requiresGuid bool,
) {
	t.Helper()
	for _, action := range domain.Actions {
		if action.Name != name {
			continue
		}
		if action.HTTPMethod != httpMethod || action.RESTPath != restPath {
			t.Fatalf("unexpected %s contract: %#v", name, action)
		}
		if !requiresGuid {
			return
		}
		if len(action.Args) != 1 || action.Args[0].Name != "handoverGuid" || action.Args[0].Type != "uuid" {
			t.Fatalf("%s must use one UUID handoverGuid argument: %#v", name, action.Args)
		}
		assertHandoverMutationFlags(t, name, action.Flags)
		return
	}
	t.Fatalf("%s action missing", name)
}

func assertHandoverMutationFlags(t *testing.T, actionName string, flags []FlagDef) {
	t.Helper()
	foundConcurrency := false
	foundIdempotency := false
	for _, flag := range flags {
		switch flag.Name {
		case "concurrency-token":
			foundConcurrency = flag.Required && flag.BodyName == "concurrencyToken"
		case "idempotency-key":
			foundIdempotency = flag.Required && flag.HeaderName == "Idempotency-Key"
		}
	}
	if !foundConcurrency || !foundIdempotency {
		t.Fatalf("%s must require concurrency body and idempotency header flags: %#v", actionName, flags)
	}
}
