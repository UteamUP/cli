package registry

import (
	"net/http"
	"testing"
)

func TestStockLocationArchiveAndRestorePostToTheGuidRoutes(t *testing.T) {
	const guid = "33333333-3333-4333-8333-333333333333"
	for _, action := range []string{"archive", "restore"} {
		requests, err := runRegisteredDomainCommand(t, "stock", "location-"+action, guid)
		if err != nil {
			t.Fatalf("location-%s: %v", action, err)
		}
		if len(requests) != 1 {
			t.Fatalf("location-%s sent %d requests, want 1", action, len(requests))
		}
		request := requests[0]
		if request.method != http.MethodPost || request.path != "/api/stock/locations/"+guid+"/"+action {
			t.Fatalf("unexpected %s route: %s %s", action, request.method, request.path)
		}
		if len(request.query) != 0 || len(request.body) != 0 {
			t.Fatalf("location-%s leaked arguments: query=%v body=%v", action, request.query, request.body)
		}
	}
}

func TestStockLocationArchiveRefusesBadGuidsBeforeSending(t *testing.T) {
	cases := [][]string{
		{"location-archive", "00000000-0000-0000-0000-000000000000"},
		{"location-archive", "not-a-guid"},
		{"location-restore"},
	}
	for _, args := range cases {
		requests, err := runRegisteredDomainCommand(t, "stock", args...)
		if err == nil || len(requests) != 0 {
			t.Fatalf("%v must be refused before any request; err=%v requests=%v", args, err, requests)
		}
	}
}

func TestStockLocationsListCanIncludeArchived(t *testing.T) {
	requests, err := runRegisteredDomainCommand(t, "stock", "locations", "--include-archived")
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 1 || requests[0].path != "/api/stock/locations" {
		t.Fatalf("unexpected list request: %v", requests)
	}
	if got := requests[0].query["includeArchived"]; len(got) != 1 || got[0] != "true" {
		t.Fatalf("includeArchived query = %v, want [true]", got)
	}
}
