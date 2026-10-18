package registry

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

// The fall-through defect: an action whose name is not a standard verb and that declares no
// RESTPath, RESTBasePath or UseDomainBasePath is sent to the bare domain path by
// buildRESTPath, with the method defaulting to GET. A "read one thing" action then silently
// returns the domain's list, and a POST lands on the domain's create endpoint. Each action
// below was checked against its controller's [Route]/[Http*] template.

var standardRESTVerbs = map[string]bool{
	"list": true, "get": true, "create": true, "update": true, "delete": true, "search": true,
}

func TestFallThroughActionsResolveToRealControllerRoutes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		domain, action, method string
		args                   map[string]any
		path                   string
	}{
		{"asset-failure", "by-asset", "GET", map[string]any{"assetGuid": "a"}, "/api/assetfailure/asset/by-guid/a"},

		{"codecatalog", "deactivate-by-guid", "DELETE", map[string]any{"guid": "g"}, "/api/codecatalog/entries/by-guid/g"},
		{"codecatalog", "remove-asset-assignment", "DELETE", map[string]any{"assetGuid": "a"}, "/api/asset/by-guid/a/codeassignment"},
		{"codecatalog", "assign-asset", "POST", map[string]any{"assetGuid": "a"}, "/api/asset/by-guid/a/codeassignment"},
		{"codecatalog", "history", "GET", map[string]any{"codeGuid": "c"}, "/api/codecatalog/c/history"},

		{"fuel-transaction", "summary", "GET", map[string]any{"assetGuid": "a"}, "/api/fueltransaction/asset/a/summary"},
		{"fuel-transaction", "efficiency", "GET", map[string]any{"assetGuid": "a"}, "/api/fueltransaction/asset/a/efficiency"},

		// FleetAssetContactController is routed api/v1/assets/{assetGuid}/contacts.
		{"fleet-contact", "list", "GET", map[string]any{"assetExternalGuid": "a"}, "/api/v1/assets/a/contacts"},
		{"fleet-contact", "add", "POST", map[string]any{"assetExternalGuid": "a"}, "/api/v1/assets/a/contacts"},
		{"fleet-contact", "delete", "DELETE", map[string]any{"assetExternalGuid": "a", "associationExternalGuid": "s"}, "/api/v1/assets/a/contacts/s"},

		{"labour-ai-offers", "compare", "POST", nil, "/api/labour-marketplace/ai/offer-comparison"},

		{"meter-schedule", "overdue", "GET", nil, "/api/meter-reading-schedules"},
		{"meter-schedule", "record-workorder", "POST", map[string]any{"workorderGuid": "w"}, "/api/workorders/by-guid/w/meter-readings"},

		{"salesbooking", "slots", "GET", nil, "/api/salesbooking/available-slots"},
		{"salesbooking", "verify", "POST", map[string]any{"bookingGuid": "b"}, "/api/salesbooking/by-guid/b/verify"},
		{"salesbooking", "cancel", "DELETE", map[string]any{"bookingGuid": "b"}, "/api/salesbooking/by-guid/b"},
		{"salesavailability", "create-override", "POST", nil, "/api/salesavailability/overrides"},
		{"salesavailability", "delete-override", "DELETE", map[string]any{"overrideGuid": "o"}, "/api/salesavailability/overrides/by-guid/o"},

		{"tenant", "invite-defaults-get", "GET", map[string]any{"tenantGuidToGet": "t"}, "/api/tenant/t/invite-defaults"},
		{"tenant", "invite-defaults-set", "PATCH", map[string]any{"tenantGuidToUpdate": "t"}, "/api/tenant/t/invite-defaults"},

		{"wholesaler", "applications", "GET", nil, "/api/wholesaler/application"},
		{"wholesaler", "catalog", "GET", map[string]any{"guid": "g"}, "/api/wholesaler/g/catalog"},
		{"wholesaler", "me", "GET", nil, "/api/wholesaler/me"},

		{"workorder", "quick-close", "POST", nil, "/api/workorder/quick-close"},
		{"workorder-template", "run-schedule-now", "POST", map[string]any{"scheduleGuid": "s"}, "/api/workordertemplate/schedules/s/runnow"},
		{"workorder-template", "analysis-preview", "GET", map[string]any{"templateGuid": "t"}, "/api/workordertemplate/t/analysis/preview"},
		{"workorder-template", "analyze", "POST", map[string]any{"templateGuid": "t"}, "/api/workordertemplate/t/analysis/analyze"},
		{"language", "translate", "POST", nil, "/api/language/translate"},

		{"document", "get-metadata", "GET", map[string]any{"guid": "d"}, "/api/document/d/metadata"},
		{"document", "get-timeline", "GET", nil, "/api/document/timeline"},
		{"knowledgeai", "translate", "POST", map[string]any{"articleGuid": "k"}, "/api/knowledgearticle/by-guid/k/upmate-translate"},
		{"knowledgespace", "usage", "GET", map[string]any{"spaceGuid": "s"}, "/api/knowledgespace/by-guid/s/usage"},
		{"knowledgespace", "list-members", "GET", map[string]any{"spaceGuid": "s"}, "/api/knowledgespace/by-guid/s/members"},
		{"knowledgespace", "add-member", "POST", map[string]any{"spaceGuid": "s"}, "/api/knowledgespace/by-guid/s/members"},
		{"project", "my-projects", "GET", nil, "/api/project/my-projects"},

		// Verified to belong on the domain path itself; they declare UseDomainBasePath.
		{"aicreditgrant", "issue", "POST", nil, "/api/aicreditgrant"},
		{"aicreditrequest", "submit", "POST", nil, "/api/aicreditrequest"},
		{"notification-preference", "set", "PUT", nil, "/api/notificationpreference"},
		{"plan-migration", "migrate", "POST", nil, "/api/planmigration"},
		{"quality-policy", "draft-create", "POST", nil, "/api/quality/policies"},
		{"labour-application-ai", "draft", "POST", nil, "/api/labour-marketplace/ai/contractor-application-draft"},
		{"workorder-template", "active", "GET", nil, "/api/workordertemplate"},
		{"schedule-assignment", "create-by-guid", "POST", nil, "/api/scheduleassignment"},
	}
	for _, test := range tests {
		t.Run(test.domain+"-"+test.action, func(t *testing.T) {
			domain, action := phantomAction(t, test.domain, test.action)
			if action.MCPOnly {
				t.Fatalf("%s %s is MCPOnly — it must not be in the REST route table", test.domain, test.action)
			}
			if action.HTTPMethod != test.method {
				t.Fatalf("method = %q, want %q", action.HTTPMethod, test.method)
			}
			args := test.args
			if args == nil {
				args = map[string]any{}
			}
			path, _ := buildRESTPath(domain, action, args)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
		})
	}
}

