package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestAssetStageRequestUsesTheExistingGuidRouteAndObservedVersion(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	const guid = "55555555-5555-4555-8555-555555555555"
	const key = "66666666-6666-4666-8666-666666666666"
	var body map[string]any
	requests := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != "POST" || r.URL.Path != "/api/assetlifecyclestage/asset/by-guid/"+guid+"/request" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			t.Error("stage request omitted the CSRF header")
		}
		if r.URL.RawQuery != "" {
			t.Errorf("stage request moved body fields into query: %s", r.URL.RawQuery)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"externalGuid":"77777777-7777-4777-8777-777777777777","status":"pending"}`))
	}))
	defer server.Close()
	if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "stage-test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	logger := logging.New(logging.LevelError)
	format := "json"
	factory := func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
	}
	command := buildDomainCommand(findDomain("asset-state"), factory, logger, &format, &ExportConfig{})
	command.SetArgs([]string{"request-stage", guid, "--to-stage", "decommissioned", "--expected-lifecycle-version", "0", "--idempotency-key", key, "--notes", "Taken out of service"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("sent %d requests, want 1", requests)
	}
	for field, want := range map[string]any{"toStage": "decommissioned", "expectedLifecycleVersion": float64(0), "idempotencyKey": key, "notes": "Taken out of service"} {
		if body[field] != want {
			t.Errorf("body[%s] = %v, want %v", field, body[field], want)
		}
	}
	for _, field := range []string{"assetGuid", "assetId", "approve", "model"} {
		if _, leaked := body[field]; leaked {
			t.Errorf("unexpected field %s in the REST stage request", field)
		}
	}
}

func TestAssetStageRequestRequiresGuidVersionAndStableReplayKey(t *testing.T) {
	domain := findDomain("asset-state")
	var action *Action
	for i := range domain.Actions {
		if domain.Actions[i].Name == "request-stage" {
			action = &domain.Actions[i]
		}
	}
	if action == nil || action.ToolName != "UteamupAssetStateRequestStage" || action.MCPOnly {
		t.Fatal("request-stage must mirror its MCP tool using the existing REST adapter")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "assetGuid" || action.Args[0].Type != "non-empty-uuid" {
		t.Fatalf("request-stage must accept a public asset GUID: %+v", action.Args)
	}
	for _, name := range []string{"to-stage", "expected-lifecycle-version", "idempotency-key"} {
		found := false
		for _, flag := range action.Flags {
			if flag.Name == name {
				found = flag.Required
				if name == "idempotency-key" && flag.Type != "non-empty-uuid" {
					t.Error("idempotency-key must be a non-empty GUID")
				}
			}
		}
		if !found {
			t.Errorf("missing required request-stage flag %s", name)
		}
	}
}
