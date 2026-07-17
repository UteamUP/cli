package registry

import (
	"strings"
	"testing"
)

func TestRootCauseDomainUsesGuidFirstRoutes(t *testing.T) {
	domain := findRootCauseDomain(t)
	if domain.APIPath != "/api/rootcauseanalysis" {
		t.Fatalf("APIPath = %q, want /api/rootcauseanalysis", domain.APIPath)
	}

	expectedPaths := map[string]string{
		"get":                     "{rcaGuid}",
		"update":                  "{rcaGuid}",
		"complete":                "{rcaGuid}/complete",
		"delete":                  "{rcaGuid}",
		"step-add":                "{rcaGuid}/steps",
		"step-update":             "{rcaGuid}/steps/{stepGuid}",
		"step-delete":             "{rcaGuid}/steps/{stepGuid}",
		"steps-reorder":           "{rcaGuid}/steps/reorder",
		"action-add":              "{rcaGuid}/actions",
		"action-update":           "{rcaGuid}/actions/by-guid/{actionGuid}",
		"action-delete":           "{rcaGuid}/actions/by-guid/{actionGuid}",
		"action-create-workorder": "{rcaGuid}/actions/by-guid/{actionGuid}/create-workorder",
		"assets-link":             "{rcaGuid}/assets",
		"asset-unlink":            "{rcaGuid}/assets/{assetGuid}",
		"parts-link":              "{rcaGuid}/parts",
		"part-unlink":             "{rcaGuid}/parts/{partGuid}",
		"workorders-link":         "{rcaGuid}/workorders",
		"workorder-unlink":        "{rcaGuid}/workorders/{workOrderGuid}",
		"knowledge-link":          "{rcaGuid}/knowledgearticles",
		"knowledge-unlink":        "{rcaGuid}/knowledgearticles/{articleGuid}",
	}

	for name, expectedPath := range expectedPaths {
		action := findRootCauseAction(t, domain, name)
		if action.RESTPath != expectedPath {
			t.Errorf("%s RESTPath = %q, want %q", name, action.RESTPath, expectedPath)
		}
		if strings.Contains(strings.ToLower(action.RESTPath), "{id}") {
			t.Errorf("%s exposes an integer identity in %q", name, action.RESTPath)
		}
		for _, argument := range action.Args {
			if strings.HasSuffix(argument.Name, "Guid") && argument.Type != "uuid" {
				t.Errorf("%s argument %s type = %q, want uuid", name, argument.Name, argument.Type)
			}
			if argument.Type == "int" {
				t.Errorf("%s exposes integer argument %+v", name, argument)
			}
		}
	}
}

func TestRootCauseNestedRouteExpandsGuids(t *testing.T) {
	domain := findRootCauseDomain(t)
	action := findRootCauseAction(t, domain, "action-create-workorder")
	args := map[string]any{
		"rcaGuid":    "11111111-1111-1111-1111-111111111111",
		"actionGuid": "22222222-2222-2222-2222-222222222222",
	}

	path, consumed := buildRESTPath(domain, *action, args)

	expected := "/api/rootcauseanalysis/11111111-1111-1111-1111-111111111111/actions/by-guid/22222222-2222-2222-2222-222222222222/create-workorder"
	if path != expected {
		t.Fatalf("path = %q, want %q", path, expected)
	}
	if len(consumed) != len(args) {
		t.Fatalf("consumed %d arguments, want %d", len(consumed), len(args))
	}
}

func findRootCauseDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "root-cause" {
			return domain
		}
	}
	t.Fatal("root-cause domain is not registered")
	return nil
}

func findRootCauseAction(t *testing.T, domain *Domain, name string) *Action {
	t.Helper()
	for index := range domain.Actions {
		if domain.Actions[index].Name == name {
			return &domain.Actions[index]
		}
	}
	t.Fatalf("%s action is not registered", name)
	return nil
}
