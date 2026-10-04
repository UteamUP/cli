package registry

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
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

func clientSecurityDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "client-security" {
			return domain
		}
	}
	t.Fatal("client-security domain missing")
	return nil
}

func TestClientSecurityPolicyUsesActiveTenantAndBackendOwnedTools(t *testing.T) {
	domain := clientSecurityDomain(t)
	expected := map[string]string{
		"get":       "UteamupTenantClientSecurityPolicyGet",
		"effective": "UteamupTenantClientSecurityPolicyEffective",
		"history":   "UteamupTenantClientSecurityPolicyHistory",
		"set":       "UteamupTenantClientSecurityPolicySet",
	}
	if domain.APIPath != "/api/tenantpolicies/client-security" || len(domain.Actions) != len(expected) {
		t.Fatal("unexpected client policy API surface")
	}
	for _, action := range domain.Actions {
		if action.ToolName != expected[action.Name] || len(action.Args) != 0 {
			t.Errorf("unexpected tool or public identity override: %+v", action)
		}
		if err := validateActionDefinition(action); err != nil {
			t.Errorf("%s definition: %v", action.Name, err)
		}
		if action.Name != "set" && action.Name != "history" && len(action.Flags) != 0 {
			t.Errorf("%s must use authenticated active-tenant scope", action.Name)
		}
	}
}

func TestClientSecurityHistoryIsOneScopedPageWithExactQueryFlags(t *testing.T) {
	domain := clientSecurityDomain(t)
	var history *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "history" {
			history = &domain.Actions[index]
		}
	}
	if history == nil {
		t.Fatal("client-security history action missing")
	}
	if history.HTTPMethod != http.MethodGet || history.RESTPath != "history" ||
		history.UseDomainBasePath || history.MCPOnly || len(history.Args) != 0 ||
		history.PollWaitFlag != "" || len(history.Flags) != 2 || !history.RejectExtraArgs {
		t.Fatalf("history must read one scoped REST page: %+v", history)
	}
	wantFlags := map[string]FlagDef{
		"cursor":    {Type: "string", BodyName: "cursor", QueryName: "cursor"},
		"page-size": {Type: "int", BodyName: "pageSize", QueryName: "pageSize", Default: 25},
	}
	for _, flag := range history.Flags {
		want, exists := wantFlags[flag.Name]
		if !exists || flag.Type != want.Type || flag.BodyName != want.BodyName ||
			flag.QueryName != want.QueryName || !reflect.DeepEqual(flag.Default, want.Default) ||
			flag.Required || flag.LocalOnly || flag.MustBeTrue || flag.HeaderName != "" {
			t.Errorf("unexpected history flag or authority/confirmation override: %+v", flag)
		}
		delete(wantFlags, flag.Name)
	}
	if len(wantFlags) != 0 {
		t.Fatalf("missing history query flags: %v", wantFlags)
	}
}

func TestClientSecurityHistoryRejectsIdentityOverridesBeforeClientCreation(t *testing.T) {
	for _, arguments := range [][]string{
		{"history", "--tenant-guid", "55555555-5555-4555-8555-555555555555"},
		{"history", "--actor-guid", "55555555-5555-4555-8555-555555555555"},
		{"history", "--user-id", "1"},
		{"history", "55555555-5555-4555-8555-555555555555"},
		{"history", "--page-size", "not-an-integer"},
	} {
		created := false
		format := "json"
		command := buildDomainCommand(clientSecurityDomain(t), func() (*client.APIClient, error) {
			created = true
			return nil, nil
		}, logging.New(logging.LevelError), &format, &ExportConfig{})
		command.SilenceErrors = true
		command.SilenceUsage = true
		command.SetArgs(arguments)
		if err := command.Execute(); err == nil || created {
			t.Errorf("identity override or malformed input reached transport: %v, created=%v", err, created)
		}
	}
}

func TestClientSecurityHistoryExtraArgumentRejectionIsOptIn(t *testing.T) {
	for _, reject := range []bool{false, true} {
		created := false
		transportReached := errors.New("client factory reached")
		format := "json"
		domain := &Domain{Name: "extra-argument-contract", Actions: []Action{{
			Name: "history", HTTPMethod: http.MethodGet, RejectExtraArgs: reject,
		}}}
		command := buildDomainCommand(domain, func() (*client.APIClient, error) {
			created = true
			return nil, transportReached
		}, logging.New(logging.LevelError), &format, &ExportConfig{})
		command.SilenceErrors = true
		command.SilenceUsage = true
		command.SetArgs([]string{"history", "unexpected"})
		err := command.Execute()
		if reject {
			if err == nil || created || errors.Is(err, transportReached) {
				t.Errorf("opted-in action reached transport: err=%v, created=%v", err, created)
			}
		} else if !created || !errors.Is(err, transportReached) {
			t.Errorf("existing action behavior changed: err=%v, created=%v", err, created)
		}
	}
}

