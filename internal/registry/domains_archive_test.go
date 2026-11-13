package registry

import (
	"net/http"
	"testing"
)

func TestArchiveListSendsEveryFilterToTheQuery(t *testing.T) {
	const archiver = "22222222-2222-4222-8222-222222222222"
	request, err := runReferenceCommand(t, "archive", "list", "--kind", "contact", "--search", "pump",
		"--archived-by", archiver, "--from", "2026-11-01T00:00:00Z", "--to", "2026-11-13T23:59:59Z",
		"--page", "2", "--page-size", "50")
	if err != nil {
		t.Fatalf("archive list: %v", err)
	}
	if request.Method != http.MethodGet || request.URL.Path != "/api/archive" {
		t.Fatalf("archive list = %s %s, want GET /api/archive", request.Method, request.URL.Path)
	}
	query := request.URL.Query()
	want := map[string]string{
		"kind": "contact", "search": "pump", "archivedByUserGuid": archiver, "from": "2026-11-01T00:00:00Z",
		"to": "2026-11-13T23:59:59Z", "page": "2", "pageSize": "50",
	}
	for name, value := range want {
		if query.Get(name) != value {
			t.Fatalf("query %s = %q, want %q (full query %s)", name, query.Get(name), value, request.URL.RawQuery)
		}
	}
}

func TestArchiveGetReadsOneRecordByKindAndGuid(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	request, err := runReferenceCommand(t, "archive", "get", "asset", guid)
	if err != nil {
		t.Fatalf("archive get: %v", err)
	}
	if request.Method != http.MethodGet || request.URL.Path != "/api/archive/asset/"+guid {
		t.Fatalf("archive get = %s %s", request.Method, request.URL.Path)
	}
}

func TestArchiveGetRefusesAnUnknownKindOrAnEmptyGuidBeforeSending(t *testing.T) {
	cases := [][]string{
		{"get", "spaceship", "11111111-1111-4111-8111-111111111111"},
		{"get", "asset", "00000000-0000-0000-0000-000000000000"},
		{"get", "asset"},
	}
	for _, args := range cases {
		request, err := runReferenceCommand(t, "archive", args...)
		if err == nil || request != nil {
			t.Fatalf("archive %v: want a refusal before sending, got err=%v request=%v", args, err, request)
		}
	}
}

func TestArchiveDomainIsReadOnly(t *testing.T) {
	domain := findDomain("archive")
	if domain == nil {
		t.Fatal("archive domain not registered")
	}
	for _, action := range domain.Actions {
		if action.HTTPMethod != "GET" {
			t.Fatalf("archive %s uses %s; the archive is read only", action.Name, action.HTTPMethod)
		}
	}
	if len(domain.Actions) != 2 {
		t.Fatalf("archive has %d actions, want list and get", len(domain.Actions))
	}
}
