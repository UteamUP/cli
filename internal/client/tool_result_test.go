package client

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeToolResultUnwrapsJSONText(t *testing.T) {
	payload := json.RawMessage(`{
		"content": [{
			"type": "text",
			"text": "{\"capabilityKey\":\"service.billing.review\",\"canChangeInvoices\":false}"
		}]
	}`)

	result, err := NormalizeToolResult(payload)
	if err != nil {
		t.Fatalf("NormalizeToolResult() error = %v", err)
	}
	if string(result) != `{"capabilityKey":"service.billing.review","canChangeInvoices":false}` {
		t.Fatalf("NormalizeToolResult() = %s", result)
	}
}

func TestNormalizeToolResultAcceptsSSEJSONRPCEnvelope(t *testing.T) {
	payload := json.RawMessage(`{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"content": [{
				"type": "text",
				"text": "{\"capabilityKey\":\"maintenance.due.explain\"}"
			}]
		}
	}`)

	result, err := NormalizeToolResult(payload)
	if err != nil {
		t.Fatalf("NormalizeToolResult() error = %v", err)
	}
	if string(result) != `{"capabilityKey":"maintenance.due.explain"}` {
		t.Fatalf("NormalizeToolResult() = %s", result)
	}
}

func TestNormalizeToolResultBoundsToolErrors(t *testing.T) {
	payload, err := json.Marshal(ToolCallResult{
		Content: []ToolCallContent{{Type: "text", Text: strings.Repeat("x", 2500)}},
		IsError: true,
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	_, normalizeErr := NormalizeToolResult(payload)
	if normalizeErr == nil {
		t.Fatal("NormalizeToolResult() error = nil, want bounded tool error")
	}
	if len([]rune(normalizeErr.Error())) > 2050 {
		t.Fatalf("NormalizeToolResult() error is not bounded: %d runes", len([]rune(normalizeErr.Error())))
	}
}

func TestNormalizeToolResultLeavesDirectJSONUnchanged(t *testing.T) {
	payload := json.RawMessage(`{"status":"Completed"}`)

	result, err := NormalizeToolResult(payload)
	if err != nil {
		t.Fatalf("NormalizeToolResult() error = %v", err)
	}
	if string(result) != string(payload) {
		t.Fatalf("NormalizeToolResult() = %s, want %s", result, payload)
	}
}
