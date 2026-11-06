package registry

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

// runInventoryCommand executes an inventory action against a TLS test server and returns the request it
// received (nil when the command failed before sending), plus the command error.
func runInventoryCommand(t *testing.T, cliArgs ...string) (*http.Request, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var received *http.Request
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		received = request.Clone(request.Context())
		received.Body = io.NopCloser(bytes.NewReader(body))
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"type":"part","guid":"11111111-1111-4111-8111-111111111111","archivedAtUtc":null}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "inventory-archive-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	var inventory *Domain
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "inventory" {
			inventory = domain
		}
	}
	if inventory == nil {
		t.Fatal("inventory domain not registered")
	}
	apiClient := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
	format := "json"
	command := buildDomainCommand(inventory, func() (*client.APIClient, error) { return apiClient, nil },
		logging.New(logging.LevelError), &format, &ExportConfig{})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(cliArgs)
	return received, command.Execute()
}

func TestInventoryArchivePostsToTheTypedGuidRoute(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for _, itemType := range []string{"part", "tool", "chemical"} {
		request, err := runInventoryCommand(t, "archive", itemType, guid)
		if err != nil {
			t.Fatalf("archive %s: %v", itemType, err)
		}
		if request.Method != http.MethodPost || request.URL.Path != "/api/inventory/"+itemType+"/by-guid/"+guid+"/archive" {
			t.Fatalf("unexpected archive route: %s %s", request.Method, request.URL)
		}
		if request.URL.RawQuery != "" {
			t.Fatalf("archive leaked arguments into the query: %s", request.URL)
		}
	}
}

func TestInventoryRestoreCoversAssets(t *testing.T) {
	const guid = "22222222-2222-4222-8222-222222222222"
	request, err := runInventoryCommand(t, "restore", "asset", guid)
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != http.MethodPost || request.URL.Path != "/api/inventory/asset/by-guid/"+guid+"/restore" {
		t.Fatalf("unexpected restore route: %s %s", request.Method, request.URL)
	}
}

func TestInventoryArchiveRefusesAssetsAndBadGuidsBeforeSending(t *testing.T) {
	cases := [][]string{
		{"archive", "asset", "11111111-1111-4111-8111-111111111111"},
		{"archive", "part", "00000000-0000-0000-0000-000000000000"},
		{"archive", "part", "not-a-guid"},
		{"restore", "vendor", "11111111-1111-4111-8111-111111111111"},
	}
	for _, args := range cases {
		request, err := runInventoryCommand(t, args...)
		if err == nil || request != nil {
			t.Fatalf("%v must be refused before any request; err=%v request=%v", args, err, request)
		}
	}
}

func TestInventoryBulkDeleteNoLongerOffersAssets(t *testing.T) {
	var inventory *Domain
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "inventory" {
			inventory = domain
		}
	}
	for _, action := range inventory.Actions {
		if action.Name != "bulk-delete" {
			continue
		}
		for _, flag := range action.Flags {
			if flag.Name == "scope" && !bytes.Contains([]byte(flag.Description), []byte("Assets are never deleted")) {
				t.Fatalf("bulk-delete scope must say assets are archived, got %q", flag.Description)
			}
		}
		return
	}
	t.Fatal("inventory bulk-delete action not registered")
}
