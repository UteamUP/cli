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

func TestLocationReadActionsUseGuidRestRoutes(t *testing.T) {
	const locationGUID = "11111111-1111-4111-8111-111111111111"
	cases := []struct {
		name      string
		args      []string
		wantPath  string
		wantQuery map[string]string
		body      string
	}{
		{name: "search is sent as nameFilter", args: []string{"list", "--search", "Vesturberg 111 - Bílskúr"}, wantPath: "/api/location",
			wantQuery: map[string]string{"nameFilter": "Vesturberg 111 - Bílskúr"}, body: `{"locations":[],"totalCount":0}`},
		{name: "list without search sends no name filter", args: []string{"list"}, wantPath: "/api/location",
			wantQuery: map[string]string{"nameFilter": ""}, body: `{"locations":[],"totalCount":0}`},
		{name: "get uses the public GUID", args: []string{"get", locationGUID}, wantPath: "/api/location/" + locationGUID, body: `{"guid":"` + locationGUID + `"}`},
		{name: "node tree reads the GUID-keyed nodes route", args: []string{"node-tree", locationGUID}, wantPath: "/api/location/" + locationGUID + "/nodes", body: `[]`},
		{name: "node type catalog uses its own REST collection", args: []string{"node-types"}, wantPath: "/api/locationnodetype", body: `[]`},
		{name: "node tree forwards public type GUID", args: []string{"node-tree", locationGUID, "--node-type-guid", locationGUID}, wantPath: "/api/location/" + locationGUID + "/nodes", wantQuery: map[string]string{"locationNodeTypeGuid": locationGUID}, body: `[]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			count := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				if r.Method != http.MethodGet || r.URL.Path != tc.wantPath {
					t.Errorf("wrong route: %s %s", r.Method, r.URL.Path)
				}
				for key, want := range tc.wantQuery {
					if got := r.URL.Query().Get(key); got != want {
						t.Errorf("query %s = %q, want %q", key, got, want)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "test", ExpiresAt: time.Now().Add(time.Hour), TenantGUID: locationGUID}); err != nil {
				t.Fatal(err)
			}
			api := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			cmd := buildDomainCommand(findDomainByName(t, "location"), func() (*client.APIClient, error) { return api, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
			cmd.SetArgs(tc.args)
			if err := cmd.Execute(); err != nil || count != 1 {
				t.Fatalf("request failed: %v count=%d", err, count)
			}
		})
	}
}

func TestLocationNodeTreeRejectsANonGuidLocation(t *testing.T) {
	for _, action := range locationActions() {
		if action.Name != "node-tree" {
			continue
		}
		if action.Args[0].Type != "non-empty-uuid" || action.HTTPMethod != "GET" || action.MCPOnly {
			t.Fatalf("node-tree must be a GUID-keyed REST read: %#v", action)
		}
		return
	}
	t.Fatal("node-tree action is not registered")
}
