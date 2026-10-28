package registry

import "testing"

func TestAssetHealthSummaryUsesTenantScopedAggregateWithoutAssetArguments(t *testing.T) {
	requests, err := runRegisteredDomainCommand(t, "route", "asset-health-summary")
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d", len(requests))
	}
	request := requests[0]
	if request.method != "GET" || request.path != "/api/v1/inspectionanalytics/summary" || len(request.body) != 0 {
		t.Fatalf("unexpected aggregate request: %#v", request)
	}
}
