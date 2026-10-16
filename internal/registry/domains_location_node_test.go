package registry

import (
	"encoding/json"
	"testing"
)

func TestLocationNodeRoutesAndOriginalMetadata(t *testing.T) {
	actions := locationNodeActions()
	expected := map[string]string{"node-create": "{locationGuid}/nodes", "node-update": "node/{nodeGuid}", "node-move": "node/{nodeGuid}/move", "node-delete": "node/{nodeGuid}"}
	for _, action := range actions {
		if action.RESTPath != expected[action.Name] || action.MCPOnly {
			t.Fatalf("incorrect node route: %#v", action)
		}
		if action.Args[0].Type != "non-empty-uuid" {
			t.Fatal("node target must be a public GUID")
		}
		confirmed := false
		for _, flag := range action.Flags {
			if flag.Name == "confirm" {
				confirmed = flag.LocalOnly && flag.Required && flag.MustBeTrue
			}
		}
		if !confirmed {
			t.Fatal("node mutations require local confirmation")
		}
		if action.Name == "node-delete" {
			continue
		}
		if validateLocationNodeOriginal(action, map[string]any{}) == nil {
			t.Fatal("missing original metadata accepted")
		}
		body := map[string]any{"idempotencyKey": "11111111-1111-4111-8111-111111111111", "mutationOutcomeVersion": json.Number("1")}
		if action.Name != "node-create" {
			body["expectedUpdatedAt"] = "2026-10-16T12:00:00.1234567Z"
		}
		if err := validateLocationNodeOriginal(action, body); err != nil {
			t.Fatal(err)
		}
		body["expectedUpdatedAt"] = "2026-10-16T12:00:00.12345678Z"
		if validateLocationNodeOriginal(action, body) == nil {
			t.Fatal("lossy reviewed revision accepted")
		}
	}
}
