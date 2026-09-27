package registry

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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
	expected := map[string]string{"backup-sharepoint": "backup-destinations/sharepoint", "validate": "validate", "stop-pilot": "complimentary/revoke", "transcript": "channels/{channelGuid}/transmissions/{transmissionGuid}/transcript", "transmissions": "channels/{channelGuid}/transmissions", "status": "environment", "policy": "policy", "channels": "history-channels", "recordings": "recordings", "get": "recordings/{recordingGuid}", "playback": "recordings/{recordingGuid}/playback"}
	for _, action := range domain.Actions {
		if action.RESTPath != expected[action.Name] {
			t.Errorf("wrong route for %s: %s", action.Name, action.RESTPath)
		}
		if !strings.HasPrefix(action.ToolName, "UteamupRadio") {
			t.Errorf("missing MCP equivalent for %s", action.Name)
		}
		for _, arg := range action.Args {
			if (arg.Name != "recordingGuid" && arg.Name != "channelGuid" && arg.Name != "transmissionGuid") || arg.Type != "non-empty-uuid" {
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
			if err := auth.SaveToken(&auth.TokenData{
				AccessToken: "radio-sharepoint-test-token",
				ExpiresAt:   time.Now().Add(time.Hour),
				TenantGUID:  tenantGUID,
			}); err != nil {
				t.Fatalf("save test token: %v", err)
			}

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
