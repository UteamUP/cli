package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format represents an output format.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// ParseFormat converts a string to a Format, defaulting to table.
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "yaml", "yml":
		return FormatYAML
	default:
		return FormatTable
	}
}

// Print outputs data in the specified format.
func Print(format Format, data json.RawMessage) error {
	switch format {
	case FormatJSON:
		return printJSON(data)
	case FormatYAML:
		return printYAML(data)
	default:
		return printTable(data)
	}
}

func printJSON(data json.RawMessage) error {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		// Fall back to raw output
		fmt.Println(string(data))
		return nil
	}
	fmt.Println(pretty.String())
	return nil
}

func printYAML(data json.RawMessage) error {
	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("parsing JSON for YAML conversion: %w", err)
	}

	out, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	fmt.Print(string(out))
	return nil
}
