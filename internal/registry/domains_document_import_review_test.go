package registry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

const (
	documentBatchGUID = "11111111-1111-4111-8111-111111111111"
	documentGUID      = "22222222-2222-4222-8222-222222222222"
	reviewerGUID      = "33333333-3333-4333-8333-333333333333"
	ackGUID           = "44444444-4444-4444-8444-444444444444"
	tenantGUID        = "55555555-5555-4555-8555-555555555555"
)

func TestDocumentImportGetUsesGuidRESTAdapter(t *testing.T) {
	domain := findDomain("document-import")
	if domain == nil {
		t.Fatal("expected document-import domain to be registered")
	}
	if domain.APIPath != "/api/documentimport" {
		t.Fatalf("APIPath = %q, want /api/documentimport", domain.APIPath)
	}

	action := findAction(domain, "get")
	if action == nil {
		t.Fatal("expected get action on document-import domain")
	}
	if action.ToolName != "UteamupDocumentImportGetBatch" || action.MCPOnly ||
		action.HTTPMethod != http.MethodGet || action.RESTPath != "batch/{batchGuid}" {
		t.Fatalf("document-import get transport is invalid: %+v", action)
	}
	assertRequiredUUIDArg(t, action, "batchGuid")
	if len(action.Flags) != 0 {
		t.Fatalf("get flags = %+v, want no legacy batch-id flag", action.Flags)
	}

	path, consumed := buildRESTPath(domain, *action, map[string]any{"batchGuid": documentBatchGUID})
	if path != "/api/documentimport/batch/"+documentBatchGUID {
		t.Fatalf("resolved path = %q", path)
	}
	if len(consumed) != 1 || consumed[0] != "batchGuid" {
		t.Fatalf("consumed args = %v, want [batchGuid]", consumed)
	}
}

