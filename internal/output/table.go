package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
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
		// statusHistory is rendered as a dedicated History: block after the key/value
		// table so long [auto-reopen]; ... notes don't get truncated to 57 chars inline.
		if key == "statusHistory" {
			continue
		}
		fmt.Fprintf(w, "%s:\t%s\n", key, formatValue(obj[key]))
	}

	if err := w.Flush(); err != nil {
		return err
	}

	if hist, ok := obj["statusHistory"]; ok {
		printStatusHistoryBlock(hist)
	}

	return nil
}

// printStatusHistoryBlock renders the `statusHistory` array as a chronological
// block of one line per transition. Used by `bugs get` so a human (or the
// uteamup-debug skill) can see the full audit trail without following up with
// `-o json` and jq. Only intended for the single-object path; list output
// intentionally skips this to stay scannable.
func printStatusHistoryBlock(v any) {
	entries, ok := v.([]any)
	if !ok || len(entries) == 0 {
		fmt.Println()
		fmt.Println("History: (none)")
		return
	}

	rows := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		if m, ok := e.(map[string]any); ok {
			rows = append(rows, m)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return stringVal(rows[i]["changedAtUtc"]) < stringVal(rows[j]["changedAtUtc"])
	})

	// Reserve ~60 chars for ts + status arrow + author + separators.
	noteBudget := terminalWidth() - 60
	if noteBudget < 20 {
		noteBudget = 20
	}

	fmt.Println()
	fmt.Println("History:")
	for _, r := range rows {
		ts := stringVal(r["changedAtUtc"])
		from := stringVal(r["fromStatus"])
		to := stringVal(r["toStatus"])
		author := stringVal(r["changedByUserEmail"])
		if author == "" {
			author = stringVal(r["changedByUserId"])
		}
		note := stringVal(r["note"])
		if len(note) > noteBudget {
			note = note[:noteBudget-3] + "..."
		}
		fmt.Printf("  %s  %s -> %s  %s  %s\n", ts, from, to, author, note)
	}
}

func stringVal(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// terminalWidth returns the caller's terminal width in columns, preferring
// the COLUMNS env var and falling back to 160 so long notes don't wrap on
// desktop-sized terminals that don't export COLUMNS.
func terminalWidth() int {
	if col := os.Getenv("COLUMNS"); col != "" {
		if w, err := strconv.Atoi(col); err == nil && w > 40 {
			return w
		}
	}
	return 160
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
