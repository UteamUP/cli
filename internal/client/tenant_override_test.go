package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/logging"
)

const (
	loginTenantGUID = "b966b8c7-04a4-45d4-aa51-519ecf2ef13a"
	otherTenantGUID = "13f4751e-be00-4428-babd-f0d27e68bb40"
)

// tenantHeaderServer records the tenant headers of each request and answers both the REST and the
// MCP (/mcp JSON-RPC) shapes, against a saved login for loginTenantGUID with numeric id 42.
func tenantHeaderServer(t *testing.T) (*httptest.Server, *[]http.Header) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var seen []http.Header
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		seen = append(seen, request.Header.Clone())
		response.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/mcp" {
			_, _ = response.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"ok":true}}`))
			return
		}
		_, _ = response.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "access-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantID:    42,
		TenantGUID:  loginTenantGUID,
	}); err != nil {
		t.Fatal(err)
	}
	return server, &seen
}

func newTenantTestClient(server *httptest.Server) *APIClient {
	return NewAPIClient(server.URL, time.Second, true, RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
}

func TestTenantHeaders(t *testing.T) {
	cases := []struct {
		name       string
		override   string
		wantGUID   string
		wantLegacy string
	}{
		{"no override uses the login tenant", "", loginTenantGUID, "42"},
		{"blank override is ignored", "   ", loginTenantGUID, "42"},
		{"override naming the login tenant keeps its numeric id", "B966B8C7-04A4-45D4-AA51-519ECF2EF13A", "B966B8C7-04A4-45D4-AA51-519ECF2EF13A", "42"},
		{"override naming another tenant drops the stale numeric id", otherTenantGUID, otherTenantGUID, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, seen := tenantHeaderServer(t)
			apiClient := newTenantTestClient(server).WithTenant(tc.override)

			if _, err := apiClient.CallREST(context.Background(), "GET", "/api/meter-readings", nil, nil, "list"); err != nil {
				t.Fatalf("CallREST: %v", err)
			}
			if _, err := apiClient.CallTool(context.Background(), "UteamupMeterreadingList", map[string]any{}); err != nil {
				t.Fatalf("CallTool: %v", err)
			}

			if len(*seen) != 2 {
				t.Fatalf("requests = %d, want 2", len(*seen))
			}
			for i, headers := range *seen {
				if got := headers.Get("X-Tenant-Guid"); got != tc.wantGUID {
					t.Errorf("request %d X-Tenant-Guid = %q, want %q", i, got, tc.wantGUID)
				}
			}
			if got := (*seen)[0].Get("X-Tenant-ID"); got != tc.wantLegacy {
				t.Errorf("REST X-Tenant-ID = %q, want %q", got, tc.wantLegacy)
			}
		})
	}
}