func TestDocumentImportGetPreservesGuidOnlyResponse(t *testing.T) {
	domain := findDomain("document-import")
	if domain == nil {
		t.Fatal("expected document-import domain to be registered")
	}
	action := findAction(domain, "get")
	if action == nil {
		t.Fatal("expected get action on document-import domain")
	}
	path, _ := buildRESTPath(domain, *action, map[string]any{"batchGuid": documentBatchGUID})

	apiClient := newDocumentContractClient(t, func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != path || request.URL.RawQuery != "" {
			t.Errorf("request = %s %s, want GET %s", request.Method, request.URL.RequestURI(), path)
		}
		assertDocumentRequestScope(t, request)
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{
			"guid":"` + documentBatchGUID + `",
			"tenantGuid":"` + tenantGUID + `",
			"projectGuid":"66666666-6666-4666-8666-666666666666",
			"workorderGuid":"77777777-7777-4777-8777-777777777777",
			"createdByGuid":"` + reviewerGUID + `",
			"items":[{
				"documentGuid":"` + documentGUID + `",
				"categoryGuid":"88888888-8888-4888-8888-888888888888",
				"tags":[{"tagGuid":"99999999-9999-4999-8999-999999999999"}],
				"assetGuids":["aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"],
				"codeCatalogEntries":[{
					"codeCatalogEntryGuid":"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
				}]
			}]
		}`))
	})

	result, err := apiClient.CallREST(
		context.Background(),
		action.HTTPMethod,
		path,
		nil,
		nil,
		action.Name,
	)
	if err != nil {
		t.Fatalf("document-import get transport failed: %v", err)
	}
	var batch map[string]json.RawMessage
	if err := json.Unmarshal(result, &batch); err != nil {
		t.Fatalf("decode document-import response: %v", err)
	}
	for _, field := range []string{"guid", "tenantGuid", "projectGuid", "workorderGuid", "createdByGuid", "items"} {
		if _, ok := batch[field]; !ok {
			t.Errorf("document-import response is missing %q: %s", field, result)
		}
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(batch["items"], &items); err != nil || len(items) != 1 {
		t.Fatalf("decode document-import items: %v (%s)", err, batch["items"])
	}
	for _, field := range []string{"documentGuid", "categoryGuid", "tags", "assetGuids", "codeCatalogEntries"} {
		if _, ok := items[0][field]; !ok {
			t.Errorf("document-import item is missing %q: %s", field, batch["items"])
		}
	}
	var decoded any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("decode GUID-only response: %v", err)
	}
	assertNoLegacyDocumentIdentityKeys(t, decoded)
}

func TestDocumentReviewQueueUsesGuidFilterAndPaginationQuery(t *testing.T) {
	domain := findDomain("document-review")
	if domain == nil {
		t.Fatal("expected document-review domain to be registered")
	}
	if domain.APIPath != "/api/documentreview" {
		t.Fatalf("APIPath = %q, want /api/documentreview", domain.APIPath)
	}

	action := findAction(domain, "queue")
	if action == nil {
		t.Fatal("expected queue action on document-review domain")
	}
	if action.ToolName != "UteamupDocumentReviewQueue" || action.MCPOnly ||
		action.HTTPMethod != http.MethodGet || action.RESTPath != "queue" {
		t.Fatalf("document-review queue transport is invalid: %+v", action)
	}

	flags := flagsToMap(action.Flags)
	assertQueryFlag(t, flags, "page", "page", "int", 1)
	assertQueryFlag(t, flags, "page-size", "pageSize", "int", 25)
	assertQueryFlag(t, flags, "batch-guid", "batchGuid", "uuid", nil)
	if _, leaked := flags["batch-id"]; leaked {
		t.Fatal("queue still exposes the legacy batch-id filter")
	}

	path, consumed := buildRESTPath(domain, *action, nil)
	path = appendQueryParameters(path, map[string]any{
		flags["page"].QueryName:       2,
		flags["page-size"].QueryName:  50,
		flags["batch-guid"].QueryName: documentBatchGUID,
	})
	want := "/api/documentreview/queue?batchGuid=" + documentBatchGUID + "&page=2&pageSize=50"
	if path != want || len(consumed) != 0 {
		t.Fatalf("resolved queue request = %q consumed=%v, want %q", path, consumed, want)
	}

	apiClient := newDocumentContractClient(t, func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.RequestURI() != want {
			t.Errorf("request = %s %s, want GET %s", request.Method, request.URL.RequestURI(), want)
		}
		assertDocumentRequestScope(t, request)
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`[{
			"documentGuid":"` + documentGUID + `",
			"importBatchGuid":"` + documentBatchGUID + `",
			"acknowledgmentCount":1,
			"effectiveReviewerThreshold":2,
			"ackedByCurrentUser":true
		}]`))
	})
	result, err := apiClient.CallREST(
		context.Background(),
		action.HTTPMethod,
		path,
		nil,
		nil,
		action.Name,
	)
	if err != nil {
		t.Fatalf("document-review queue transport failed: %v", err)
	}
	var queue []map[string]json.RawMessage
	if err := json.Unmarshal(result, &queue); err != nil || len(queue) != 1 {
		t.Fatalf("decode document-review queue: %v (%s)", err, result)
	}
	for _, field := range []string{"documentGuid", "importBatchGuid"} {
		if _, ok := queue[0][field]; !ok {
			t.Errorf("document-review queue item is missing %q: %s", field, result)
		}
	}
	var decoded any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("decode GUID-only queue: %v", err)
	}
	assertNoLegacyDocumentIdentityKeys(t, decoded)
}

func TestDocumentReviewAcknowledgeSendsOnlyCommentAndPreservesGuidResponse(t *testing.T) {
	domain := findDomain("document-review")
	if domain == nil {
		t.Fatal("expected document-review domain to be registered")
	}
	action := findAction(domain, "acknowledge")
	if action == nil {
		t.Fatal("expected acknowledge action on document-review domain")
	}
	if action.ToolName != "UteamupDocumentReviewAcknowledge" || action.MCPOnly ||
		action.HTTPMethod != http.MethodPost || action.RESTPath != "{documentGuid}/acknowledge" {
		t.Fatalf("document-review acknowledge transport is invalid: %+v", action)
	}
	assertRequiredUUIDArg(t, action, "documentGuid")
	flags := flagsToMap(action.Flags)
	comment, ok := flags["comment"]
	if !ok || comment.Required || comment.Type != "string" || comment.BodyName != "comment" {
		t.Fatalf("acknowledge comment flag = %+v", comment)
	}
	if _, leaked := flags["document-id"]; leaked {
		t.Fatal("acknowledge still exposes the legacy document-id flag")
	}

	arguments := map[string]any{
		"documentGuid":   documentGUID,
		comment.BodyName: "Reviewed and accepted",
	}
	path, consumed := buildRESTPath(domain, *action, arguments)
	for _, name := range consumed {
		delete(arguments, name)
	}
	if path != "/api/documentreview/"+documentGUID+"/acknowledge" {
		t.Fatalf("resolved acknowledge path = %q", path)
	}
	if len(consumed) != 1 || consumed[0] != "documentGuid" {
		t.Fatalf("consumed args = %v, want [documentGuid]", consumed)
	}

	apiClient := newDocumentContractClient(t, func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != path {
			t.Errorf("request = %s %s, want POST %s", request.Method, request.URL.Path, path)
		}
		if request.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			t.Error("acknowledge request is missing the CSRF transport header")
		}
		assertDocumentRequestScope(t, request)
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read acknowledge body: %v", err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("decode acknowledge body: %v", err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(payload) != 1 || payload["comment"] != "Reviewed and accepted" {
			t.Errorf("acknowledge payload = %s, want comment only", body)
		}
		if _, leaked := payload["documentGuid"]; leaked {
			t.Error("path documentGuid leaked into the acknowledge body")
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{
			"guid":"` + ackGUID + `",
			"documentGuid":"` + documentGUID + `",
			"userGuid":"` + reviewerGUID + `",
			"userName":"Reviewer",
			"ackedAt":"2026-08-06T12:00:00Z",
			"comment":"Reviewed and accepted"
		}`))
	})

	result, err := apiClient.CallREST(
		context.Background(),
		action.HTTPMethod,
		path,
		arguments,
		nil,
		action.Name,
	)
	if err != nil {
		t.Fatalf("acknowledge transport failed: %v", err)
	}
	var response map[string]json.RawMessage
	if err := json.Unmarshal(result, &response); err != nil {
		t.Fatalf("decode acknowledge response: %v", err)
	}
	for _, field := range []string{"guid", "documentGuid", "userGuid", "userName", "ackedAt", "comment"} {
		if _, ok := response[field]; !ok {
			t.Errorf("acknowledge response is missing %q: %s", field, result)
		}
	}
	for _, legacy := range []string{"id", "documentId", "userId"} {
		if _, ok := response[legacy]; ok {
			t.Errorf("acknowledge response leaked legacy field %q: %s", legacy, result)
		}
	}
	var decoded any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("decode GUID-only acknowledgement: %v", err)
	}
	assertNoLegacyDocumentIdentityKeys(t, decoded)
}

