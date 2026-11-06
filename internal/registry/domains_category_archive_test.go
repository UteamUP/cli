package registry

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

// runCategoryCommand executes a category action against a TLS test server and returns the request it received
// (nil when the command failed before sending), plus the command error.
func runCategoryCommand(t *testing.T, cliArgs ...string) (*http.Request, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var received *http.Request
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		received = request.Clone(request.Context())
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"guid":"11111111-1111-4111-8111-111111111111","name":"Electrical","archivedAtUtc":null}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "category-archive-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}
	apiClient := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
	format := "json"
	command := buildDomainCommand(findDomainByName(t, "category"), func() (*client.APIClient, error) { return apiClient, nil },
		logging.New(logging.LevelError), &format, &ExportConfig{})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(cliArgs)
	return received, command.Execute()
}

func TestCategoryArchiveAndRestorePostToTheGuidRoutes(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for _, action := range []string{"archive", "restore"} {
		request, err := runCategoryCommand(t, action, guid)
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if request.Method != http.MethodPost || request.URL.Path != "/api/category/by-guid/"+guid+"/"+action {
			t.Fatalf("unexpected %s route: %s %s", action, request.Method, request.URL)
		}
		if request.URL.RawQuery != "" {
			t.Fatalf("%s leaked arguments into the query: %s", action, request.URL)
		}
	}
}

func TestCategoryArchiveRefusesBadGuidsAndExtraArgsBeforeSending(t *testing.T) {
	cases := [][]string{
		{"archive", "00000000-0000-0000-0000-000000000000"},
		{"archive", "not-a-guid"},
		{"restore"},
		{"restore", "11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"},
	}
	for _, args := range cases {
		request, err := runCategoryCommand(t, args...)
		if err == nil || request != nil {
			t.Fatalf("%v must be refused before any request; err=%v request=%v", args, err, request)
		}
	}
}

func TestCategoryListShowsArchivedOnlyWhenAsked(t *testing.T) {
	request, err := runCategoryCommand(t, "list", "--archived")
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != http.MethodGet || request.URL.Path != "/api/category" || request.URL.Query().Get("archived") != "true" {
		t.Fatalf("unexpected archived list request: %s %s", request.Method, request.URL)
	}

	request, err = runCategoryCommand(t, "list")
	if err != nil {
		t.Fatal(err)
	}
	if request.URL.Query().Get("archived") == "true" {
		t.Fatalf("the default list must hide archived categories: %s", request.URL)
	}
}

func TestOnlyCategoriesGainArchive(t *testing.T) {
	for _, action := range findDomainByName(t, "tag").Actions {
		if action.Name == "archive" || action.Name == "restore" {
			t.Fatalf("tag has no archive, but registers %q", action.Name)
		}
	}
}
