package registry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

// UTE-960: the backend lets a space manager pick the platform permissions and tenant
// roles that grant access to a knowledge space. The CLI mirrors that contract on the
// existing knowledgespace domain: `access-catalog` reads the choices, and create/update
// carry them as requiredPermissions (string[]) and allowedRoleGuids (Guid[]).

const knowledgeSpaceGUID = "11111111-1111-4111-8111-111111111111"
const knowledgeSpaceRoleGUID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

// knowledgeSpaceAccessCatalogPayload is the backend contract shape:
// {permissions:[{id,name,category}],roles:[{guid,name}],supportsRoleGroups:false}.
const knowledgeSpaceAccessCatalogPayload = `{"permissions":[` +
	`{"id":"Asset.View","name":"View assets","category":"Asset"},` +
	`{"id":"Knowledge.ManageSpaces","name":"Manage knowledge spaces","category":"Knowledge"}` +
	`],"roles":[{"guid":"` + knowledgeSpaceRoleGUID + `","name":"Site supervisor"}],` +
	`"supportsRoleGroups":false}`

// recordedSpaceRequest is what the CLI actually sent for one knowledge-space action.
type recordedSpaceRequest struct {
	method string
	path   string
	body   map[string]any
	stdout string
}

// runKnowledgeSpaceAction executes one knowledgespace action against a TLS test API and
// returns the request it received, so the tests prove the real route and the real body.
func runKnowledgeSpaceAction(t *testing.T, cliArgs ...string) recordedSpaceRequest {
	t.Helper()
	return runKnowledgeSpaceActionWithPayload(t, `{}`, cliArgs...)
}

func runKnowledgeSpaceActionWithPayload(t *testing.T, payload string, cliArgs ...string) recordedSpaceRequest {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	recorded := recordedSpaceRequest{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		bodyBytes, _ := io.ReadAll(request.Body)
		recorded = recordedSpaceRequest{method: request.Method, path: request.URL.Path}
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &recorded.body); err != nil {
				t.Errorf("decode body %q: %v", bodyBytes, err)
			}
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(payload))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{
		APIOrigin:   server.URL,
		AccessToken: "knowledge-space-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		TenantGUID:  "55555555-5555-4555-8555-555555555555",
	}); err != nil {
		t.Fatalf("save test token: %v", err)
	}

	domain := findDomain("knowledgespace")
	if domain == nil {
		t.Fatal("knowledgespace domain is not registered")
	}
	apiClient := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
	format := "json"
	command := buildDomainCommand(
		domain,
		func() (*client.APIClient, error) { return apiClient, nil },
		logging.New(logging.LevelError),
		&format,
		&ExportConfig{},
	)
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(cliArgs)
	var runErr error
	stdout := captureRegistryStdout(t, func() { runErr = command.Execute() })
	if runErr != nil {
		t.Fatalf("knowledgespace %v error = %v (stdout %q)", cliArgs, runErr, stdout)
	}
	recorded.stdout = stdout
	return recorded
}

func knowledgeSpaceAction(t *testing.T, name string) *Action {
	t.Helper()
	domain := findDomain("knowledgespace")
	if domain == nil {
		t.Fatal("knowledgespace domain is not registered")
	}
	action := findAction(domain, name)
	if action == nil {
		t.Fatalf("knowledgespace %s action is missing", name)
	}
	return action
}

// The catalog is the manager-only read that feeds the picker, so it must name the
// scope permission it needs and hit the backend's access-catalog route.
func TestKnowledgeSpaceAccessCatalogRouteAndScopePermission(t *testing.T) {
	action := knowledgeSpaceAction(t, "access-catalog")
	if action.ToolName != "UteamupKnowledgeSpaceAccessCatalog" {
		t.Errorf("access-catalog tool = %q, want UteamupKnowledgeSpaceAccessCatalog", action.ToolName)
	}
	if action.MCPOnly {
		t.Error("access-catalog is served by GET /api/KnowledgeSpace/access-catalog and must not be MCPOnly")
	}
	if action.HTTPMethod != http.MethodGet {
		t.Errorf("access-catalog method = %q, want GET", action.HTTPMethod)
	}
	if !strings.Contains(action.Description, "Knowledge.ManageSpaces") {
		t.Errorf("access-catalog must name its scope permission, got %q", action.Description)
	}

	domain := findDomain("knowledgespace")
	path, consumed := buildRESTPath(domain, *action, map[string]any{})
	if path != "/api/knowledgespace/access-catalog" || len(consumed) != 0 {
		t.Errorf("access-catalog path = %q (consumed %v), want /api/knowledgespace/access-catalog", path, consumed)
	}
}

