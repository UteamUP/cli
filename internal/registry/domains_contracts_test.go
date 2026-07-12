package registry

import "testing"

func findLabourRateDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "labour-rate" {
			return domain
		}
	}
	t.Fatal("labour-rate domain is not registered")
	return nil
}

func TestLabourRateDomainUsesGuidFirstRoutes(t *testing.T) {
	domain := findLabourRateDomain(t)
	if domain.APIPath != "/api/labourrate" {
		t.Fatalf("APIPath = %q, want /api/labourrate", domain.APIPath)
	}

	expected := map[string]struct {
		method   string
		path     string
		guidName string
	}{
		"list-rules":      {method: "GET", path: "rules"},
		"create-rule":     {method: "POST", path: "rules"},
		"update-rule":     {method: "PUT", path: "rules/{ruleGuid}", guidName: "ruleGuid"},
		"delete-rule":     {method: "DELETE", path: "rules/{ruleGuid}", guidName: "ruleGuid"},
		"list-modifiers":  {method: "GET", path: "modifiers"},
		"create-modifier": {method: "POST", path: "modifiers"},
		"update-modifier": {
			method: "PUT", path: "modifiers/{modifierGuid}", guidName: "modifierGuid",
		},
		"delete-modifier": {
			method: "DELETE", path: "modifiers/{modifierGuid}", guidName: "modifierGuid",
		},
	}

	if len(domain.Actions) != len(expected) {
		t.Fatalf("action count = %d, want %d", len(domain.Actions), len(expected))
	}

	for _, action := range domain.Actions {
		want, ok := expected[action.Name]
		if !ok {
			t.Fatalf("unexpected labour-rate action %q", action.Name)
		}
		if action.HTTPMethod != want.method || action.RESTPath != want.path {
			t.Errorf(
				"%s route = %s %s, want %s %s",
				action.Name,
				action.HTTPMethod,
				action.RESTPath,
				want.method,
				want.path,
			)
		}
		if want.guidName == "" {
			continue
		}
		if len(action.Args) != 1 {
			t.Fatalf("%s args = %+v, want one GUID arg", action.Name, action.Args)
		}
		arg := action.Args[0]
		if arg.Name != want.guidName || arg.Type != "string" || !arg.Required {
			t.Errorf("%s GUID arg = %+v", action.Name, arg)
		}
	}
}

func TestLabourRateDomainDoesNotExposeIntegerIdArguments(t *testing.T) {
	domain := findLabourRateDomain(t)

	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Name == "id" || arg.Type == "int" {
				t.Errorf("%s exposes legacy integer argument %+v", action.Name, arg)
			}
		}
	}
}
