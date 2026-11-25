package registry

import (
	"encoding/json"
	"github.com/uteamup/cli/internal/client"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestIntegrationStatusCommandsUseMCPWithoutIdentityOverrides(t *testing.T) {
	domain := findDomain("integrations")
	if domain == nil || len(domain.Actions) != 3 {
		t.Fatal("expected three integration status reads")
	}
	expected := map[string]string{
		"storage":          "uteamup_storage_status_get",
		"sharepoint":       "uteamup_sharepoint_status_get",
		"business-central": "uteamup_business_central_status_get",
	}
	for _, action := range domain.Actions {
		tool, ok := expected[action.Name]
		if !ok || action.ToolName != tool || !action.MCPOnly {
			t.Fatalf("wrong MCP route: %+v", action)
		}
		if len(action.Args) != 0 || len(action.Flags) != 0 {
			t.Fatalf("status read accepts identity/config overrides: %+v", action)
		}
	}
}

func TestIntegrationStatusTransportUsesAuthenticatedMCPRead(t *testing.T) {
	for _, action := range findDomain("integrations").Actions {
		t.Run(action.Name, func(t *testing.T) {
			var calls atomic.Int32
			apiClient := projectTransportClient(t, "apikey", client.RetryOptions{}, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != http.MethodPost || r.URL.Path != "/mcp" || r.URL.RawQuery != "" {
					t.Errorf("unexpected route: %s %s", r.Method, r.URL)
				}
				var rpc struct {
					Method string                `json:"method"`
					Params client.ToolCallParams `json:"params"`
				}
				if err := json.NewDecoder(r.Body).Decode(&rpc); err != nil {
					t.Fatal(err)
				}
				if rpc.Method != "tools/call" || rpc.Params.Name != action.ToolName || len(rpc.Params.Arguments) != 0 {
					t.Errorf("unexpected MCP call: %+v", rpc)
				}
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"status":"not_configured","healthStatus":"not_checked"}}`))
			})
			if err := executeProjectTransport(t, apiClient, "integrations", []string{action.Name}); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 1 {
				t.Fatalf("calls = %d, want one", calls.Load())
			}
		})
	}
}
