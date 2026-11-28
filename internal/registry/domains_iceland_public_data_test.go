package registry

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/uteamup/cli/internal/client"
)

const icelandSourceKey = "is-hms-stadfangaskra"
const icelandUpdateJSON = `{"expectedRevision":"11111111-1111-4111-8111-111111111111","enabled":false,"refreshHours":24}`

func TestIcelandPublicDataContract(t *testing.T) {
	domain := findDomain("iceland-public-data")
	if domain == nil || len(domain.Actions) != 3 {
		t.Fatal("expected list, update and refresh for Icelandic public data")
	}
	for _, tc := range []struct{ name, method, path, tool string }{
		{"list", "GET", "/api/globaladmin/iceland-data", "uteamup_iceland_public_data_list"},
		{"update", "PUT", "/api/globaladmin/iceland-data/" + icelandSourceKey, "uteamup_iceland_public_data_update"},
		{"refresh", "POST", "/api/globaladmin/iceland-data/" + icelandSourceKey + "/refresh", "uteamup_iceland_public_data_refresh"},
	} {
		action := findAction(domain, tc.name)
		if action == nil {
			t.Fatalf("missing %s", tc.name)
		}
		path, _ := buildRESTPath(domain, *action, map[string]any{"sourceKey": icelandSourceKey})
		if path != tc.path || action.HTTPMethod != tc.method || mcpToolName(action.ToolName) != tc.tool || action.MCPOnly || !action.RejectExtraArgs {
			t.Fatalf("unexpected contract for %s: %s %+v", tc.name, path, action)
		}
		if tc.name != "list" && (len(action.Args) != 1 || !reflect.DeepEqual(action.Args[0].AllowedValues, []string{icelandSourceKey})) {
			t.Fatalf("mutation must target the supported source only: %+v", action.Args)
		}
	}
	update := findAction(domain, "update")
	if len(update.Flags) != 1 || !update.Flags[0].Required || !update.Flags[0].RootJSONObjectFile {
		t.Fatal("update requires reviewed root JSON")
	}
}

func TestIcelandPublicDataRESTTransport(t *testing.T) {
	for _, tc := range []struct {
		name, method, path string
		args               []string
	}{
		{"list", "GET", "/api/globaladmin/iceland-data", []string{"list"}},
		{"update", "PUT", "/api/globaladmin/iceland-data/" + icelandSourceKey, []string{"update", icelandSourceKey, "--from-json"}},
		{"refresh", "POST", "/api/globaladmin/iceland-data/" + icelandSourceKey + "/refresh", []string{"refresh", icelandSourceKey}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := tc.args
			if tc.name == "update" {
				args = append(args, writeRegistryJSONFixture(t, icelandUpdateJSON))
			}
			requests, err := runRegisteredDomainCommand(t, "iceland-public-data", args...)
			if err != nil || len(requests) != 1 {
				t.Fatalf("transport: %v %+v", err, requests)
			}
			request := requests[0]
			if request.method != tc.method || request.path != tc.path || len(request.query) != 0 || request.header.Get("Authorization") == "" {
				t.Fatalf("incorrect authenticated route: %+v", request)
			}
			if tc.name == "update" {
				var want map[string]any
				if err := json.Unmarshal([]byte(icelandUpdateJSON), &want); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(request.body, want) {
					t.Fatalf("root fields changed: %+v", request.body)
				}
			} else if len(request.body) != 0 {
				t.Fatalf("unexpected body: %+v", request.body)
			}
		})
	}
}

func TestIcelandPublicDataRejectsUnsafeInput(t *testing.T) {
	for _, args := range [][]string{
		{"refresh", "../hosting"}, {"refresh", "metadata-only-source"}, {"list", "extra"},
		{"update", icelandSourceKey},
		{"update", icelandSourceKey, "--from-json", writeRegistryJSONFixture(t, `[]`)},
		{"update", icelandSourceKey, "--from-json", writeRegistryJSONFixture(t, `{"enabled":true,"enabled":false}`)},
	} {
		requests, err := runRegisteredDomainCommand(t, "iceland-public-data", args...)
		if err == nil || len(requests) != 0 {
			t.Fatalf("unsafe input sent: %v %v %+v", args, err, requests)
		}
	}
}

func TestIcelandPublicDataPreservesAuthorityAndConfigurationFailures(t *testing.T) {
	for _, tc := range []struct {
		status        int
		message, body string
	}{
		{403, "Configured Global Admin required", icelandUpdateJSON},
		{409, "Configuration revision changed", icelandUpdateJSON},
		{400, "Unknown configuration member downloadUrl", `{"expectedRevision":"11111111-1111-4111-8111-111111111111","enabled":true,"refreshHours":24,"downloadUrl":"https://invalid.test/"}`},
		{400, "refreshHours must be 24..720", `{"expectedRevision":"11111111-1111-4111-8111-111111111111","enabled":true,"refreshHours":1}`},
	} {
		t.Run(tc.message, func(t *testing.T) {
			calls := 0
			apiClient := projectTransportClient(t, "login", client.RetryOptions{}, func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(tc.status)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": tc.message})
			})
			err := executeProjectTransport(t, apiClient, "iceland-public-data", []string{"update", icelandSourceKey, "--from-json", writeRegistryJSONFixture(t, tc.body)})
			if err == nil || !strings.Contains(err.Error(), tc.message) || calls != 1 {
				t.Fatalf("backend refusal lost or retried: calls=%d err=%v", calls, err)
			}
		})
	}
}
