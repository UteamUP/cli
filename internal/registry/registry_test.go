package registry

import (
	"testing"
)

// TestBuildRESTPathUpdateSubRoutes locks in the update-<sub> sub-route convention:
// `update-status` → /{id}/status (explicit case), `update-notes` → /{id}/notes
// (generic fallback). Without these, PATCH endpoints route to the basePath and
// produce 405/404 from the backend.
func TestBuildRESTPathUpdateSubRoutes(t *testing.T) {
	domain := &Domain{Name: "bugsandfeatures", APIPath: "/api/bugsandfeatures"}
	cases := []struct {
		actionName string
		argKey     string
		argValue   any
		want       string
	}{
		{"update-status", "externalGuid", "abc-123", "/api/bugsandfeatures/abc-123/status"},
		{"update-notes", "externalGuid", "abc-123", "/api/bugsandfeatures/abc-123/notes"},
		{"update-status", "id", 42, "/api/bugsandfeatures/42/status"},
		{"update-priority", "externalGuid", "g1", "/api/bugsandfeatures/g1/priority"},
		{"get", "externalGuid", "g1", "/api/bugsandfeatures/g1"},
	}
	for _, tc := range cases {
		t.Run(tc.actionName, func(t *testing.T) {
			got := buildRESTPath(domain, Action{Name: tc.actionName}, map[string]any{tc.argKey: tc.argValue})
			if got != tc.want {
				t.Errorf("buildRESTPath(%s) = %q, want %q", tc.actionName, got, tc.want)
			}
		})
	}
}

func TestHTTPMethodForUpdateNotes(t *testing.T) {
	if HTTPMethod["update-notes"] != "PATCH" {
		t.Errorf("update-notes HTTPMethod = %q, want PATCH", HTTPMethod["update-notes"])
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"page", "page"},
		{"page-size", "pageSize"},
		{"sort-by", "sortBy"},
		{"sort-order", "sortOrder"},
		{"from-json", "fromJson"},
		{"asset-type-id", "assetTypeId"},
		{"a", "a"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := toCamelCase(tc.input)
			if result != tc.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestDefaultRegistryHasDomains(t *testing.T) {
	domains := DefaultRegistry.Domains()
	if len(domains) == 0 {
		t.Fatal("expected at least one registered domain")
	}

	// Check that our 3 starter domains are registered
	domainNames := make(map[string]bool)
	for _, d := range domains {
		domainNames[d.Name] = true
	}

	expected := []string{"asset", "workorder", "user"}
	for _, name := range expected {
		if !domainNames[name] {
			t.Errorf("expected domain %q to be registered", name)
		}
	}
}

func TestAssetDomainActions(t *testing.T) {
	var assetDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "asset" {
			assetDomain = d
			break
		}
	}
	if assetDomain == nil {
		t.Fatal("asset domain not found")
	}

	expectedActions := []string{"list", "get", "get-by-guid", "create", "update", "delete", "search"}
	actionNames := make(map[string]bool)
	for _, a := range assetDomain.Actions {
		actionNames[a.Name] = true
	}

	for _, name := range expectedActions {
		if !actionNames[name] {
			t.Errorf("expected action %q in asset domain", name)
		}
	}
}

func TestWorkorderDomainAliases(t *testing.T) {
	var woDomain *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "workorder" {
			woDomain = d
			break
		}
	}
	if woDomain == nil {
		t.Fatal("workorder domain not found")
	}

	hasWO := false
	for _, alias := range woDomain.Aliases {
		if alias == "wo" {
			hasWO = true
		}
	}
	if !hasWO {
		t.Error("workorder domain should have 'wo' alias")
	}
}

func TestDomainToolNames(t *testing.T) {
	for _, d := range DefaultRegistry.Domains() {
		for _, a := range d.Actions {
			if a.ToolName == "" {
				t.Errorf("domain %s action %s has empty ToolName", d.Name, a.Name)
			}
		}
	}
}