func TestActionsWithoutARESTEndpointGoThroughMCP(t *testing.T) {
	t.Parallel()
	for _, item := range []struct{ domain, action string }{
		{"asset-failure", "statistics"},
		{"asset-failure", "classify"},
		{"asset-failure", "resolve"},
		{"knowledgeai", "generate-from-text"},
		{"stock", "po-marketplace-draft"},
		{"stock", "optimization-prepare"},
		{"schedule-assignment", "workorder-options"},
		{"workorder", "closeout-prepare"},
		{"workorder", "closeout-complete"},
	} {
		_, action := phantomAction(t, item.domain, item.action)
		if !action.MCPOnly {
			t.Errorf("%s %s has no REST endpoint and must be MCPOnly", item.domain, item.action)
		}
	}
}

// Guards every registered action, so a new one cannot reintroduce the silent fall-through.
func TestEveryNonStandardRESTActionDeclaresItsRoute(t *testing.T) {
	t.Parallel()
	for _, domain := range DefaultRegistry.Domains() {
		for _, action := range domain.Actions {
			if action.MCPOnly || standardRESTVerbs[action.Name] || strings.HasPrefix(action.Name, "update-") {
				continue
			}
			if action.RESTPath == "" && action.RESTBasePath == "" && !action.UseDomainBasePath {
				t.Errorf("%s %s declares no RESTPath, RESTBasePath or UseDomainBasePath, so it falls through "+
					"to the domain path; set the real route, or MCPOnly if no REST endpoint exists",
					domain.Name, action.Name)
			}
		}
	}
}

// Quick close sends the request to its own endpoint with the Idempotency-Key header the
// backend requires, generating one when the caller leaves the flag out and telling them which.
func TestWorkorderQuickCloseGeneratesAndPrintsIdempotencyKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var received *http.Request
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		received = request.Clone(request.Context())
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "quick-close-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	apiClient := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
	format := "json"
	command := buildDomainCommand(
		findDomain("workorder"),
		func() (*client.APIClient, error) { return apiClient, nil },
		logging.New(logging.LevelError),
		&format,
		&ExportConfig{},
	)
	var stderr bytes.Buffer
	command.SetErr(&stderr)
	command.SetOut(&bytes.Buffer{})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs([]string{"quick-close",
		"--template", "11111111-1111-4111-8111-111111111111",
		"--asset", "22222222-2222-4222-8222-222222222222",
		"--note", "Replaced the filter"})
	if err := command.Execute(); err != nil {
		t.Fatalf("quick-close error = %v", err)
	}
	if received == nil {
		t.Fatal("quick-close sent no request")
	}
	if received.Method != http.MethodPost || received.URL.Path != "/api/workorder/quick-close" {
		t.Fatalf("request = %s %s, want POST /api/workorder/quick-close", received.Method, received.URL.Path)
	}
	key := received.Header.Get("Idempotency-Key")
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(key) {
		t.Fatalf("Idempotency-Key = %q, want a generated v4 GUID", key)
	}
	if !strings.Contains(stderr.String(), key) {
		t.Fatalf("the generated key must be printed so a retry can reuse it; stderr = %q", stderr.String())
	}
}
