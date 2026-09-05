package registry

import "testing"

func TestShiftHandoverAcceptanceActionsUseGuidContracts(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")

	assertShiftHandoverAction(t, domain, "pending-acceptances", "GET", "acceptances/pending", false)
	assertShiftHandoverAction(t, domain, "submit", "PUT", "by-guid/{handoverGuid}/submit", true)
	assertShiftHandoverAction(t, domain, "start-review", "PUT", "by-guid/{handoverGuid}/start-review", true)
	assertShiftHandoverAction(t, domain, "accept", "PUT", "by-guid/{handoverGuid}/accept", true)
	assertShiftHandoverAction(t, domain, "complete", "PUT", "by-guid/{handoverGuid}/complete", true)
	assertShiftHandoverAction(t, domain, "reject", "PUT", "by-guid/{handoverGuid}/reject", true)
	assertShiftHandoverAction(t, domain, "update", "PUT", "by-guid/{handoverGuid}", true)
	assertShiftHandoverAction(
		t,
		domain,
		"decline-acceptance",
		"PUT",
		"by-guid/{handoverGuid}/decline-acceptance",
		true,
	)
}

func TestShiftHandoverRejectionRequiresRecordedReason(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	for _, action := range domain.Actions {
		if action.Name != "reject" {
			continue
		}
		for _, flag := range action.Flags {
			if flag.Name == "reason" && flag.Required && flag.Type == "string" {
				return
			}
		}
	}
	t.Fatal("reject must require a recorded reason")
}

func TestShiftHandoverUpdateAssignsOnlyByPublicWorkerGuid(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	count := 0
	for _, action := range domain.Actions {
		if action.Name == "update" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected one canonical update action, found %d", count)
	}
	for _, action := range domain.Actions {
		if action.Name != "update" {
			continue
		}
		for _, flag := range action.Flags {
			if flag.Name == "incoming-operator-id" || flag.Name == "incoming-operator-name" {
				t.Fatal("draft updates must not trust storage IDs or supplied operator names")
			}
		}
		for _, flag := range action.Flags {
			if flag.Name == "incoming-operator-guid" && flag.BodyName == "incomingOperatorGuid" && flag.Type == "uuid" {
				return
			}
		}
	}
	t.Fatal("draft update must expose the incoming operator GUID")
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

func TestShiftHandoverArchiveUsesGuidVersionQueryAndRetryHeader(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	for _, action := range domain.Actions {
		if action.Name != "archive" {
			continue
		}
		if action.HTTPMethod != "DELETE" || action.RESTPath != "by-guid/{handoverGuid}" || len(action.Args) != 1 || action.Args[0].Type != "uuid" {
			t.Fatalf("unexpected archive contract: %#v", action)
		}
		version, receipt := false, false
		for _, flag := range action.Flags {
			if flag.Name == "concurrency-token" {
				version = flag.Required && flag.QueryName == "concurrencyToken"
			}
			if flag.Name == "idempotency-key" {
				receipt = flag.Required && flag.HeaderName == "Idempotency-Key"
			}
		}
		if !version || !receipt {
			t.Fatalf("archive must require version and retry evidence: %#v", action.Flags)
		}
		return
	}
	t.Fatal("archive action missing")
}

func assertShiftHandoverAction(
	t *testing.T,
	domain *Domain,
	name string,
	httpMethod string,
	restPath string,
	requiresGUID bool,
) {
	t.Helper()
	for _, action := range domain.Actions {
		if action.Name != name {
			continue
		}
		if action.HTTPMethod != httpMethod || action.RESTPath != restPath {
			t.Fatalf("unexpected %s contract: %#v", name, action)
		}
		if !requiresGUID {
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

func TestShiftHandoverCoreActionsMatchTheCanonicalAPI(t *testing.T) {
	domain := findRegisteredDomain(t, "shift-handover")
	expected := map[string]string{"get": "/api/shifthandover/by-guid/186d0e09-c70b-45ba-8763-a415b1420568", "list": "/api/shifthandover/search", "create": "/api/shifthandover"}
	counts := map[string]int{}
	for _, action := range domain.Actions {
		counts[action.Name]++
		if action.Name == "delete" {
			t.Fatal("legacy unversioned delete must not coexist with the archive command")
		}
		if path, ok := expected[action.Name]; ok {
			actual, _ := buildRESTPath(domain, action, map[string]any{"handoverGuid": "186d0e09-c70b-45ba-8763-a415b1420568"})
			if actual != path {
				t.Fatalf("%s routes to %s, expected %s", action.Name, actual, path)
			}
			if action.Name == "get" {
				if action.HTTPMethod != "GET" || len(action.Args) != 1 || action.Args[0].Type != "uuid" {
					t.Fatalf("get must require the public GUID: %#v", action)
				}
			} else if action.HTTPMethod != "POST" {
				t.Fatalf("%s must post to the existing API", action.Name)
			}
		}
	}
	for name := range expected {
		if counts[name] != 1 {
			t.Fatalf("expected one %s action", name)
		}
	}
	if counts["archive"] != 1 {
		t.Fatal("retain one versioned archive action")
	}
}
