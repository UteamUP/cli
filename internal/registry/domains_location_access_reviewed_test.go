package registry

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLocationAccessReviewedOriginals(t *testing.T) {
	for _, action := range locationAccessReviewedActions() {
		if action.RESTBasePath != "/api/location" || !strings.HasSuffix(action.RESTPath, "/reviewed") || action.MCPOnly {
			t.Fatal(action)
		}
		if !action.Flags[0].Required || !action.Flags[1].MustBeTrue || !action.Flags[1].LocalOnly {
			t.Fatal("reviewed originals and local confirmation required")
		}
		body := map[string]any{"idempotencyKey": "11111111-2222-4333-8444-555555555555", "mutationOutcomeVersion": json.Number("1")}
		if action.HTTPMethod != "POST" {
			body["expectedUpdatedAtUtc"] = "2026-10-16T12:00:00.1234567Z"
		}
		if err := validateLocationAccessReviewedOriginal(action, body); err != nil {
			t.Fatal(err)
		}
		body["expectedUpdatedAtUtc"] = "2026-10-16T12:00:00+01:00"
		if validateLocationAccessReviewedOriginal(action, body) == nil {
			t.Fatal("noncanonical revision accepted")
		}
		delete(body, "idempotencyKey")
		if validateLocationAccessReviewedOriginal(action, body) == nil {
			t.Fatal("missing original accepted")
		}
	}
}