// The catalog payload carries the scope permission metadata the picker needs; the CLI
// must hand it back untouched, including the explicit "no role groups" answer.
func TestKnowledgeSpaceAccessCatalogKeepsScopePermissionMetadata(t *testing.T) {
	recorded := runKnowledgeSpaceActionWithPayload(t, knowledgeSpaceAccessCatalogPayload, "access-catalog")
	if recorded.method != http.MethodGet || recorded.path != "/api/knowledgespace/access-catalog" {
		t.Fatalf("access-catalog request = %s %s, want GET /api/knowledgespace/access-catalog", recorded.method, recorded.path)
	}
	if len(recorded.body) != 0 {
		t.Errorf("access-catalog must not send a body, got %v", recorded.body)
	}

	var catalog struct {
		Permissions []struct {
			ID       string  `json:"id"`
			Name     string  `json:"name"`
			Category *string `json:"category"`
		} `json:"permissions"`
		Roles []struct {
			GUID string `json:"guid"`
			Name string `json:"name"`
		} `json:"roles"`
		SupportsRoleGroups bool `json:"supportsRoleGroups"`
	}
	if err := json.Unmarshal([]byte(recorded.stdout), &catalog); err != nil {
		t.Fatalf("read catalog output %q: %v", recorded.stdout, err)
	}
	if len(catalog.Permissions) != 2 {
		t.Fatalf("catalog permissions = %#v, want the two permission options", catalog.Permissions)
	}
	if catalog.Permissions[0].ID != "Asset.View" || catalog.Permissions[0].Name != "View assets" {
		t.Errorf("catalog permission = %#v, want Asset.View / View assets", catalog.Permissions[0])
	}
	if catalog.Permissions[0].Category == nil || *catalog.Permissions[0].Category != "Asset" {
		t.Errorf("catalog permission category = %#v, want Asset", catalog.Permissions[0].Category)
	}
	if len(catalog.Roles) != 1 || catalog.Roles[0].GUID != knowledgeSpaceRoleGUID || catalog.Roles[0].Name != "Site supervisor" {
		t.Errorf("catalog roles = %#v, want the one tenant role", catalog.Roles)
	}
	if catalog.SupportsRoleGroups {
		t.Error("role groups have no persisted template lineage; the catalog must report them unsupported")
	}
}

// Create and update must expose the same access choices the backend accepts: the
// obsolete single requiredPermission stays, and the two arrays are added next to it.
func TestKnowledgeSpaceCreateAndUpdateCarryAccessChoices(t *testing.T) {
	for _, name := range []string{"create", "update"} {
		action := knowledgeSpaceAction(t, name)
		flags := flagsToMap(action.Flags)

		legacy, ok := flags["required-permission"]
		if !ok {
			t.Fatalf("%s must keep the obsolete --required-permission flag", name)
		}
		if legacy.BodyName != "requiredPermission" || legacy.Type != "string" {
			t.Errorf("%s --required-permission = %#v, want string mapped to requiredPermission", name, legacy)
		}

		permissions, ok := flags["required-permissions"]
		if !ok {
			t.Fatalf("%s must expose --required-permissions", name)
		}
		if permissions.BodyName != "requiredPermissions" || permissions.Type != "stringSlice" {
			t.Errorf("%s --required-permissions = %#v, want stringSlice mapped to requiredPermissions", name, permissions)
		}
		if permissions.Required {
			t.Errorf("%s --required-permissions must stay optional so an update can preserve the stored list", name)
		}
		if permissions.Default != nil {
			t.Errorf("%s --required-permissions must have no default so an omitted list is never sent", name)
		}

		roles, ok := flags["allowed-role-guids"]
		if !ok {
			t.Fatalf("%s must expose --allowed-role-guids", name)
		}
		if roles.BodyName != "allowedRoleGuids" || roles.Type != "stringSlice" {
			t.Errorf("%s --allowed-role-guids = %#v, want stringSlice mapped to allowedRoleGuids", name, roles)
		}
		if roles.Required {
			t.Errorf("%s --allowed-role-guids must stay optional so an update can preserve the stored list", name)
		}
		if roles.Default != nil {
			t.Errorf("%s --allowed-role-guids must have no default so an omitted list is never sent", name)
		}
	}
}

// The access choices only reach the backend if the update lands on its by-guid route.
func TestKnowledgeSpaceItemRoutesUseTheByGuidRoute(t *testing.T) {
	domain := findDomain("knowledgespace")
	for _, name := range []string{"get", "update", "delete"} {
		action := knowledgeSpaceAction(t, name)
		path, consumed := buildRESTPath(domain, *action, map[string]any{"spaceGuid": knowledgeSpaceGUID})
		if path != "/api/knowledgespace/by-guid/"+knowledgeSpaceGUID {
			t.Errorf("%s path = %q, want /api/knowledgespace/by-guid/%s", name, path, knowledgeSpaceGUID)
		}
		if len(consumed) != 1 {
			t.Errorf("%s consumed %v, want the spaceGuid to leave the body", name, consumed)
		}
	}
}