func TestClientSecurityHistoryTransportUsesActiveTenantAndPreservesOnePage(t *testing.T) {
	const tenantGuid = "55555555-5555-4555-8555-555555555555"
	const timestamp = "2026-10-04T01:02:03.1234567Z"
	policy := map[string]any{
		"tenantGuid": tenantGuid, "policyVersion": 1, "updatedAt": nil,
		"offlineEnabled": true, "maxOfflineDays": 60,
		"localPassphraseAllowed": true, "osAuthenticationAllowed": true,
		"lockOnOsLock": true, "lockOnSleep": true, "lockOnSessionChange": true,
		"lockOnShutdown": true, "idleLockMinutes": 15, "allowPwaLifecycleFallback": true,
		"alwaysLockOnStartup": true,
	}
	after := make(map[string]any, len(policy))
	for key, value := range policy {
		after[key] = value
	}
	after["policyVersion"] = 2
	after["updatedAt"] = timestamp
	fullPage, err := json.Marshal(map[string]any{
		"tenantGuid": tenantGuid,
		"items": []any{map[string]any{
			"guid": "11111111-1111-4111-8111-111111111111", "occurredAtUtc": timestamp,
			"actorName": nil, "before": policy, "after": after,
		}},
		"nextCursor": nil, "skippedCount": 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	const specialCursor = "opaque+/=?&% #🚢"
	for _, scenario := range []struct {
		name      string
		arguments []string
		query     url.Values
		status    int
		body      string
	}{
		{"defaults", nil, url.Values{"pageSize": {"25"}}, http.StatusOK, string(fullPage)},
		{"special cursor", []string{"--cursor", specialCursor, "--page-size", "100"},
			url.Values{"cursor": {specialCursor}, "pageSize": {"100"}}, http.StatusOK, string(fullPage)},
		{"empty continued page", []string{"--page-size", "3"}, url.Values{"pageSize": {"3"}}, http.StatusOK,
			`{"tenantGuid":"` + tenantGuid + `","items":[],"nextCursor":"opaque-next","skippedCount":3}`},
		{"empty terminal page", nil, url.Values{"pageSize": {"25"}}, http.StatusOK,
			`{"tenantGuid":"` + tenantGuid + `","items":[],"nextCursor":null,"skippedCount":0}`},
		{"server rejects zero page size", []string{"--page-size", "0"}, url.Values{"pageSize": {"0"}}, http.StatusBadRequest,
			`{"message":"PageSize must be between 1 and 100."}`},
		{"server rejects excess page size", []string{"--page-size", "101"}, url.Values{"pageSize": {"101"}}, http.StatusBadRequest,
			`{"message":"PageSize must be between 1 and 100."}`},
		{"server rejects long cursor", []string{"--cursor", strings.Repeat("a", 1025)},
			url.Values{"cursor": {strings.Repeat("a", 1025)}, "pageSize": {"25"}}, http.StatusBadRequest,
			`{"message":"Invalid client policy history cursor."}`},
		{"permission or membership denied", nil, url.Values{"pageSize": {"25"}}, http.StatusForbidden,
			`{"message":"Tenant.Update and current membership are required."}`},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			isolatedHome := t.TempDir()
			t.Setenv("HOME", isolatedHome)
			t.Setenv("USERPROFILE", isolatedHome)
			var requests atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				requests.Add(1)
				if request.Method != http.MethodGet || request.URL.Path != "/api/tenantpolicies/client-security/history" ||
					request.URL.RawQuery != scenario.query.Encode() {
					t.Errorf("unexpected history method/path/query: %s %s", request.Method, request.URL)
				}
				if request.Header.Get("X-Tenant-Guid") != tenantGuid ||
					request.Header.Get("Authorization") != "Bearer client-history-contract-token" {
					t.Error("history must use the authenticated active tenant")
				}
				body, err := io.ReadAll(request.Body)
				if err != nil || len(body) != 0 {
					t.Errorf("read-only history sent a body: %q, %v", body, err)
				}
				response.Header().Set("Content-Type", "application/json")
				response.WriteHeader(scenario.status)
				_, _ = response.Write([]byte(scenario.body))
			}))
			t.Cleanup(server.Close)
			if err := auth.SaveToken(&auth.TokenData{
				APIOrigin: server.URL, AccessToken: "client-history-contract-token",
				ExpiresAt: time.Now().Add(time.Hour), TenantGUID: tenantGuid,
			}); err != nil {
				t.Fatal(err)
			}
			apiClient := client.NewAPIClient(server.URL, time.Second, true,
				client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			command := buildDomainCommand(clientSecurityDomain(t), func() (*client.APIClient, error) {
				return apiClient, nil
			}, logging.New(logging.LevelError), &format, &ExportConfig{})
			command.SilenceErrors = true
			command.SilenceUsage = true
			command.SetArgs(append([]string{"history"}, scenario.arguments...))
			var runErr error
			stdout := captureRegistryStdout(t, func() { runErr = command.Execute() })
			if requests.Load() != 1 {
				t.Fatalf("history must read exactly one page: requests=%d", requests.Load())
			}
			if scenario.status != http.StatusOK {
				var apiErr *clierrors.APIError
				if !errors.As(runErr, &apiErr) || apiErr.StatusCode != scenario.status || apiErr.Body != scenario.body {
					t.Fatalf("server error was not preserved: %v", runErr)
				}
				if stdout != "" {
					t.Fatalf("failed history printed a successful page: %q", stdout)
				}
				return
			}
			if runErr != nil {
				t.Fatal(runErr)
			}
			var actual, expected map[string]any
			if err := json.Unmarshal([]byte(stdout), &actual); err != nil {
				t.Fatalf("decode printed history: %v", err)
			}
			if err := json.Unmarshal([]byte(scenario.body), &expected); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("history fields/nulls/precision/continuation changed: %#v", actual)
			}
		})
	}
}

