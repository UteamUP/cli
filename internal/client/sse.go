package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/uteamup/cli/internal/security"
	"io"
	"strings"
)

// SSEEvent represents a single Server-Sent Event.
type SSEEvent struct {
	Event string
	Data  string
}

// ParseSSE reads an SSE stream and returns all data payloads as parsed JSON.
func ParseSSE(reader io.Reader) ([]json.RawMessage, error) {
	var results []json.RawMessage
	bounded, err := security.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(bounded)))

	var currentData strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			currentData.WriteString(strings.TrimPrefix(line, "data: "))
		} else if strings.HasPrefix(line, "data:") {
			currentData.WriteString(strings.TrimPrefix(line, "data:"))
		} else if line == "" && currentData.Len() > 0 {
			// Empty line = event boundary
			data := strings.TrimSpace(currentData.String())
			currentData.Reset()

			if data == "[DONE]" {
				continue
			}

			if json.Valid([]byte(data)) {
				if len(results) >= 1024 {
					return nil, fmt.Errorf("SSE event limit exceeded")
				}
				results = append(results, json.RawMessage(data))
			}
		}
	}

	// Handle final event without trailing newline
	if currentData.Len() > 0 {
		data := strings.TrimSpace(currentData.String())
		if data != "[DONE]" && json.Valid([]byte(data)) {
			if len(results) >= 1024 {
				return nil, fmt.Errorf("SSE event limit exceeded")
			}
			results = append(results, json.RawMessage(data))
		}
	}

	if err := scanner.Err(); err != nil {
		return results, fmt.Errorf("reading SSE stream: %w", err)
	}

	return results, nil
}

// ExtractResult extracts the final result from SSE events.
// It returns the last event's data, which is typically the complete response.
func ExtractResult(events []json.RawMessage) json.RawMessage {
	if len(events) == 0 {
		return nil
	}
	return events[len(events)-1]
}
