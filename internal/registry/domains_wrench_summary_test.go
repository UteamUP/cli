package registry

import "testing"

func TestWrenchSummaryUsesOneTenantRouteAndPreservesCalendarFilters(t *testing.T) {
	requests, err := runRegisteredDomainCommand(t, "time-entry", "wrench-summary", "--date-from", "2026-10-01", "--date-to", "2026-10-28", "--team", "24f69ca1-e664-421e-ac0d-47ee64d7d725", "--page", "2", "--page-size", "10")
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d", len(requests))
	}
	request := requests[0]
	if request.method != "GET" || request.path != "/api/timeentry/wrench-time/summary" || len(request.body) != 0 {
		t.Fatalf("unexpected request: %#v", request)
	}
	for key, want := range map[string]string{"dateFrom": "2026-10-01", "dateTo": "2026-10-28", "teamGuid": "24f69ca1-e664-421e-ac0d-47ee64d7d725", "page": "2", "pageSize": "10"} {
		if got := request.query[key]; len(got) != 1 || got[0] != want {
			t.Errorf("%s = %v, want %q", key, got, want)
		}
	}
}

func TestWrenchSummaryRequiresBothCalendarDates(t *testing.T) {
	requests, err := runRegisteredDomainCommand(t, "time-entry", "wrench-summary", "--date-from", "2026-10-01")
	if err == nil || len(requests) != 0 {
		t.Fatalf("missing date must fail before a request: %v, %#v", err, requests)
	}
}