func assertRequiredUUIDArg(t *testing.T, action *Action, name string) {
	t.Helper()
	if len(action.Args) != 1 || action.Args[0].Name != name ||
		action.Args[0].Type != "uuid" || !action.Args[0].Required {
		t.Fatalf("%s args = %+v, want one required uuid %s", action.Name, action.Args, name)
	}
}

func assertQueryFlag(
	t *testing.T,
	flags map[string]FlagDef,
	name string,
	queryName string,
	flagType string,
	defaultValue any,
) {
	t.Helper()
	flag, ok := flags[name]
	if !ok || flag.QueryName != queryName || flag.Type != flagType || flag.Default != defaultValue {
		t.Fatalf("query flag %q = %+v, want QueryName=%q Type=%q Default=%v", name, flag, queryName, flagType, defaultValue)
	}
}

func newDocumentContractClient(t *testing.T, handler http.HandlerFunc) *client.APIClient {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := auth.SaveToken(&auth.TokenData{
		AccessToken: "document-contract-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  tenantGUID,
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	return client.NewAPIClient(
		server.URL,
		time.Second,
		true,
		client.RetryOptions{MaxRetries: 0},
		logging.New(logging.LevelError),
	)
}

func assertDocumentRequestScope(t *testing.T, request *http.Request) {
	t.Helper()
	if request.Header.Get("Authorization") != "Bearer document-contract-token" {
		t.Errorf("authorization header = %q", request.Header.Get("Authorization"))
	}
	if request.Header.Get("X-Tenant-Guid") != tenantGUID {
		t.Errorf("tenant GUID header = %q", request.Header.Get("X-Tenant-Guid"))
	}
	if request.Header.Get("X-Tenant-ID") != "" {
		t.Errorf("integer tenant identity leaked in request header: %q", request.Header.Get("X-Tenant-ID"))
	}
}

func assertNoLegacyDocumentIdentityKeys(t *testing.T, value any) {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			switch key {
			case "id", "tenantId", "projectId", "workorderId", "documentId", "categoryId",
				"tagId", "assetIds", "codeCatalogEntryId", "importBatchId", "userId":
				t.Errorf("document contract leaked legacy identity key %q", key)
			}
			assertNoLegacyDocumentIdentityKeys(t, child)
		}
	case []any:
		for _, child := range typed {
			assertNoLegacyDocumentIdentityKeys(t, child)
		}
	}
}
