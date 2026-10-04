package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func clientDeviceDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "client-device" {
			return domain
		}
	}
	t.Fatal("client-device domain missing")
	return nil
}

func TestClientDeviceContractUsesBackendToolsAndGuidOnlySelectors(t *testing.T) {
	domain := clientDeviceDomain(t)
	tools := map[string]string{"mine": "UteamupClientDeviceMine", "list": "UteamupClientDeviceList", "get": "UteamupClientDeviceGet", "revoke": "UteamupClientDeviceRevoke"}
	if domain.APIPath != "/api/device/enrolled" || len(domain.Actions) != len(tools) {
		t.Fatal("unexpected client device surface")
	}
	for _, action := range domain.Actions {
		if action.ToolName != tools[action.Name] {
			t.Errorf("unexpected tool: %+v", action)
		}
		if err := validateActionDefinition(action); err != nil {
			t.Errorf("%s: %v", action.Name, err)
		}
		if action.Name == "get" || action.Name == "revoke" {
			if len(action.Args) != 1 || action.Args[0].Type != "non-empty-uuid" || action.Args[0].BodyName != "deviceGuid" {
				t.Error("selector must be a non-empty GUID")
			}
		} else if len(action.Args) != 0 {
			t.Error("list must use server-resolved tenant/account")
		}
		for _, flag := range action.Flags {
			if strings.Contains(flag.Name, "tenant") || strings.Contains(flag.Name, "actor") || strings.Contains(flag.Name, "user") || strings.Contains(flag.Name, "wipe") {
				t.Errorf("unexpected authority or wipe argument: %s", flag.Name)
			}
		}
	}
}

func TestClientDeviceRevokeConfirmationAndGuidFailBeforeClientCreation(t *testing.T) {
	for _, arguments := range [][]string{
		{"revoke", "11111111-1111-4111-8111-111111111111", "--expected-enrollment-version", "1"},
		{"revoke", "11111111-1111-4111-8111-111111111111", "--expected-enrollment-version", "1", "--confirm=false"},
		{"revoke", "11111111-1111-4111-8111-111111111111", "--confirm"},
		{"revoke", "00000000-0000-0000-0000-000000000000", "--expected-enrollment-version", "1", "--confirm"},
		{"get", "not-a-guid"},
		{"get", "00000000-0000-0000-0000-000000000000"},
	} {
		created := false
		format := "json"
		command := buildDomainCommand(clientDeviceDomain(t), func() (*client.APIClient, error) {
			created = true
			return nil, nil
		}, logging.New(logging.LevelError), &format, &ExportConfig{})
		command.SilenceErrors = true
		command.SilenceUsage = true
		command.SetArgs(arguments)
		if err := command.Execute(); err == nil || created {
			t.Errorf("invalid request reached transport: %v, created=%v", err, created)
		}
	}
}

func TestClientDeviceTransportPreservesRevisionAndExcludesConfirmation(t *testing.T) {
	const tenantGuid = "55555555-5555-4555-8555-555555555555"
	const deviceGuid = "11111111-1111-4111-8111-111111111111"
	for _, scenario := range []struct{ action, revision string }{{"mine", ""}, {"list", ""}, {"get", ""}, {"revoke", "7"}, {"revoke", "9007199254740991"}} {
		action := scenario.action
		t.Run(action+scenario.revision, func(t *testing.T) {
			isolatedHome := t.TempDir()
			t.Setenv("HOME", isolatedHome)
			t.Setenv("USERPROFILE", isolatedHome)
			requests := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				requests++
				wantPath := "/api/device/enrolled"
				wantMethod := http.MethodGet
				if action == "mine" {
					wantPath += "/mine"
				} else if action == "get" || action == "revoke" {
					wantPath += "/" + deviceGuid
					if action == "revoke" {
						wantPath += "/revoke"
						wantMethod = http.MethodPost
					}
				}
				if request.Method != wantMethod || request.URL.Path != wantPath || request.Header.Get("X-Tenant-Guid") != tenantGuid {
					t.Errorf("unexpected method/path/tenant: %s %s", request.Method, request.URL)
				}
				if action == "mine" || action == "list" {
					query := request.URL.Query()
					if len(query) != 2 || query.Get("page") != "2" || query.Get("pageSize") != "3" {
						t.Errorf("unexpected paging query: %s", request.URL.RawQuery)
					}
				} else if request.URL.RawQuery != "" {
					t.Error("unexpected selector/authority query")
				}
				if action == "revoke" {
					var body map[string]any
					decoder := json.NewDecoder(request.Body)
					decoder.UseNumber()
					if err := decoder.Decode(&body); err != nil || len(body) != 1 || body["expectedEnrollmentVersion"] != json.Number(scenario.revision) {
						t.Errorf("revision changed or selector/confirmation leaked: %#v, %v", body, err)
					}
				}
				response.Header().Set("Content-Type", "application/json")
				_, _ = response.Write([]byte(`{"deviceGuid":"` + deviceGuid + `","enrollmentVersion":8,"offlineAuthorized":false,"capabilitiesVerified":false}`))
			}))
			t.Cleanup(server.Close)
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "device-contract-token", ExpiresAt: time.Now().Add(time.Hour), TenantGUID: tenantGuid}); err != nil {
				t.Fatal(err)
			}
			apiClient := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			command := buildDomainCommand(clientDeviceDomain(t), func() (*client.APIClient, error) { return apiClient, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
			command.SilenceErrors = true
			command.SilenceUsage = true
			arguments := []string{action}
			if action == "mine" || action == "list" {
				arguments = append(arguments, "--page", "2", "--page-size", "3")
			} else {
				arguments = append(arguments, deviceGuid)
				if action == "revoke" {
					arguments = append(arguments, "--expected-enrollment-version", scenario.revision, "--confirm")
				}
			}
			command.SetArgs(arguments)
			if err := command.Execute(); err != nil || requests != 1 {
				t.Fatalf("request did not complete once: err=%v, count=%d", err, requests)
			}
		})
	}
}
