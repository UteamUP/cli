package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPersonalLabourRoutesRequirePublicGuidsAndDoNotOfferWorkerSelectors(t *testing.T) {
	domain := findDomain("labour-marketplace-workspace")
	for name, route := range map[string]string{
		"accept-dispatch":     "dispatches/{dispatchGuid}/acceptance",
		"create-my-timesheet": "agreements/{agreementGuid}/my-timesheets",
		"my-timesheets":       "agreements/{agreementGuid}/my-timesheets",
	} {
		action := findAction(domain, name)
		if action == nil || action.RESTPath != route || action.MCPOnly || !action.RejectExtraArgs || len(action.Args) != 1 {
			t.Fatalf("invalid personal route %s: %+v", name, action)
		}
		if !action.Args[0].Required || action.Args[0].Type != "non-empty-uuid" {
			t.Fatalf("personal identity must be a required public GUID: %+v", action.Args)
		}
		for _, flag := range action.Flags {
			if flag.Name != "from-json" && flag.Name != "accept" && flag.Name != "reason" {
				t.Fatalf("unexpected personal actor or provider control: %+v", flag)
			}
		}
	}
}

func TestPersonalDispatchNeverInventsAnAcceptanceDecision(t *testing.T) {
	const dispatch = "11111111-1111-4111-8111-111111111111"
	if requests, err := runRegisteredDomainCommand(t, "labour-marketplace-workspace", "accept-dispatch", dispatch); err == nil || len(requests) != 0 {
		t.Fatal("missing explicit decision must fail before sending")
	}
	for _, decision := range []string{"true", "false"} {
		requests, err := runRegisteredDomainCommand(t, "labour-marketplace-workspace", "accept-dispatch",
			dispatch, "--accept="+decision, "--reason", "Worker reviewed the assignment")
		if err != nil || len(requests) != 1 {
			t.Fatalf("decision request failed: %v, %+v", err, requests)
		}
		r := requests[0]
		if r.method != "POST" || r.path != "/api/labour-marketplace/dispatches/"+dispatch+"/acceptance" || r.body["accept"] != (decision == "true") {
			t.Fatalf("explicit decision changed: %+v", r)
		}
	}
}

func TestPersonalTimesheetPreservesTheStableKeyAndEntriesInTheCanonicalRestBody(t *testing.T) {
	const agreement = "22222222-2222-4222-8222-222222222222"
	const key = "33333333-3333-4333-8333-333333333333"
	file := filepath.Join(t.TempDir(), "personal-time.json")
	body := `{"idempotencyKey":"33333333-3333-4333-8333-333333333333","periodStart":"2026-11-01T00:00:00Z","periodEnd":"2026-11-02T00:00:00Z","entries":[{"dispatchGuid":"11111111-1111-4111-8111-111111111111","timeType":"regular","startedAt":"2026-11-01T08:00:00Z","endedAt":"2026-11-01T09:00:00Z","notes":"Own time"}]}`
	if err := os.WriteFile(file, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		requests, err := runRegisteredDomainCommand(t, "labour-marketplace-workspace", "create-my-timesheet", agreement, "--from-json", file)
		if err != nil || len(requests) != 1 {
			t.Fatalf("personal time request failed: %v, %+v", err, requests)
		}
		r := requests[0]
		if r.method != "POST" || r.path != "/api/labour-marketplace/agreements/"+agreement+"/my-timesheets" || r.body["idempotencyKey"] != key {
			t.Fatalf("stable original request changed: %+v", r)
		}
		if _, wrapped := r.body["model"]; wrapped {
			t.Fatal("REST body must not be wrapped in model")
		}
		if len(r.body["entries"].([]any)) != 1 {
			t.Fatal("personal entry was changed")
		}
	}
	requests, err := runRegisteredDomainCommand(t, "labour-marketplace-workspace", "my-timesheets", agreement)
	if err != nil || len(requests) != 1 || requests[0].method != "GET" || len(requests[0].body) != 0 {
		t.Fatalf("personal read failed: %v, %+v", err, requests)
	}
}