// The two lists are JSON arrays on the wire: permission IDs stay strings and role GUIDs
// stay GUID strings, never one joined string.
func TestKnowledgeSpaceCreateSerializesAccessChoicesAsJsonArrays(t *testing.T) {
	recorded := runKnowledgeSpaceAction(t, "create", "--name", "Ops space",
		"--required-permissions", "Asset.View,WorkOrder.View",
		"--allowed-role-guids", knowledgeSpaceRoleGUID)
	if recorded.method != http.MethodPost || recorded.path != "/api/knowledgespace" {
		t.Fatalf("create request = %s %s, want POST /api/knowledgespace", recorded.method, recorded.path)
	}

	permissions, ok := recorded.body["requiredPermissions"].([]any)
	if !ok {
		t.Fatalf("requiredPermissions = %#v, want a JSON array", recorded.body["requiredPermissions"])
	}
	if len(permissions) != 2 || permissions[0] != "Asset.View" || permissions[1] != "WorkOrder.View" {
		t.Errorf("requiredPermissions = %#v, want [Asset.View WorkOrder.View]", permissions)
	}

	roles, ok := recorded.body["allowedRoleGuids"].([]any)
	if !ok {
		t.Fatalf("allowedRoleGuids = %#v, want a JSON array", recorded.body["allowedRoleGuids"])
	}
	if len(roles) != 1 || roles[0] != knowledgeSpaceRoleGUID {
		t.Errorf("allowedRoleGuids = %#v, want [%s]", roles, knowledgeSpaceRoleGUID)
	}
	if _, present := recorded.body["requiredPermission"]; present {
		t.Error("the obsolete requiredPermission must stay absent unless the caller asks for it")
	}
}

// Update semantics: an omitted list is never sent so the backend preserves it, an empty
// list is sent as [] so the backend clears it, and the obsolete field can still be set
// explicitly alongside the arrays.
func TestKnowledgeSpaceUpdateDistinguishesOmittedAndClearedAccessLists(t *testing.T) {
	t.Run("omitted lists are not sent", func(t *testing.T) {
		recorded := runKnowledgeSpaceAction(t, "update", knowledgeSpaceGUID, "--name", "Renamed space")
		if recorded.method != http.MethodPut || recorded.path != "/api/knowledgespace/by-guid/"+knowledgeSpaceGUID {
			t.Fatalf("update request = %s %s, want PUT /api/knowledgespace/by-guid/%s", recorded.method, recorded.path, knowledgeSpaceGUID)
		}
		for _, key := range []string{"requiredPermissions", "allowedRoleGuids"} {
			if _, present := recorded.body[key]; present {
				t.Errorf("%s must be absent so the stored list is preserved, got %#v", key, recorded.body[key])
			}
		}
		if _, present := recorded.body["spaceGuid"]; present {
			t.Errorf("the space GUID belongs in the route, not the body, got %#v", recorded.body["spaceGuid"])
		}
		if recorded.body["name"] != "Renamed space" {
			t.Errorf("name = %#v, want Renamed space", recorded.body["name"])
		}
	})

	t.Run("empty lists clear the stored lists", func(t *testing.T) {
		recorded := runKnowledgeSpaceAction(t, "update", knowledgeSpaceGUID,
			"--required-permissions=", "--allowed-role-guids=")
		for _, key := range []string{"requiredPermissions", "allowedRoleGuids"} {
			list, ok := recorded.body[key].([]any)
			if !ok {
				t.Fatalf("%s = %#v, want an empty JSON array", key, recorded.body[key])
			}
			if len(list) != 0 {
				t.Errorf("%s = %#v, want an empty JSON array", key, list)
			}
		}
	})

	t.Run("the obsolete field can be supplied alongside", func(t *testing.T) {
		recorded := runKnowledgeSpaceAction(t, "update", knowledgeSpaceGUID,
			"--required-permissions", "Asset.View",
			"--required-permission", "System.All")
		if recorded.body["requiredPermission"] != "System.All" {
			t.Errorf("requiredPermission = %#v, want System.All", recorded.body["requiredPermission"])
		}
		permissions, ok := recorded.body["requiredPermissions"].([]any)
		if !ok || len(permissions) != 1 || permissions[0] != "Asset.View" {
			t.Errorf("requiredPermissions = %#v, want [Asset.View]", recorded.body["requiredPermissions"])
		}
	})
}
