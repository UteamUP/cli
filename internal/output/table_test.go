package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs fn, redirecting stdout to a buffer, and returns what was
// printed. Restores the original stdout even on panic.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()

	_ = w.Close()
	<-done
	_ = r.Close()
	return buf.String()
}

func TestPrintObjectTable_RendersStatusHistoryBlock_Chronological(t *testing.T) {
	payload := `{
      "externalGuid": "00000000-0000-0000-0000-000000000001",
      "title": "Example",
      "status": "New",
      "statusHistory": [
        {"changedAtUtc":"2026-04-24T10:15:02Z","fromStatus":"Confirmed","toStatus":"New","changedByUserId":"system:auto-ingest","changedByUserEmail":null,"note":"[auto-reopen]; prev=Confirmed; source=BackendAuto; exception=System.NullReferenceException; route=GET /api/x; user=thelma@uteamup.com; occurrence=4; at=2026-04-24T10:15:02Z"},
        {"changedAtUtc":"2026-04-20T09:00:00Z","fromStatus":"Validated","toStatus":"Fixed","changedByUserId":"user-123","changedByUserEmail":"eng@uteamup.com","note":"[debug-skill:post-review] summary: verified"}
      ]
    }`

	out := captureStdout(t, func() {
		if err := Print(FormatTable, json.RawMessage(payload)); err != nil {
			t.Fatalf("Print: %v", err)
		}
	})

	// Block header present
	if !strings.Contains(out, "History:") {
		t.Fatalf("expected 'History:' header in output, got:\n%s", out)
	}
	// Chronological order: 2026-04-20 (oldest) must appear before 2026-04-24
	idxOld := strings.Index(out, "2026-04-20T09:00:00Z")
	idxNew := strings.Index(out, "2026-04-24T10:15:02Z")
	if idxOld < 0 || idxNew < 0 || idxOld > idxNew {
		t.Fatalf("entries out of chronological order. old=%d new=%d output:\n%s", idxOld, idxNew, out)
	}
	// Author fallback: system:auto-ingest (email is null) must appear
	if !strings.Contains(out, "system:auto-ingest") {
		t.Fatalf("expected system:auto-ingest author in output, got:\n%s", out)
	}
	// Email-populated author visible
	if !strings.Contains(out, "eng@uteamup.com") {
		t.Fatalf("expected email author in output, got:\n%s", out)
	}
	// Original statusHistory key is suppressed from the key/value table (would
	// otherwise show as a truncated JSON blob).
	if strings.Contains(out, `statusHistory: [`) {
		t.Fatalf("statusHistory should not be shown as a truncated key line, got:\n%s", out)
	}
}

func TestPrintObjectTable_TruncatesLongAutoReopenNote(t *testing.T) {
	// Force a narrow terminal so the truncation path runs deterministically.
	t.Setenv("COLUMNS", "80")

	longNote := "[auto-reopen]; prev=Confirmed; source=BackendAuto; exception=System.Exception; route=GET /api/very/long/path/" + strings.Repeat("x", 400)
	payload := map[string]any{
		"title": "t",
		"statusHistory": []map[string]any{
			{
				"changedAtUtc":       "2026-04-24T10:15:02Z",
				"fromStatus":         "Confirmed",
				"toStatus":           "New",
				"changedByUserId":    "system:auto-ingest",
				"changedByUserEmail": nil,
				"note":               longNote,
			},
		},
	}
	raw, _ := json.Marshal(payload)

	out := captureStdout(t, func() {
		_ = Print(FormatTable, raw)
	})

	// Head of the note must survive truncation so "[auto-reopen]" remains visible
	// as the first signal to a human scanning history output.
	if !strings.Contains(out, "[auto-reopen]") {
		t.Fatalf("truncation should preserve the leading [auto-reopen] tag, got:\n%s", out)
	}
	if !strings.Contains(out, "...") {
		t.Fatalf("long note should be truncated with ..., got:\n%s", out)
	}
	if strings.Contains(out, strings.Repeat("x", 400)) {
		t.Fatalf("truncation did not trim the tail, got:\n%s", out)
	}
}

func TestPrintObjectTable_EmptyHistoryRendersNone(t *testing.T) {
	payload := `{"title":"t","statusHistory":[]}`

	out := captureStdout(t, func() {
		_ = Print(FormatTable, json.RawMessage(payload))
	})

	if !strings.Contains(out, "History: (none)") {
		t.Fatalf("expected 'History: (none)' for empty history, got:\n%s", out)
	}
}

func TestPrintArrayTable_ListOutput_DoesNotExpandStatusHistory(t *testing.T) {
	// List endpoints return a paginated shape; each row carries statusHistory
	// inline. The list renderer MUST NOT render a standalone History: block
	// per row — that would turn a one-screen list into screens of noise.
	payload := `{"items":[
      {"externalGuid":"00000000-0000-0000-0000-000000000001","title":"A","statusHistory":[{"changedAtUtc":"2026-04-20T09:00:00Z","toStatus":"Validated","note":"manual triage"}]},
      {"externalGuid":"00000000-0000-0000-0000-000000000002","title":"B","statusHistory":[{"changedAtUtc":"2026-04-21T09:00:00Z","toStatus":"Fixed","note":"shipped"}]}
    ],"totalCount":2,"page":1,"pageSize":50}`

	out := captureStdout(t, func() {
		_ = Print(FormatTable, json.RawMessage(payload))
	})

	if strings.Contains(out, "History:") {
		t.Fatalf("list output must NOT render a History: block per row, got:\n%s", out)
	}
}

