package registry

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAssetArchivePreviewUsesReviewedSourceAndTargetGuids(t *testing.T) {
	const source = "11111111-1111-4111-8111-111111111111"
	const target = "22222222-2222-4222-8222-222222222222"
	request := runAssetCommand(t, "archive-preview", source, "--target-asset-guid", target)
	if request.Method != http.MethodGet || request.URL.Path != "/api/asset/by-guid/"+source+"/archive-preview" {
		t.Fatalf("unexpected preview route: %s %s", request.Method, request.URL)
	}
	assertAssetQuery(t, request.URL.Query(), map[string]string{"targetAssetGuid": target})
	if request.URL.Query().Has("assetGuid") || request.URL.Query().Has("tenantGuid") {
		t.Fatalf("source or tenant leaked into query: %s", request.URL)
	}
}

func TestAssetArchivePostsOnlyReviewedMigration(t *testing.T) {
	const source = "11111111-1111-4111-8111-111111111111"
	const target = "22222222-2222-4222-8222-222222222222"
	token := strings.Repeat("a", 64)
	request := runAssetCommand(t, "archive", source, "--review-token", token, "--target-asset-guid", target)
	if request.Method != http.MethodPost || request.URL.Path != "/api/asset/by-guid/"+source+"/archive" {
		t.Fatalf("unexpected archive route: %s %s", request.Method, request.URL)
	}
	var body map[string]any
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["reviewToken"] != token || body["targetAssetGuid"] != target || len(body) != 2 {
		t.Fatalf("archive must post only the reviewed token and target: %+v", body)
	}
	if request.URL.RawQuery != "" {
		t.Fatalf("archive unexpectedly used query parameters: %s", request.URL)
	}
}

func TestAssetArchiveRequiresReviewAndGuidBoundaries(t *testing.T) {
	for _, name := range []string{"archive-preview", "archive"} {
		action := findAssetAction(t, name)
		if len(action.Args) != 1 || !action.Args[0].Required || action.Args[0].Type != "non-empty-uuid" {
			t.Fatalf("%s must require a source GUID: %+v", name, action.Args)
		}
	}
	action := findAssetAction(t, "archive")
	for _, flag := range action.Flags {
		if flag.Name == "review-token" && flag.Required {
			return
		}
	}
	t.Fatal("archive must require the reviewed token")
}
