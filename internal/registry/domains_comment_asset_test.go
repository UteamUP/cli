package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAssetCommentListCarriesActivityAndPageToGuidRoute(t *testing.T) {
	const asset = "11111111-1111-4111-8111-111111111111"
	const activity = "22222222-2222-4222-8222-222222222222"
	requests, err := runRegisteredDomainCommand(t, "comment", "list-by-asset", asset,
		"--activity-log-guid", activity, "--page", "2", "--page-size", "15")
	if err != nil || len(requests) != 1 {
		t.Fatalf("request failed: %v, requests=%d", err, len(requests))
	}
	r := requests[0]
	if r.method != "GET" || r.path != "/api/comment/asset/"+asset+"/paginated" {
		t.Fatalf("wrong read route: %s %s", r.method, r.path)
	}
	for key, expected := range map[string]string{"activityLogGuid": activity, "page": "2", "pageSize": "15"} {
		if len(r.query[key]) != 1 || r.query[key][0] != expected {
			t.Errorf("query %s = %v, expected %s", key, r.query[key], expected)
		}
	}
	if _, leaked := r.query["assetGuid"]; leaked {
		t.Fatal("asset GUID must stay in the route")
	}
}

func TestAssetCommentCreatePreservesReplyBody(t *testing.T) {
	const asset = "11111111-1111-4111-8111-111111111111"
	const activity = "22222222-2222-4222-8222-222222222222"
	input := filepath.Join(t.TempDir(), "comment.json")
	if err := os.WriteFile(input, []byte(`{"text":"Reason for change","activityLogGuid":"`+activity+`"}`), 0600); err != nil {
		t.Fatal(err)
	}
	requests, err := runRegisteredDomainCommand(t, "comment", "create-for-asset", asset,
		"--from-json", input)
	if err != nil || len(requests) != 1 {
		t.Fatalf("request failed: %v, requests=%d", err, len(requests))
	}
	r := requests[0]
	if r.method != "POST" || r.path != "/api/comment/asset/"+asset {
		t.Fatalf("wrong create route: %s %s", r.method, r.path)
	}
	if r.body["activityLogGuid"] != activity || r.body["text"] != "Reason for change" {
		t.Fatalf("reply body lost: %v", r.body)
	}
	if r.header.Get("X-Requested-With") != "XMLHttpRequest" {
		t.Fatal("comment mutation must retain the CSRF header")
	}
}

func TestAssetCommentListRejectsInvalidGuidsBeforeNetwork(t *testing.T) {
	const asset = "11111111-1111-4111-8111-111111111111"
	for _, args := range [][]string{
		{"list-by-asset", "42"},
		{"list-by-asset", "00000000-0000-0000-0000-000000000000"},
		{"list-by-asset", asset, "--activity-log-guid", "42"},
	} {
		requests, err := runRegisteredDomainCommand(t, "comment", args...)
		if err == nil || len(requests) != 0 {
			t.Errorf("invalid GUID sent a request: %v, error=%v", args, err)
		}
	}
}
