package registry

import (
	"strings"
	"testing"
)

// Every {placeholder} in a RESTPath must name a key executeAction actually stores. Positional
// args are stored under BodyName, else their raw Name (no camelCase), while flags are stored
// under BodyName, else camelCase(Name). A dashed arg such as asset-guid therefore never fills
// {assetGuid}: the request went to a literal /api/assets/{assetGuid}/... and answered 404.
func TestEveryRESTPathPlaceholderResolvesToAStoredArgKey(t *testing.T) {
	for _, domain := range DefaultRegistry.Domains() {
		for _, action := range domain.Actions {
			keys := map[string]bool{}
			for _, arg := range action.Args {
				if arg.QueryName != "" {
					continue
				}
				if arg.BodyName != "" {
					keys[arg.BodyName] = true
				} else {
					keys[arg.Name] = true
				}
			}
			for _, flag := range action.Flags {
				if flag.LocalOnly || (flag.HeaderName != "" && !flag.MirrorHeaderInBody) {
					continue
				}
				if flag.BodyName != "" {
					keys[flag.BodyName] = true
				} else {
					keys[toCamelCase(flag.Name)] = true
				}
			}

			for _, name := range placeholderNames(action.RESTPath) {
				if !keys[name] {
					t.Errorf("%s %s: RESTPath %q has {%s} but no positional arg or flag stores that key (set BodyName: %q on the arg)",
						domain.Name, action.Name, action.RESTPath, name, name)
				}
			}
		}
	}
}

func placeholderNames(path string) []string {
	var names []string
	for {
		start := strings.Index(path, "{")
		if start < 0 {
			return names
		}
		end := strings.Index(path[start:], "}")
		if end < 0 {
			return names
		}
		names = append(names, path[start+1:start+end])
		path = path[start+end+1:]
	}
}

// Drives the real Cobra command, so the positional args travel the same executeAction path a user's do.
func TestMeterReadingHistoryPositionalGuidsFillTheRequestPath(t *testing.T) {
	const assetGuid = "11111111-1111-4111-8111-111111111111"
	const attributeGuid = "22222222-2222-4222-8222-222222222222"
	action := findDomainAction(t, "meter-reading", "history")

	requests, _, _ := runResponseBodyAction(t, *action, t.TempDir(), assetGuid, attributeGuid)

	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	want := "/api/report/" + assetGuid + "/meter-readings/" + attributeGuid + "/history"
	if requests[0].method != "GET" || requests[0].path != want {
		t.Fatalf("request = %s %s, want GET %s", requests[0].method, requests[0].path, want)
	}
	for key := range requests[0].query {
		if strings.Contains(strings.ToLower(key), "guid") {
			t.Errorf("path GUID leaked into the query string as %q", key)
		}
	}
}
