package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

func printTable(data json.RawMessage) error {
	if data == nil || string(data) == "null" {
		fmt.Println("(no data)")
		return nil
	}

	// Try as array first
	var arr []map[string]any
	if err := json.Unmarshal(data, &arr); err == nil && len(arr) > 0 {
		return printArrayTable(arr)
	}

	// Try as paginated response (common pattern: {items: [...], totalCount: N})
	var paginated struct {
		Items      []map[string]any `json:"items"`
		TotalCount int              `json:"totalCount"`
		Page       int              `json:"page"`
		PageSize   int              `json:"pageSize"`
	}
	if err := json.Unmarshal(data, &paginated); err == nil && paginated.Items != nil {
		if err := printArrayTable(paginated.Items); err != nil {
			return err
		}
		if paginated.TotalCount > 0 {
			fmt.Printf("\nShowing %d of %d total (page %d)\n", len(paginated.Items), paginated.TotalCount, paginated.Page)
		}
		return nil
	}

	// Try as single object
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err == nil {
		return printObjectTable(obj)
	}

	// Fall back to raw output
	fmt.Println(string(data))
	return nil
}

func printArrayTable(items []map[string]any) error {
	if len(items) == 0 {
		fmt.Println("(no items)")
		return nil
	}

	// Collect column names from first item
	columns := collectColumns(items[0])

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = strings.ToUpper(col)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Separator
	seps := make([]string, len(columns))
	for i, h := range headers {
		seps[i] = strings.Repeat("-", len(h))
	}
	fmt.Fprintln(w, strings.Join(seps, "\t"))

	// Rows
	for _, item := range items {
		values := make([]string, len(columns))
		for i, col := range columns {
			values[i] = formatValue(item[col])
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}

	return w.Flush()
}

func printObjectTable(obj map[string]any) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	keys := collectColumns(obj)
	for _, key := range keys {
		fmt.Fprintf(w, "%s:\t%s\n", key, formatValue(obj[key]))
	}

	return w.Flush()
}

func collectColumns(obj map[string]any) []string {
	// Prioritize common columns first, then alphabetical
	priority := []string{"id", "name", "title", "status", "type", "createdAt", "updatedAt"}
	seen := make(map[string]bool)
	var columns []string

	for _, col := range priority {
		if _, ok := obj[col]; ok {
			columns = append(columns, col)
			seen[col] = true
		}
	}

	var remaining []string
	for key := range obj {
		if !seen[key] {
			remaining = append(remaining, key)
		}
	}
	sort.Strings(remaining)
	columns = append(columns, remaining...)

	return columns
}

func formatValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		if len(val) > 60 {
			return val[:57] + "..."
		}
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case bool:
		if val {
			return "yes"
		}
		return "no"
	case map[string]any, []any:
		b, _ := json.Marshal(val)
		s := string(b)
		if len(s) > 60 {
			return s[:57] + "..."
		}
		return s
	default:
		return fmt.Sprintf("%v", val)
	}
}
