package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/logging"
)

func callToolClient(t *testing.T, authMethod string, calls *atomic.Int32) *APIClient {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		if request.Method != http.MethodPost || request.URL.Path != "/mcp" {
			t.Errorf("request = %s %s, want POST /mcp", request.Method, request.URL.Path)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"items":[]}}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "call-tool-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		AuthMethod:  authMethod,
		TenantGUID:  "b966b8c7-04a4-45d4-aa51-519ecf2ef13a",
	}); err != nil {
		t.Fatal(err)
	}
	return NewAPIClient(server.URL, time.Second, true, RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
}

// The backend resolves the /mcp tenant only from API-key token claims, so an
// email/password session must be refused locally with an actionable message
// instead of being sent to /mcp for a 401 about API keys (bug 4499a609).
func TestCallToolRefusesHumanOrUnknownSessionBeforeSending(t *testing.T) {
	for _, method := range []string{"login", "saml", "unknown", ""} {
		t.Run(method, func(t *testing.T) { testCallToolRefusesHumanSession(t, method) })
	}
}

func testCallToolRefusesHumanSession(t *testing.T, method string) {
	var calls atomic.Int32
	apiClient := callToolClient(t, method, &calls)

	_, err := apiClient.CallTool(context.Background(), "UteamupAssetTypeList", map[string]any{})

	var authErr *clierrors.AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("error = %v, want AuthError", err)
	}
	for _, want := range []string{"UteamupAssetTypeList", "human sign-in", "ut login --api-key-auth"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err.Error(), want)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("HTTP calls = %d, want none for a session /mcp cannot serve", calls.Load())
	}
}

func TestCallToolSendsAPIKeySessionToMCP(t *testing.T) {
	var calls atomic.Int32
	apiClient := callToolClient(t, "apikey", &calls)

	result, err := apiClient.CallTool(context.Background(), "UteamupAssetTypeList", map[string]any{})
	if err != nil {
		t.Fatalf("API-key session must reach /mcp: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("HTTP calls = %d, want one tools/call", calls.Load())
	}
	if !strings.Contains(string(result), `"items"`) {
		t.Fatalf("result = %s", result)
	}
}
