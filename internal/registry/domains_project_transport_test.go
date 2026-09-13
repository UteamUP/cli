package registry

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

const (
	projectTransportProjectGUID  = "11111111-1111-4111-8111-111111111111"
	projectTransportResourceGUID = "22222222-2222-4222-8222-222222222222"
	projectTransportRequestGUID  = "33333333-3333-4333-8333-333333333333"
	projectTransportTenantGUID   = "44444444-4444-4444-8444-444444444444"
)

// These tests inspect the actual command transport, not application validation.
// The opaque fixture deliberately includes nested/new fields and explicit nulls
// that must reach the owning DTO unchanged as its schema evolves.
func TestProjectReviewedCommandsSendRootJSONForBothTokenModes(t *testing.T) {
	for _, mode := range []string{"login", "apikey"} {
		for _, scenario := range []struct {
			domain, action, method, suffix string
			resource                       bool
		}{
			{"project-process", "preview", "POST", "/process/preview", false},
			{"project-process", "adopt", "POST", "/process/adopt", false},
			{"project-gate", "preview", "POST", "/stages/{resource}/decisions/submissions/preview", true},
			{"project-gate", "submit", "POST", "/stages/{resource}/decisions/submissions", true},
			{"project-gate", "decide", "POST", "/stages/{resource}/decisions", true},
			{"project-cost", "create", "POST", "/cost-records", false},
			{"project-deliverable-cost", "plan", "PUT", "/outputitems/{resource}/costs/plan", true},
			{"project-cost-reconciliation", "review", "POST", "/cost-records/reconciliation", false},
			{"project-member", "add", "POST", "/members", false},
			{"project-member", "update", "PUT", "/members/{resource}", true},
			{"project-dependency", "add", "POST", "/dependencies", false},
			{"project-comment", "add", "POST", "/comments", false},
			{"project-comment", "update", "PUT", "/comments/{resource}", true},
			{"project-baseline", "capture", "POST", "/baselines", false},
			{"project-change-request", "create", "POST", "/change-requests", false},
			{"project-change-request", "submit", "POST", "/change-requests/{resource}/submit", true},
			{"project-change-request", "approve", "POST", "/change-requests/{resource}/approve", true},
			{"project-change-request", "reject", "POST", "/change-requests/{resource}/reject", true},
			{"project-change-request", "apply", "POST", "/change-requests/{resource}/apply", true},
			{"project-schedule", "preview", "POST", "/timeline/preview", false},
			{"project-schedule", "apply", "POST", "/timeline/apply", false},
			{"project-schedule", "calendar", "PUT", "/timeline/calendar", false},
			{"project-schedule", "dependency-create", "POST", "/timeline/dependencies", false},
			{"project-schedule", "dependency-update", "PUT", "/timeline/dependencies/{resource}", true},
			{"project-schedule", "dependency-remove", "POST", "/timeline/dependencies/{resource}/remove", true},
		} {
			t.Run(mode+"/"+scenario.domain+"/"+scenario.action, func(t *testing.T) {
				action := findDomainAction(t, scenario.domain, scenario.action)
				fileFlag := findFlag(action, "from-json")
				if fileFlag == nil || !fileFlag.Required || !fileFlag.RootJSONObjectFile || fileFlag.JSONFile || action.MCPOnly {
					t.Fatalf("reviewed REST action must require a root JSON file: %+v", action)
				}
				body := projectTransportPayload()
				var want map[string]any
				if err := json.Unmarshal([]byte(body), &want); err != nil {
					t.Fatal(err)
				}
				var calls atomic.Int32
				apiClient := projectTransportClient(t, mode, client.RetryOptions{}, func(response http.ResponseWriter, request *http.Request) {
					calls.Add(1)
					wantPath := "/api/projects/" + projectTransportProjectGUID + strings.ReplaceAll(scenario.suffix, "{resource}", projectTransportResourceGUID)
					if request.Method != scenario.method || request.URL.Path != wantPath || request.URL.RawQuery != "" {
						t.Errorf("request = %s %s, want %s %s", request.Method, request.URL, scenario.method, wantPath)
					}
					if request.Header.Get("X-Requested-With") != "XMLHttpRequest" {
						t.Error("mutation lost its CSRF request header")
					}
					var got map[string]any
					if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
						t.Errorf("decode body: %v", err)
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("body = %#v, want exact request %#v", got, want)
					}
					_, _ = response.Write([]byte(`{"retained":true}`))
				})
				args := []string{scenario.action, projectTransportProjectGUID}
				if scenario.resource {
					args = append(args, projectTransportResourceGUID)
				}
				args = append(args, "--from-json", writeRegistryJSONFixture(t, body))
				if err := executeProjectTransport(t, apiClient, scenario.domain, args); err != nil {
					t.Fatalf("command failed: %v", err)
				}
				if calls.Load() != 1 {
					t.Fatalf("HTTP calls = %d, want one owning REST request", calls.Load())
				}
			})
		}
	}
}

