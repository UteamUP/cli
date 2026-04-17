package registry

import (
	"testing"
)

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
