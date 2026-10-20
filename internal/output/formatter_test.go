package output

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"Json", FormatJSON},
		{"yaml", FormatYAML},
		{"YAML", FormatYAML},
		{"yml", FormatYAML},
		{"table", FormatTable},
		{"", FormatTable},
		{"unknown", FormatTable},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ParseFormat(tc.input)
			if result != tc.expected {
				t.Errorf("ParseFormat(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestPrintJSONEscapesTerminalControlsWhenRawJSONIsInvalid(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = write
	t.Cleanup(func() { os.Stdout = previous })

	if err := Print(FormatJSON, json.RawMessage("{\x1b[31m}")); err != nil {
		t.Fatal(err)
	}
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = previous
	output, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsRune(string(output), '\x1b') || !strings.Contains(string(output), `\u001b`) {
		t.Fatalf("raw terminal control was not escaped: %q", output)
	}
}
