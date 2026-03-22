package client

import (
	"strings"
	"testing"
)

func TestParseSSESingleEvent(t *testing.T) {
	input := `data: {"result": "hello"}

`
	events, err := ParseSSE(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if string(events[0]) != `{"result": "hello"}` {
		t.Errorf("unexpected data: %s", events[0])
	}
}

func TestParseSSEMultipleEvents(t *testing.T) {
	input := `data: {"id": 1}

data: {"id": 2}

data: {"id": 3}

`
	events, err := ParseSSE(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
}

func TestParseSSEDoneMarker(t *testing.T) {
	input := `data: {"result": "ok"}

data: [DONE]

`
	events, err := ParseSSE(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event (DONE skipped), got %d", len(events))
	}
}

func TestParseSSENoTrailingNewline(t *testing.T) {
	input := `data: {"final": true}`
	events, err := ParseSSE(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestParseSSEInvalidJSON(t *testing.T) {
	input := `data: not-json

data: {"valid": true}

`
	events, err := ParseSSE(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Invalid JSON should be skipped
	if len(events) != 1 {
		t.Fatalf("expected 1 valid event, got %d", len(events))
	}
}

func TestParseSSEEmpty(t *testing.T) {
	events, err := ParseSSE(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestExtractResult(t *testing.T) {
	input := `data: {"partial": true}

data: {"complete": true, "items": [1,2,3]}

`
	events, _ := ParseSSE(strings.NewReader(input))
	result := ExtractResult(events)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if string(result) != `{"complete": true, "items": [1,2,3]}` {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestExtractResultEmpty(t *testing.T) {
	result := ExtractResult(nil)
	if result != nil {
		t.Errorf("expected nil result, got %s", result)
	}
}
