package registry

import (
	"net/http"
	"net/url"
	"testing"
)

func TestAssetTypeArchiveAndRestorePostToTheGuidRoutes(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for _, action := range []string{"archive", "restore"} {
		requests, err := runRegisteredDomainCommand(t, "asset-type", action, guid)
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if len(requests) != 1 {
			t.Fatalf("%s sent %d requests, want 1", action, len(requests))
		}
		request := requests[0]
		if request.method != http.MethodPost || request.path != "/api/asset-type/by-guid/"+guid+"/"+action {
			t.Fatalf("unexpected %s route: %s %s", action, request.method, request.path)
		}
		if len(request.query) != 0 {
			t.Fatalf("%s leaked arguments into the query: %v", action, request.query)
		}
	}
}

func TestAssetTypeArchiveRefusesEmptyAndMalformedGuidsBeforeSending(t *testing.T) {
	cases := [][]string{
		{"archive", "00000000-0000-0000-0000-000000000000"},
		{"archive", "not-a-guid"},
		{"restore", "00000000-0000-0000-0000-000000000000"},
		{"archive", "11111111-1111-4111-8111-111111111111", "extra"},
	}
	for _, args := range cases {
		requests, err := runRegisteredDomainCommand(t, "asset-type", args...)
		if err == nil || len(requests) != 0 {
			t.Fatalf("%v must be refused before any request; err=%v requests=%d", args, err, len(requests))
		}
	}
}

func TestAssetTypeListShowsArchivedOnlyWhenAsked(t *testing.T) {
	requests, err := runRegisteredDomainCommand(t, "asset-type", "list", "--archived")
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 1 || requests[0].path != "/api/asset-type" || url.Values(requests[0].query).Get("archived") != "true" {
		t.Fatalf("list --archived must ask for archived types: %+v", requests)
	}
}
