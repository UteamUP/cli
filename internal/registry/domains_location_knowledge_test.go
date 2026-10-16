package registry

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLocationKnowledgeReviewedRoutes(t *testing.T) {
	for _, action := range locationKnowledgeActions() {
		if !strings.HasSuffix(action.RESTPath, "/reviewed") || len(action.Args) != 2 {
			t.Fatal("unsafe relation route")
		}
		for _, arg := range action.Args {
			if arg.Type != "non-empty-uuid" || !arg.Required {
				t.Fatal("non-GUID boundary")
			}
		}
		confirm := action.Flags[1]
		if !confirm.LocalOnly || !confirm.MustBeTrue || !confirm.Required {
			t.Fatal("confirmation missing")
		}
		body := map[string]any{"idempotencyKey": "11111111-1111-1111-1111-111111111111", "mutationOutcomeVersion": json.Number("1")}
		if validateLocationKnowledgeOriginal(action, body) == nil {
			t.Fatal("missing baseline accepted")
		}
		if action.HTTPMethod == "POST" {
			body["expectedLinkGuid"] = nil
		} else {
			body["expectedLinkGuid"] = "22222222-2222-2222-2222-222222222222"
		}
		if err := validateLocationKnowledgeOriginal(action, body); err != nil {
			t.Fatal(err)
		}
		body["expectedLinkGuid"] = "00000000-0000-0000-0000-000000000000"
		if validateLocationKnowledgeOriginal(action, body) == nil {
			t.Fatal("invalid baseline accepted")
		}
	}
}
