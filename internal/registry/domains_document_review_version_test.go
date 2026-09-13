package registry

import (
	"net/http"
	"testing"
)

func TestDocumentReviewExactVersionUsesBodyFields(t *testing.T) {
	domain := findDomain("document-review")
	action := findAction(domain, "acknowledge")
	flags := flagsToMap(action.Flags)
	version := flags["document-version-guid"]
	stamp := flags["expected-document-updated-at"]
	if version.Type != "uuid" || version.BodyName != "documentVersionGuid" || version.QueryName != "" ||
		stamp.BodyName != "expectedDocumentUpdatedAt" || stamp.QueryName != "" {
		t.Fatalf("exact review must send GUID and observed token in body: %+v, %+v", version, stamp)
	}
	args := map[string]any{"documentGuid": documentGUID, version.BodyName: reviewerGUID,
		stamp.BodyName: "2026-09-13T10:30:00Z"}
	path, consumed := buildRESTPath(domain, *action, args)
	for _, key := range consumed {
		delete(args, key)
	}
	if action.HTTPMethod != http.MethodPost || path != "/api/documentreview/"+documentGUID+"/acknowledge" ||
		len(args) != 2 || args[version.BodyName] != reviewerGUID || args[stamp.BodyName] != "2026-09-13T10:30:00Z" {
		t.Fatalf("exact review transport changed: %s %+v", path, args)
	}
}

func TestDocumentReviewOriginalReceiptUsesExactGUIDPath(t *testing.T) {
	domain := findDomain("document-review")
	action := findAction(domain, "receipt")
	if action == nil || action.HTTPMethod != http.MethodGet || action.ToolName != "UteamupDocumentReviewAcknowledgment" {
		t.Fatalf("missing receipt read action: %+v", action)
	}
	if len(action.Args) != 2 ||
		action.Args[0].Name != "documentGuid" || action.Args[0].Type != "uuid" || !action.Args[0].Required ||
		action.Args[1].Name != "acknowledgmentGuid" || action.Args[1].Type != "uuid" || !action.Args[1].Required {
		t.Fatalf("receipt args = %+v, want required uuid documentGuid and acknowledgmentGuid", action.Args)
	}
	path, _ := buildRESTPath(domain, *action, map[string]any{"documentGuid": documentGUID, "acknowledgmentGuid": ackGUID})
	if path != "/api/documentreview/"+documentGUID+"/acknowledgments/"+ackGUID {
		t.Fatalf("receipt path = %s", path)
	}
}
