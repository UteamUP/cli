package registry

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/logging"
)

func TestRadioPilotAdministrationUsesExplicitSelectedTenantBoundary(t *testing.T) {
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name != "radio-admin" {
			continue
		}
		if domain.APIPath != "/api/admin/radio" {
			t.Fatal("pilot administration requires its platform-only boundary")
		}
		expected := map[string]string{"validate": "validate", "status": "environment", "grant": "complimentary", "revoke": "complimentary/revoke"}
		for _, action := range domain.Actions {
			if action.RESTPath != expected[action.Name] {
				t.Fatalf("wrong admin route for %s", action.Name)
			}
			if len(action.Args) != 0 {
				t.Fatal("pilot commands must use the explicitly selected tenant")
			}
			if action.Name == "status" {
				if action.HTTPMethod != "GET" {
					t.Fatal("status is read-only")
				}
			} else if action.Name == "validate" {
				if action.HTTPMethod != "POST" || len(action.Flags) != 0 || action.ToolName != "UteamupRadioAdministrationValidate" {
					t.Fatal("validation must use the selected-tenant platform probe without identity overrides")
				}
			} else if action.HTTPMethod != "POST" || len(action.Flags) != 1 || !action.Flags[0].Required || !action.Flags[0].RootJSONObjectFile {
				t.Fatal("pilot changes require an explicit reviewed JSON request")
			}
			delete(expected, action.Name)
		}
		if len(expected) != 0 {
			t.Fatalf("missing admin actions: %v", expected)
		}
		return
	}
	t.Fatal("radio-admin domain missing")
}

func TestRadioRoutesUseActiveTenantAndPublicGuids(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "radio" {
			domain = candidate
		}
	}
	if domain == nil {
		t.Fatal("radio domain missing")
	}
	if domain.APIPath != "/api/radio" {
		t.Fatal("radio must use its explicit REST boundary")
	}
	expected := map[string]string{"team-eligibility": "teams/{teamGuid}/eligibility", "backup-sharepoint": "backup-destinations/sharepoint", "validate": "validate", "stop-pilot": "complimentary/revoke", "transcript": "channels/{channelGuid}/transmissions/{transmissionGuid}/transcript", "transmissions": "channels/{channelGuid}/transmissions", "status": "environment", "policy": "policy", "channels": "history-channels", "recordings": "recordings", "always-transcribe": "me/always-transcribe", "transcribe": "recordings/{recordingGuid}/transcription", "get": "recordings/{recordingGuid}", "playback": "recordings/{recordingGuid}/playback"}
	for _, action := range domain.Actions {
		if action.RESTPath != expected[action.Name] {
			t.Errorf("wrong route for %s: %s", action.Name, action.RESTPath)
		}
		if !strings.HasPrefix(action.ToolName, "UteamupRadio") {
			t.Errorf("missing MCP equivalent for %s", action.Name)
		}
		for _, arg := range action.Args {
			if (arg.Name != "recordingGuid" && arg.Name != "channelGuid" && arg.Name != "transmissionGuid" && arg.Name != "teamGuid") || arg.Type != "non-empty-uuid" {
				t.Errorf("unexpected public argument %s", arg.Name)
			}
		}
		for _, flag := range action.Flags {
			if strings.Contains(normalize(flag.Name), "userid") || strings.Contains(normalize(flag.Name), "tenantid") {
				t.Errorf("identity override %s", flag.Name)
			}
		}
		if action.Name == "playback" && !action.DisableResponseExport {
			t.Error("playback URLs must not enter automatic exports")
		}
		if action.Name == "backup-sharepoint" && (action.HTTPMethod != "GET" || len(action.Args) != 0 || len(action.Flags) != 0 || action.ToolName != "UteamupRadioBackupSharePointAvailability") {
			t.Error("SharePoint availability must be read-only with no tenant or configuration override")
		}
		if action.Name == "validate" && (action.HTTPMethod != "POST" || len(action.Args) != 0 || len(action.Flags) != 0 || action.ToolName != "UteamupRadioValidate") {
			t.Error("validation must use the membership-bound probe without identity overrides")
		}
		if action.Name == "policy" && (action.HTTPMethod != "PUT" || !action.Flags[0].RootJSONObjectFile) {
			t.Error("complete reviewed policy must use a root JSON object")
		}
		if action.Name == "stop-pilot" && (action.HTTPMethod != "POST" || len(action.Flags) != 1 || !action.Flags[0].Required || !action.Flags[0].RootJSONObjectFile) {
			t.Error("stopping a pilot requires an explicit reviewed JSON request")
		}
		if action.Name == "always-transcribe" && (action.HTTPMethod != "PUT" || len(action.Args) != 0 || len(action.Flags) != 1 || action.Flags[0].Name != "enabled" || action.Flags[0].Type != "bool" || !action.Flags[0].Required || action.Flags[0].BodyName != "enabled" || action.Flags[0].QueryName != "" || action.ToolName != "UteamupRadioAlwaysTranscribeSet") {
			t.Error("always-transcribe must PUT one required enabled body field for the caller only")
		}
		if action.Name == "transcribe" && (action.HTTPMethod != "POST" || len(action.Args) != 1 || action.Args[0].Name != "recordingGuid" || len(action.Flags) != 0 || action.ToolName != "UteamupRadioTranscriptionRequest") {
			t.Error("transcribe must POST for one public recording GUID without extra fields")
		}
		if action.Name == "recordings" {
			wantQuery := map[string]string{"channel-guid": "channelGuid", "scope": "scope", "from": "from", "to": "to", "team-guid": "teamGuid", "has-transcript": "hasTranscript", "transcript-contains": "transcriptContains"}
			for _, flag := range action.Flags {
				if want, ok := wantQuery[flag.Name]; ok {
					if flag.QueryName != want || flag.Required || flag.Default != nil {
						t.Errorf("recordings --%s must be an optional query filter %s", flag.Name, want)
					}
					delete(wantQuery, flag.Name)
				}
			}
			if len(wantQuery) != 0 {
				t.Errorf("recordings is missing history filters: %v", wantQuery)
			}
		}
		delete(expected, action.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing radio actions: %v", expected)
	}
}