func TestProjectModelCommandsUseNamedMCPArgumentsForBothTokenModes(t *testing.T) {
	for _, mode := range []string{"login", "apikey"} {
		for _, scenario := range []struct{ action, tool string }{
			{"create", "UteamupProjectCreate"}, {"update", "UteamupProjectUpdate"},
		} {
			t.Run(mode+"/"+scenario.action, func(t *testing.T) {
				model := `{"name":"Equipment delivery","customerGuid":null,"locationGuids":[],"notes":"Retain reviewed scope"}`
				var wantModel map[string]any
				if err := json.Unmarshal([]byte(model), &wantModel); err != nil {
					t.Fatal(err)
				}
				want := map[string]any{"model": wantModel}
				args := []string{scenario.action}
				if scenario.action == "update" {
					want["projectGuid"] = projectTransportProjectGUID
					args = append(args, projectTransportProjectGUID)
				}
				var calls atomic.Int32
				apiClient := projectTransportClient(t, mode, client.RetryOptions{}, func(response http.ResponseWriter, request *http.Request) {
					calls.Add(1)
					if request.Method != http.MethodPost || request.URL.Path != "/mcp" {
						t.Errorf("typed model must use MCP; request = %s %s", request.Method, request.URL)
					}
					var rpc struct {
						JSONRPC string                `json:"jsonrpc"`
						Method  string                `json:"method"`
						Params  client.ToolCallParams `json:"params"`
					}
					if err := json.NewDecoder(request.Body).Decode(&rpc); err != nil {
						t.Errorf("decode RPC: %v", err)
					}
					if rpc.JSONRPC != "2.0" || rpc.Method != "tools/call" || rpc.Params.Name != scenario.tool || !reflect.DeepEqual(rpc.Params.Arguments, want) {
						t.Errorf("RPC = %+v, want %s arguments %#v", rpc, scenario.tool, want)
					}
					_, _ = response.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"externalGuid":"11111111-1111-4111-8111-111111111111"}}`))
				})
				args = append(args, "--from-json", writeRegistryJSONFixture(t, model))
				if err := executeProjectTransport(t, apiClient, "project", args); err != nil {
					t.Fatalf("model command failed: %v", err)
				}
				if calls.Load() != 1 {
					t.Fatalf("HTTP calls = %d, want one typed tool call", calls.Load())
				}
			})
		}
	}
}

func TestProjectScheduleRetryResendsIdenticalReviewedBodyAfter503(t *testing.T) {
	var bodies [][]byte
	var mu sync.Mutex
	apiClient := projectTransportClient(t, "apikey", client.RetryOptions{
		MaxRetries: 1, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond,
	}, func(response http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read attempted body: %v", err)
		}
		mu.Lock()
		defer mu.Unlock()
		bodies = append(bodies, body)
		if len(bodies) == 1 {
			response.WriteHeader(http.StatusServiceUnavailable)
			_, _ = response.Write([]byte(`{"error":"temporary gateway failure"}`))
			return
		}
		_, _ = response.Write([]byte(`{"retained":true}`))
	})
	err := executeProjectTransport(t, apiClient, "project-schedule", []string{
		"apply", projectTransportProjectGUID, "--from-json", writeRegistryJSONFixture(t, projectTransportPayload()),
	})
	if err != nil {
		t.Fatalf("retried reviewed command failed: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 2 || len(bodies[0]) == 0 || !bytes.Equal(bodies[0], bodies[1]) {
		t.Fatalf("attempts must carry identical nonempty bytes: %q", bodies)
	}
	var retained map[string]any
	if err := json.Unmarshal(bodies[1], &retained); err != nil || retained["requestGuid"] != projectTransportRequestGUID {
		t.Fatalf("retry lost original mutation identity: %#v, error %v", retained, err)
	}
}

func TestProjectReviewedCommandRejectsWrappedPayloadBeforeSending(t *testing.T) {
	var calls atomic.Int32
	apiClient := projectTransportClient(t, "login", client.RetryOptions{}, func(response http.ResponseWriter, request *http.Request) {
		calls.Add(1)
	})
	err := executeProjectTransport(t, apiClient, "project-process", []string{
		"adopt", projectTransportProjectGUID, "--from-json", writeRegistryJSONFixture(t, `{"model":{"reason":"Reviewed"}}`),
	})
	if err == nil || !strings.Contains(err.Error(), `without a "model" wrapper`) || calls.Load() != 0 {
		t.Fatalf("wrapped file must fail before sending: calls=%d, error=%v", calls.Load(), err)
	}
}

func projectTransportPayload() string {
	return `{"requestGuid":"` + projectTransportRequestGUID + `","expectedReviewFingerprint":"` + strings.Repeat("a", 64) +
		`","expectedUpdatedAt":"2026-09-13T09:12:13.123456Z","reviewNote":null,"changes":[{"node":{"kind":"deliverable","guid":"` +
		projectTransportResourceGUID + `"},"forecastStart":null}],"futureEvidence":{"verified":false,"references":[]}}`
}

func projectTransportClient(t *testing.T, mode string, retry client.RetryOptions, handler http.HandlerFunc) *client.APIClient {
	t.Helper()
	fixtureHome := t.TempDir()
	t.Setenv("HOME", fixtureHome)
	t.Setenv("USERPROFILE", fixtureHome)
	if err := auth.SaveToken(&auth.TokenData{
		AccessToken: "project-transport-token", ExpiresAt: time.Now().Add(time.Hour),
		AuthMethod: mode, TenantGUID: projectTransportTenantGUID,
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer project-transport-token" || request.Header.Get("X-Tenant-Guid") != projectTransportTenantGUID {
			t.Error("request lost the authenticated actor or tenant context")
		}
		response.Header().Set("Content-Type", "application/json")
		handler(response, request)
	}))
	t.Cleanup(server.Close)
	return client.NewAPIClient(server.URL, time.Second, true, retry, logging.New(logging.LevelError))
}

func executeProjectTransport(t *testing.T, apiClient *client.APIClient, domainName string, args []string) error {
	t.Helper()
	domain := findDomain(domainName)
	if domain == nil {
		t.Fatalf("missing domain %s", domainName)
	}
	format := "json"
	command := buildDomainCommand(domain, func() (*client.APIClient, error) { return apiClient, nil },
		logging.New(logging.LevelError), &format, &ExportConfig{})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs(args)
	return command.Execute()
}
