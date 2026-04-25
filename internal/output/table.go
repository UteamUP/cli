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

// blockRenderedKeys are JSON fields that get their own dedicated block under
// the key/value table instead of being inline-formatted (which would truncate
// them to 60 chars and lose all signal). Each key here has a matching printer
// invoked after the table flush. Order matters: blocks render in this order
// so the admin/skill reads "what the user did" → "what the app did" → "what
// the database did" → audit trail.
var blockRenderedKeys = map[string]bool{
	"userActions":         true,
	"additionalNotes":     true,
	"recentApiCalls":      true,
	"recentStoreActions":  true,
	"involvedSourceFiles": true,
	"recentSqlCommands":   true,
	"statusHistory":       true,
}

func printObjectTable(obj map[string]any) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	keys := collectColumns(obj)
	for _, key := range keys {
		// Block-rendered keys are printed below the inline table so long lists
		// and free-form text don't get truncated to 57 chars inline.
		if blockRenderedKeys[key] {
			continue
		}
		fmt.Fprintf(w, "%s:\t%s\n", key, formatValue(obj[key]))
	}

	if err := w.Flush(); err != nil {
		return err
	}

	// Order matches blockRenderedKeys docstring: user actions first (most
	// directly answers "what was the user doing"), then app telemetry, then SQL,
	// then the audit history.
	if v, ok := obj["userActions"]; ok {
		printUserActionsBlock(v)
	}
	if v, ok := obj["additionalNotes"]; ok {
		printAdditionalNotesBlock(v)
	}
	if v, ok := obj["recentApiCalls"]; ok {
		printRecentApiCallsBlock(v)
	}
	if v, ok := obj["recentStoreActions"]; ok {
		printRecentStoreActionsBlock(v)
	}
	if v, ok := obj["involvedSourceFiles"]; ok {
		printStringListBlock("Source files involved", v)
	}
	if v, ok := obj["recentSqlCommands"]; ok {
		printStringListBlock("Recent SQL commands", v)
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

// printUserActionsBlock renders the per-page user-action trail as a numbered
// list of "user → <verb> <target>" lines. Each entry is `{type, target,
// oldValue?, newValue?, ts}`. Skipped silently when the field is null/empty
// so unrelated `bugs get` calls don't grow noisy.
func printUserActionsBlock(v any) {
	entries, ok := v.([]any)
	if !ok || len(entries) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("User actions:")
	for i, e := range entries {
		m, ok := e.(map[string]any)
		if !ok {
			continue
		}
		typ := stringVal(m["type"])
		target := stringVal(m["target"])
		old := stringVal(m["oldValue"])
		newv := stringVal(m["newValue"])
		line := fmt.Sprintf("user → %s %s", typ, target)
		if old != "" || newv != "" {
			if old != "" && newv != "" {
				line += fmt.Sprintf(" → %s → %s", old, newv)
			} else if newv != "" {
				line += fmt.Sprintf(" → %s", newv)
			}
		}
		fmt.Printf("  %d. %s\n", i+1, line)
	}
}

// printAdditionalNotesBlock renders the freeform admin notes as an indented
// paragraph below the inline table. Skipped when null or empty.
func printAdditionalNotesBlock(v any) {
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return
	}
	fmt.Println()
	fmt.Println("Additional notes:")
	for _, line := range strings.Split(s, "\n") {
		fmt.Printf("  %s\n", line)
	}
}

// printRecentApiCallsBlock renders the API trail as one line per call:
// "<status> <method> <endpoint>  <duration>ms  (called from <file>)".
// Failed calls (status >= 400) are prefixed with "!" so they jump out in a
// scan. Skipped when null or empty.
func printRecentApiCallsBlock(v any) {
	entries, ok := v.([]any)
	if !ok || len(entries) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("Recent API calls:")
	for _, e := range entries {
		m, ok := e.(map[string]any)
		if !ok {
			continue
		}
		statusStr := stringVal(m["status"])
		marker := " "
		if n, ok := m["status"].(float64); ok && n >= 400 {
			marker = "!"
		}
		fmt.Printf("  %s %s %s %s  %sms",
			marker,
			statusStr,
			stringVal(m["method"]),
			stringVal(m["endpoint"]),
			stringVal(m["durationMs"]),
		)
		if from := stringVal(m["calledFromFile"]); from != "" {
			fmt.Printf("  (from %s)", from)
		}
		fmt.Println()
	}
}

// printRecentStoreActionsBlock renders the Pinia store-action trail as one
// line per invocation: "<outcome> <store>/<action>  <duration>ms".
// Skipped when null or empty.
func printRecentStoreActionsBlock(v any) {
	entries, ok := v.([]any)
	if !ok || len(entries) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("Recent store actions:")
	for _, e := range entries {
		m, ok := e.(map[string]any)
		if !ok {
			continue
		}
		outcome := stringVal(m["outcome"])
		marker := " "
		if outcome == "error" {
			marker = "!"
		}
		fmt.Printf("  %s %s %s/%s",
			marker,
			outcome,
			stringVal(m["store"]),
			stringVal(m["action"]),
		)
		if d := stringVal(m["durationMs"]); d != "" && d != "<nil>" {
			fmt.Printf("  %sms", d)
		}
		fmt.Println()
	}
}

// printStringListBlock renders an array of strings as a numbered list under
// the given heading. Used for `involvedSourceFiles` and `recentSqlCommands`.
// Skipped when null or empty.
func printStringListBlock(heading string, v any) {
	entries, ok := v.([]any)
	if !ok || len(entries) == 0 {
		return
	}
	fmt.Println()
	fmt.Printf("%s:\n", heading)
	for i, e := range entries {
		s, ok := e.(string)
		if !ok {
			s = fmt.Sprintf("%v", e)
		}
		fmt.Printf("  %d. %s\n", i+1, s)
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
