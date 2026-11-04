package registry

import (
	"os"
	"path/filepath"
	"testing"
)

const commentTestPart = "11111111-1111-4111-8111-111111111111"
const commentTestGuid = "22222222-2222-4222-8222-222222222222"

func TestPartCommentListUsesGuidRouteAndBoundedPageParameters(t *testing.T) {
	requests, err := runRegisteredDomainCommand(t, "comment", "list-by-part", commentTestPart,
		"--page", "2", "--page-size", "3")
	if err != nil || len(requests) != 1 {
		t.Fatalf("list failed: %v, requests=%d", err, len(requests))
	}
	r := requests[0]
	if r.method != "GET" || r.path != "/api/comment/part/"+commentTestPart+"/paginated" {
		t.Fatalf("wrong route: %s %s", r.method, r.path)
	}
	if len(r.query["page"]) != 1 || r.query["page"][0] != "2" || len(r.query["pageSize"]) != 1 || r.query["pageSize"][0] != "3" {
		t.Fatalf("bounded page lost: %v", r.query)
	}
	if _, leaked := r.query["partGuid"]; leaked {
		t.Fatal("part identity belongs in the route only")
	}
}

func TestPartCommentCreateAndUpdateKeepOnlyTheReviewedTextInBody(t *testing.T) {
	input := filepath.Join(t.TempDir(), "comment.json")
	if err := os.WriteFile(input, []byte(`{"text":"Consider Wi-Fi 8 when purchasing"}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []struct {
		action, method, path string
		args                 []string
	}{
		{"create-for-part", "POST", "/api/comment/part/" + commentTestPart, []string{commentTestPart}},
		{"update-for-part", "PUT", "/api/comment/part/" + commentTestPart + "/" + commentTestGuid, []string{commentTestPart, commentTestGuid}},
	} {
		args := append([]string{entry.action}, entry.args...)
		args = append(args, "--from-json", input)
		requests, err := runRegisteredDomainCommand(t, "comment", args...)
		if err != nil || len(requests) != 1 {
			t.Fatalf("%s failed: %v, requests=%d", entry.action, err, len(requests))
		}
		r := requests[0]
		if r.method != entry.method || r.path != entry.path || len(r.body) != 1 || r.body["text"] != "Consider Wi-Fi 8 when purchasing" {
			t.Fatalf("%s changed scope/body: %s %s %v", entry.action, r.method, r.path, r.body)
		}
		if r.header.Get("X-Requested-With") != "XMLHttpRequest" {
			t.Fatal("comment mutation lost the existing CSRF header")
		}
	}
}

func TestCommentGenericGuidReadAndRemovalUseCanonicalRoutes(t *testing.T) {
	for _, entry := range []struct{ action, method string }{{"get", "GET"}, {"delete", "DELETE"}} {
		requests, err := runRegisteredDomainCommand(t, "comment", entry.action, commentTestGuid)
		if err != nil || len(requests) != 1 {
			t.Fatalf("%s failed: %v", entry.action, err)
		}
		if requests[0].method != entry.method || requests[0].path != "/api/comment/by-guid/"+commentTestGuid {
			t.Fatalf("%s still uses the obsolete integer-style route", entry.action)
		}
	}
}

func TestPartCommentInvalidIdentifiersFailBeforeNetwork(t *testing.T) {
	input := filepath.Join(t.TempDir(), "comment.json")
	if err := os.WriteFile(input, []byte(`{"text":"Observation"}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"list-by-part", "42"},
		{"list-by-part", emptyUUIDValue},
		{"create-for-part", "42"},
		{"update-for-part", commentTestPart, "42"},
		{"update-for-part", "42", commentTestGuid},
		{"get", "42"},
		{"delete", emptyUUIDValue},
	} {
		if args[0] == "create-for-part" || args[0] == "update-for-part" {
			args = append(args, "--from-json", input)
		}
		requests, err := runRegisteredDomainCommand(t, "comment", args...)
		if err == nil || len(requests) != 0 {
			t.Errorf("invalid identity reached the network: args=%v, err=%v", args, err)
		}
	}
}

func TestCommentRegistryExposesOnlyExistingGuidSafeTools(t *testing.T) {
	actions := append(commentActions(), commentAssetActions()...)
	for _, action := range actions {
		if action.ToolName == "UteamupCommentList" || action.ToolName == "UteamupCommentUpdate" {
			t.Fatalf("phantom generic tool advertised: %s", action.ToolName)
		}
		for _, arg := range action.Args {
			if arg.Type != "non-empty-uuid" || !arg.Required {
				t.Errorf("%s selector is not a required public GUID: %+v", action.Name, arg)
			}
		}
	}
}
