package registry

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkorderScopeUsesOwningGuidRoutesAndParsedRootBody(t *testing.T) {
	domain := findDomain("workorder")
	guid := "11111111-1111-4111-8111-111111111111"
	for name, method := range map[string]string{"scope-get": http.MethodGet, "scope-update": http.MethodPut} {
		action := findAction(domain, name)
		if action == nil || action.HTTPMethod != method || action.MCPOnly {
			t.Fatalf("invalid owning scope action %s: %+v", name, action)
		}
		assertRequiredUUIDArg(t, action, "workorderGuid")
		path, consumed := buildRESTPath(domain, *action, map[string]any{"workorderGuid": guid})
		if path != "/api/workorder/by-guid/"+guid+"/project-scope" || len(consumed) != 1 {
			t.Fatalf("unexpected scope path %s: %v", path, consumed)
		}
		if name == "scope-update" {
			if len(action.Flags) != 1 || action.Flags[0].Name != "from-json" || !action.Flags[0].Required || !action.Flags[0].RootJSONObjectFile {
				t.Fatalf("reviewed scope must parse one required root-object file: %+v", action.Flags)
			}
		}
	}
	file := filepath.Join(t.TempDir(), "review.json")
	if err := os.WriteFile(file, []byte(`{"requestGuid":"22222222-2222-4222-8222-222222222222","expectedUpdatedAt":"2026-09-13T00:00:00Z","expectedProjectGuid":"00000000-0000-0000-0000-000000000000","expectedProjectStageGuid":"00000000-0000-0000-0000-000000000000","expectedProjectOutputItemGuid":"00000000-0000-0000-0000-000000000000","projectOutputItemGuid":null}`), 0600); err != nil {
		t.Fatal(err)
	}
	parsed, err := readRootJSONObjectFile(file)
	if err != nil {
		t.Fatal(err)
	}
	body := map[string]any{}
	if err := mergeRootJSONObject(body, parsed); err != nil {
		t.Fatal(err)
	}
	if body["expectedProjectGuid"] != "00000000-0000-0000-0000-000000000000" || body["projectOutputItemGuid"] != nil || len(body) != 6 {
		t.Fatalf("scope presence and explicit unlinked assertion changed: %+v", body)
	}
}
