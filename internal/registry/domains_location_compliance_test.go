package registry

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLocationComplianceReviewedOriginals(t *testing.T) {
	for _, action := range locationComplianceActions() {
		if action.RESTBasePath != "/api/compliance" || !strings.HasSuffix(action.RESTPath, "/reviewed") || action.MCPOnly {
			t.Fatal(action)
		}
		if !action.Flags[0].Required || !action.Flags[1].LocalOnly || !action.Flags[1].MustBeTrue {
			t.Fatal("reviewed input and local confirmation required")
		}
		body := map[string]any{"idempotencyKey": "11111111-2222-4333-8444-555555555555", "mutationOutcomeVersion": json.Number("1")}
		if action.HTTPMethod != "POST" {
			body["expectedUpdatedAt"] = "2026-10-16T12:00:00.1234567Z"
		}
		if err := validateLocationComplianceOriginal(action, body); err != nil {
			t.Fatal(err)
		}
		body["documentGuids"] = []any{}
		if validateLocationComplianceOriginal(action, body) == nil {
			t.Fatal("document edits accepted")
		}
		delete(body, "documentGuids")
		delete(body, "idempotencyKey")
		if validateLocationComplianceOriginal(action, body) == nil {
			t.Fatal("missing operation accepted")
		}
	}
}