func TestPrintObjectTable_EmailFallsBackToUserId(t *testing.T) {
	payload := `{"title":"t","statusHistory":[
      {"changedAtUtc":"2026-04-24T10:00:00Z","fromStatus":"Validated","toStatus":"InProgress","changedByUserId":"user-abc","changedByUserEmail":"","note":"work started"}
    ]}`

	out := captureStdout(t, func() {
		_ = Print(FormatTable, json.RawMessage(payload))
	})

	if !strings.Contains(out, "user-abc") {
		t.Fatalf("empty email should fall back to changedByUserId, got:\n%s", out)
	}
}

func TestPrintObjectTable_FromToArrowRendered(t *testing.T) {
	payload := `{"title":"t","statusHistory":[
      {"changedAtUtc":"2026-04-24T10:00:00Z","fromStatus":"Validated","toStatus":"InProgress","changedByUserId":"eng@x","changedByUserEmail":"eng@uteamup.com","note":"n"}
    ]}`

	out := captureStdout(t, func() {
		_ = Print(FormatTable, json.RawMessage(payload))
	})

	if !strings.Contains(out, "Validated -> InProgress") {
		t.Fatalf("expected 'Validated -> InProgress' transition marker, got:\n%s", out)
	}
}

func TestPrintObjectTable_RendersTransitionAvailabilityAsADedicatedBlock(t *testing.T) {
	payload := `{
      "nonConformance":{"externalGuid":"11111111-1111-4111-8111-111111111111","status":"Triage"},
      "transitionAvailability":{
        "blockedReasonCode":"POLICY_PARTIAL",
        "blockedReason":"Some transitions require another role.",
        "options":[
          {
            "actionKey":"ncr.accept-triage",
            "toStatus":"Containment",
            "isAvailable":true,
            "requirements":{
              "requiresReason":true,
              "requiresEvidence":false,
              "requiresDecision":true,
              "requiresIndependentActor":false,
              "requiresAttestation":true
            },
            "eligibleRoles":[
              {"roleKey":"quality-triager","permissionKey":"NonConformance.Triage"}
            ],
            "approval":{
              "decisionKind":"Approve",
              "minimumApprovers":2,
              "requiredRoles":[
                {"roleKey":"quality-manager","permissionKey":"NonConformance.Disposition"}
              ]
            }
          },
          {
            "actionKey":"ncr.verify-close",
            "toStatus":"Closed",
            "isAvailable":false,
            "requirements":{
              "requiresReason":false,
              "requiresEvidence":true,
              "requiresDecision":false,
              "requiresIndependentActor":true,
              "requiresAttestation":false
            },
            "eligibleRoles":[],
            "blockedReasonCode":"EVIDENCE_REQUIRED",
            "blockedReason":"Attach retained verification evidence."
          }
        ]
      }
    }`

	out := captureStdout(t, func() {
		if err := Print(FormatTable, json.RawMessage(payload)); err != nil {
			t.Fatalf("Print: %v", err)
		}
	})

	wantFragments := []string{
		"Transition availability:",
		"blocked: POLICY_PARTIAL - Some transitions require another role.",
		"ncr.accept-triage -> Containment  [available]",
		"requirements: reason=yes, evidence=no, decision=yes, independentActor=no, attestation=yes",
		"eligible roles: quality-triager (NonConformance.Triage)",
		"approval: Approve, minimum approvers=2",
		"required roles: quality-manager (NonConformance.Disposition)",
		"ncr.verify-close -> Closed  [blocked]",
		"requirements: reason=no, evidence=yes, decision=no, independentActor=yes, attestation=no",
		"blocked: EVIDENCE_REQUIRED - Attach retained verification evidence.",
	}
	for _, fragment := range wantFragments {
		if !strings.Contains(out, fragment) {
			t.Errorf("table output is missing %q:\n%s", fragment, out)
		}
	}
	if strings.Contains(out, "transitionAvailability:") {
		t.Errorf("transitionAvailability should not be rendered as a truncated key/value row:\n%s", out)
	}
}

func TestTransitionAvailabilityRemainsOpaqueInJSONAndYAML(t *testing.T) {
	payload := json.RawMessage(`{
      "transitionAvailability":{
        "options":[{"actionKey":"ncr.submit","toStatus":"Submitted","isAvailable":true}]
      }
    }`)

	jsonOut := captureStdout(t, func() {
		if err := Print(FormatJSON, payload); err != nil {
			t.Fatalf("JSON Print: %v", err)
		}
	})
	if !strings.Contains(jsonOut, `"transitionAvailability": {`) ||
		!strings.Contains(jsonOut, `"actionKey": "ncr.submit"`) {
		t.Errorf("JSON output changed the response shape:\n%s", jsonOut)
	}
	if strings.Contains(jsonOut, "Transition availability:") {
		t.Errorf("JSON output received human table decoration:\n%s", jsonOut)
	}

	yamlOut := captureStdout(t, func() {
		if err := Print(FormatYAML, payload); err != nil {
			t.Fatalf("YAML Print: %v", err)
		}
	})
	if !strings.Contains(yamlOut, "transitionAvailability:") ||
		!strings.Contains(yamlOut, "actionKey: ncr.submit") {
		t.Errorf("YAML output changed the response shape:\n%s", yamlOut)
	}
	if strings.Contains(yamlOut, "Transition availability:") {
		t.Errorf("YAML output received human table decoration:\n%s", yamlOut)
	}
}