func TestClientSecurityConfirmationFailsBeforeCreatingClient(t *testing.T) {
	for _, confirmation := range []string{"", "--confirm=false"} {
		t.Run(confirmation, func(t *testing.T) {
			created := false
			format := "json"
			command := buildDomainCommand(clientSecurityDomain(t), func() (*client.APIClient, error) {
				created = true
				return nil, nil
			}, logging.New(logging.LevelError), &format, &ExportConfig{})
			command.SilenceErrors = true
			command.SilenceUsage = true
			arguments := []string{"set", "--file", filepath.Join(t.TempDir(), "unused.json")}
			if confirmation != "" {
				arguments = append(arguments, confirmation)
			}
			command.SetArgs(arguments)
			if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "confirm") {
				t.Fatalf("policy mutation accepted without explicit confirmation: %v", err)
			}
			if created {
				t.Fatal("client created before confirmation validation")
			}
		})
	}
}

func TestClientSecurityTransportPreservesVersionAndExcludesLocalConfirmation(t *testing.T) {
	const tenantGuid = "55555555-5555-4555-8555-555555555555"
	for _, action := range []string{"get", "effective", "set"} {
		t.Run(action, func(t *testing.T) {
			policy := map[string]any{
				"expectedPolicyVersion": float64(7), "offlineEnabled": true, "maxOfflineDays": float64(60),
				"localPassphraseAllowed": true, "osAuthenticationAllowed": false,
				"lockOnOsLock": true, "lockOnSleep": true, "lockOnSessionChange": true,
				"lockOnShutdown": true, "idleLockMinutes": float64(15), "allowPwaLifecycleFallback": true,
			}
			encoded, err := json.Marshal(policy)
			if err != nil {
				t.Fatal(err)
			}
			requestFile := writeRegistryJSONFixture(t, string(encoded))
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			requests := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				requests++
				wantPath := "/api/tenantpolicies/client-security"
				wantMethod := http.MethodGet
				if action == "effective" {
					wantPath += "/effective"
				} else if action == "set" {
					wantMethod = http.MethodPut
				}
				if request.Method != wantMethod || request.URL.Path != wantPath || request.URL.RawQuery != "" {
					t.Errorf("unexpected request: %s %s", request.Method, request.URL.String())
				}
				if request.Header.Get("X-Tenant-Guid") != tenantGuid {
					t.Errorf("request does not use active tenant GUID")
				}
				if action == "set" {
					var actual map[string]any
					if err := json.NewDecoder(request.Body).Decode(&actual); err != nil {
						t.Errorf("decode request: %v", err)
					} else if !reflect.DeepEqual(actual, policy) {
						t.Errorf("reviewed policy changed or local confirmation leaked: %#v", actual)
					}
				}
				response.Header().Set("Content-Type", "application/json")
				_, _ = response.Write([]byte(`{"tenantGuid":"` + tenantGuid + `","policyVersion":8}`))
			}))
			t.Cleanup(server.Close)
			if err := auth.SaveToken(&auth.TokenData{
				APIOrigin: server.URL, AccessToken: "client-policy-contract-token",
				ExpiresAt: time.Now().Add(time.Hour), TenantGUID: tenantGuid,
			}); err != nil {
				t.Fatal(err)
			}
			apiClient := client.NewAPIClient(server.URL, time.Second, true,
				client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			command := buildDomainCommand(clientSecurityDomain(t), func() (*client.APIClient, error) {
				return apiClient, nil
			}, logging.New(logging.LevelError), &format, &ExportConfig{})
			command.SilenceErrors = true
			command.SilenceUsage = true
			arguments := []string{action}
			if action == "set" {
				arguments = append(arguments, "--file", requestFile, "--confirm")
			}
			command.SetArgs(arguments)
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Errorf("request count = %d, want 1", requests)
			}
		})
	}
}
