package registry

import "testing"

func TestLocationNodeTypeWritesUseReviewedGuidRoutes(t *testing.T) {
	const typeGUID = "11111111-1111-4111-8111-111111111111"
	domain := findDomainByName(t, "location")
	for _, action := range locationNodeTypeActions() {
		if err := validateActionDefinition(action); err != nil {
			t.Fatal(err)
		}
		path, consumed := buildRESTPath(domain, action, map[string]any{"typeGuid": typeGUID})
		want := "/api/locationnodetype"
		if action.Name == "node-type-update" {
			want += "/" + typeGUID
			if len(consumed) != 1 || consumed[0] != "typeGuid" || action.Args[0].Type != "non-empty-uuid" {
				t.Fatal("update must consume only the public type GUID")
			}
		}
		if path != want || action.MCPOnly {
			t.Fatalf("%s routes to %s, want %s", action.Name, path, want)
		}
		if action.HTTPMethod == "GET" {
			continue
		}
		confirmed := false
		for _, flag := range action.Flags {
			if flag.Name == "confirm" {
				confirmed = flag.LocalOnly && flag.Required && flag.MustBeTrue
			}
		}
		if !confirmed {
			t.Fatalf("%s must require reviewed confirmation", action.Name)
		}
	}
}