func TestRadioBackupSharePointUsesSelectedTenantAndPreservesAvailability(t *testing.T) {
	domain := findDomain("radio")
	if domain == nil {
		t.Fatal("radio domain missing")
	}
	const tenantGUID = "44444444-4444-4444-8444-444444444444"
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{
			name:   "available",
			status: http.StatusOK,
			body:   `{"configurationGuid":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","available":true,"libraryName":"Crew Documents","defaultFolder":"Radio recordings","errorCode":null}`,
		},
		{
			name:   "not configured",
			status: http.StatusOK,
			body:   `{"configurationGuid":null,"available":false,"libraryName":null,"defaultFolder":null,"errorCode":"RADIO_SHAREPOINT_NOT_CONFIGURED"}`,
		},
		{
			name:   "membership or permission denied",
			status: http.StatusForbidden,
			body:   `{"code":"RADIO_ACCESS_DENIED"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixtureHome := t.TempDir()
			t.Setenv("HOME", fixtureHome)
			t.Setenv("USERPROFILE", fixtureHome)

			var requestCount atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				requestCount.Add(1)
				if request.Method != http.MethodGet || request.URL.Path != "/api/radio/backup-destinations/sharepoint" {
					t.Errorf("request = %s %s, want read-only tenant SharePoint availability", request.Method, request.URL.Path)
				}
				if request.Header.Get("X-Tenant-Guid") != tenantGUID || request.Header.Get("Authorization") != "Bearer radio-sharepoint-test-token" {
					t.Error("request must carry the authenticated, selected tenant context")
				}
				body, err := io.ReadAll(request.Body)
				if err != nil {
					t.Errorf("read request body: %v", err)
				}
				if request.URL.RawQuery != "" || len(body) != 0 {
					t.Errorf("availability must not send configuration or identity overrides: query=%q body=%q", request.URL.RawQuery, body)
				}
				response.Header().Set("Content-Type", "application/json")
				response.WriteHeader(tc.status)
				_, _ = response.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			if err := auth.SaveToken(&auth.TokenData{
				APIOrigin:   server.URL,
				AccessToken: "radio-sharepoint-test-token",
				ExpiresAt:   time.Now().Add(time.Hour),
				TenantGUID:  tenantGUID,
			}); err != nil {
				t.Fatalf("save test token: %v", err)
			}

			apiClient := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			command := buildDomainCommand(domain, func() (*client.APIClient, error) { return apiClient, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
			command.SilenceErrors = true
			command.SilenceUsage = true
			command.SetArgs([]string{"backup-sharepoint"})
			var runErr error
			stdout := captureRegistryStdout(t, func() { runErr = command.Execute() })
			if requestCount.Load() != 1 {
				t.Fatalf("requests = %d, want exactly one availability lookup", requestCount.Load())
			}
			if tc.status == http.StatusForbidden {
				var apiErr *clierrors.APIError
				if !errors.As(runErr, &apiErr) || apiErr.StatusCode != http.StatusForbidden || apiErr.Body != tc.body {
					t.Fatalf("denial was not preserved: %v", runErr)
				}
				if stdout != "" {
					t.Fatalf("denial printed a successful availability result: %q", stdout)
				}
				return
			}
			if runErr != nil {
				t.Fatalf("backup-sharepoint command: %v", runErr)
			}
			var got, want map[string]any
			if err := json.Unmarshal([]byte(stdout), &got); err != nil {
				t.Fatalf("availability stdout %q: %v", stdout, err)
			}
			if err := json.Unmarshal([]byte(tc.body), &want); err != nil {
				t.Fatalf("availability fixture: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("availability = %#v, want %#v", got, want)
			}
		})
	}
}

func TestRadioLicenseCommandsRequireExplicitTenantAndMemberGuids(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "radio-license" {
			domain = candidate
		}
	}
	if domain == nil {
		t.Fatal("radio-license domain missing")
	}
	if domain.APIPath != "/api/tenant" {
		t.Fatalf("unexpected API path %q", domain.APIPath)
	}
	expected := map[string]struct{ method, path, tool string }{
		"list":   {"GET", "{tenantGuid}/radio-license-users", "UteamupRadioLicenseList"},
		"assign": {"POST", "{tenantGuid}/radio-licenses/assign", "UteamupRadioLicenseAssign"},
		"remove": {"POST", "{tenantGuid}/radio-licenses/remove", "UteamupRadioLicenseRemove"},
	}
	for _, action := range domain.Actions {
		want, ok := expected[action.Name]
		if !ok {
			t.Fatalf("unexpected action %q", action.Name)
		}
		if action.HTTPMethod != want.method || action.RESTPath != want.path || action.ToolName != want.tool {
			t.Fatalf("%s: got %s %s %s", action.Name, action.HTTPMethod, action.RESTPath, action.ToolName)
		}
		flags := map[string]FlagDef{}
		for _, flag := range action.Flags {
			flags[flag.Name] = flag
		}
		if tenant := flags["tenant"]; !tenant.Required || tenant.BodyName != "tenantGuid" || tenant.Type != "non-empty-uuid" {
			t.Fatalf("%s must require the tenant GUID", action.Name)
		}
		if action.Name != "list" {
			if member := flags["user-guid"]; !member.Required || member.BodyName != "userGuid" || member.Type != "non-empty-uuid" {
				t.Fatalf("%s must require the member GUID", action.Name)
			}
		}
		delete(expected, action.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing actions: %v", expected)
	}
}

type radioRecordedRequest struct {
	method string
	path   string
	query  url.Values
	body   map[string]any
}

// runRadioCommand executes one radio CLI command against a TLS test API and
// records every request it sends.
func runRadioCommand(t *testing.T, status int, responseBody string, args ...string) ([]radioRecordedRequest, error) {
	t.Helper()
	domain := findDomain("radio")
	if domain == nil {
		t.Fatal("radio domain missing")
	}
	fixtureHome := t.TempDir()
	t.Setenv("HOME", fixtureHome)
	t.Setenv("USERPROFILE", fixtureHome)

	var requestsMu sync.Mutex
	var requests []radioRecordedRequest
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		recorded := radioRecordedRequest{method: request.Method, path: request.URL.Path, query: request.URL.Query()}
		raw, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &recorded.body); err != nil {
				t.Errorf("request body %q is not a JSON object: %v", raw, err)
			}
		}
		requestsMu.Lock()
		requests = append(requests, recorded)
		requestsMu.Unlock()
		if responseBody != "" {
			response.Header().Set("Content-Type", "application/json")
		}
		response.WriteHeader(status)
		if responseBody != "" {
			_, _ = response.Write([]byte(responseBody))
		}
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "radio-command-test-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "44444444-4444-4444-8444-444444444444",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	apiClient := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
	format := "json"
	command := buildDomainCommand(domain, func() (*client.APIClient, error) { return apiClient, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(args)
	var runErr error
	_ = captureRegistryStdout(t, func() { runErr = command.Execute() })

	requestsMu.Lock()
	defer requestsMu.Unlock()
	return append([]radioRecordedRequest(nil), requests...), runErr
}

func TestRadioRecordingsHistoryFiltersTravelAsQuery(t *testing.T) {
	const teamGUID = "55555555-5555-4555-8555-555555555555"
	for _, tc := range []struct {
		name string
		args []string
		want map[string]string
	}{
		{
			name: "defaults keep the existing own-recordings contract",
			args: []string{"recordings"},
			want: map[string]string{"skip": "0", "take": "50"},
		},
		{
			name: "every filter with scope All",
			args: []string{"recordings", "--scope", "All", "--from", "2026-10-01T00:00:00Z", "--to", "2026-10-02T00:00:00Z", "--team-guid", teamGUID, "--has-transcript", "--transcript-contains", "pump room", "--skip", "50", "--take", "25"},
			want: map[string]string{"scope": "All", "from": "2026-10-01T00:00:00Z", "to": "2026-10-02T00:00:00Z", "teamGuid": teamGUID, "hasTranscript": "true", "transcriptContains": "pump room", "skip": "50", "take": "25"},
		},
		{
			name: "explicit false keeps the tri-state has-transcript filter",
			args: []string{"recordings", "--scope", "Mine", "--has-transcript=false"},
			want: map[string]string{"scope": "Mine", "hasTranscript": "false", "skip": "0", "take": "50"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests, err := runRadioCommand(t, http.StatusOK, `[]`, tc.args...)
			if err != nil {
				t.Fatalf("recordings command: %v", err)
			}
			if len(requests) != 1 {
				t.Fatalf("requests = %d, want one history read", len(requests))
			}
			request := requests[0]
			if request.method != http.MethodGet || request.path != "/api/radio/recordings" || request.body != nil {
				t.Fatalf("request = %s %s body=%v, want a bodyless history read", request.method, request.path, request.body)
			}
			got := map[string]string{}
			for key, values := range request.query {
				if len(values) != 1 {
					t.Fatalf("query %s repeated: %v", key, values)
				}
				got[key] = values[0]
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("query = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRadioRecordingsRejectsInvalidFiltersBeforeRequest(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{name: "unknown scope", args: []string{"recordings", "--scope", "Everyone"}, want: "--scope"},
		{name: "scope is exact", args: []string{"recordings", "--scope", "all"}, want: "--scope"},
		{name: "empty team guid", args: []string{"recordings", "--team-guid", "00000000-0000-0000-0000-000000000000"}, want: "--team-guid"},
		{name: "malformed team guid", args: []string{"recordings", "--team-guid", "team-1"}, want: "--team-guid"},
		{name: "non-boolean has-transcript", args: []string{"recordings", "--has-transcript=maybe"}, want: "has-transcript"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests, err := runRadioCommand(t, http.StatusOK, `[]`, tc.args...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want rejection naming %s", err, tc.want)
			}
			if len(requests) != 0 {
				t.Fatalf("invalid filters reached the API: %+v", requests)
			}
		})
	}
}

func TestRadioAlwaysTranscribeSendsOnlyTheExplicitPreference(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want bool
	}{
		{name: "enable", args: []string{"always-transcribe", "--enabled"}, want: true},
		{name: "disable", args: []string{"always-transcribe", "--enabled=false"}, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests, err := runRadioCommand(t, http.StatusNoContent, "", tc.args...)
			if err != nil {
				t.Fatalf("always-transcribe command: %v", err)
			}
			if len(requests) != 1 {
				t.Fatalf("requests = %d, want one preference write", len(requests))
			}
			request := requests[0]
			if request.method != http.MethodPut || request.path != "/api/radio/me/always-transcribe" || len(request.query) != 0 {
				t.Fatalf("request = %s %s?%s, want the caller's preference route", request.method, request.path, request.query.Encode())
			}
			if !reflect.DeepEqual(request.body, map[string]any{"enabled": tc.want}) {
				t.Fatalf("body = %v, want only enabled=%v", request.body, tc.want)
			}
		})
	}

	t.Run("missing enabled is rejected", func(t *testing.T) {
		requests, err := runRadioCommand(t, http.StatusNoContent, "", "always-transcribe")
		if err == nil || !strings.Contains(err.Error(), "enabled") {
			t.Fatalf("error = %v, want the required --enabled flag", err)
		}
		if len(requests) != 0 {
			t.Fatalf("an implicit preference reached the API: %+v", requests)
		}
	})

	t.Run("AI.Use denial is preserved", func(t *testing.T) {
		const denial = `{"code":"RADIO_TRANSCRIPTION_NOT_ALLOWED"}`
		requests, err := runRadioCommand(t, http.StatusForbidden, denial, "always-transcribe", "--enabled")
		var apiErr *clierrors.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusForbidden || apiErr.Body != denial {
			t.Fatalf("denial was not preserved: %v", err)
		}
		if len(requests) != 1 {
			t.Fatalf("requests = %d, want one preference write", len(requests))
		}
	})
}

func TestRadioTranscribeRequestsOneRecordingWithoutBody(t *testing.T) {
	const recordingGUID = "66666666-6666-4666-8666-666666666666"
	requests, err := runRadioCommand(t, http.StatusOK, `{"recordingGuid":"`+recordingGUID+`","transcriptionStatus":"Pending"}`, "transcribe", recordingGUID)
	if err != nil {
		t.Fatalf("transcribe command: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want one transcription request", len(requests))
	}
	request := requests[0]
	if request.method != http.MethodPost || request.path != "/api/radio/recordings/"+recordingGUID+"/transcription" || len(request.query) != 0 || len(request.body) != 0 {
		t.Fatalf("request = %s %s?%s body=%v, want a bodyless transcription request", request.method, request.path, request.query.Encode(), request.body)
	}

	t.Run("missing audio conflict is preserved", func(t *testing.T) {
		const conflict = `{"code":"RADIO_RECORDING_NO_AUDIO"}`
		_, err := runRadioCommand(t, http.StatusConflict, conflict, "transcribe", recordingGUID)
		var apiErr *clierrors.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusConflict || apiErr.Body != conflict {
			t.Fatalf("conflict was not preserved: %v", err)
		}
	})

	for _, invalid := range []string{"00000000-0000-0000-0000-000000000000", "../recordings", "recording-1"} {
		t.Run("rejects "+invalid, func(t *testing.T) {
			requests, err := runRadioCommand(t, http.StatusOK, `{}`, "transcribe", invalid)
			if err == nil || !strings.Contains(err.Error(), "recordingGuid") {
				t.Fatalf("error = %v, want recordingGuid rejection", err)
			}
			if len(requests) != 0 {
				t.Fatalf("invalid recording GUID reached the API: %+v", requests)
			}
		})
	}
}
