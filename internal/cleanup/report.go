package cleanup

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// UsageRow mirrors the subset of the backend UsageStatResponse the diff needs.
type UsageRow struct {
	Type         string `json:"type"`
	Key          string `json:"key"`
	HitCount     int64  `json:"hitCount"`
	LastSeenDate string `json:"lastSeenDate"`
}

var reportTypeOrder = []string{
	TypeBackendEndpoint, TypeBackendRepository,
	TypeFrontendPage, TypeFrontendComponent,
	TypeMobilePage, TypeMobileComponent,
}

// Opt-in types are only meaningfully "unused" for the units that are actually instrumented.
var optInTypes = map[string]bool{TypeFrontendComponent: true, TypeMobileComponent: true}

// TypeSummary is the per-type rollup for the report.
type TypeSummary struct {
	Type        string
	TotalInCode int
	Eligible    int // instrumented subset for opt-in types; == TotalInCode otherwise
	Used        int
	Unused      int
	NotTracked  int
	UnusedKeys  []CatalogEntry
	ReverseDiff []string
	Unreliable  bool
}

// ReportInput bundles everything BuildSummaries / WriteMarkdown need.
type ReportInput struct {
	Catalog      Catalog
	Usage        []UsageRow
	Env          string
	EnabledSince *time.Time
	Now          time.Time
	MinDays      int
	FilterType   string
}

// BuildSummaries diffs the scanned catalog against runtime usage.
func BuildSummaries(in ReportInput) []TypeSummary {
	hits := map[string]map[string]int64{}       // type -> key -> hitCount
	runtimeKeys := map[string]map[string]bool{} // type -> set of runtime keys
	for _, r := range in.Usage {
		if hits[r.Type] == nil {
			hits[r.Type] = map[string]int64{}
			runtimeKeys[r.Type] = map[string]bool{}
		}
		hits[r.Type][r.Key] = r.HitCount
		runtimeKeys[r.Type][r.Key] = true
	}

	catalogByType := map[string][]CatalogEntry{}
	catalogKeys := map[string]map[string]bool{}
	for _, e := range in.Catalog.Entries {
		catalogByType[e.Type] = append(catalogByType[e.Type], e)
		if catalogKeys[e.Type] == nil {
			catalogKeys[e.Type] = map[string]bool{}
		}
		catalogKeys[e.Type][e.Key] = true
	}

	var out []TypeSummary
	for _, t := range reportTypeOrder {
		if in.FilterType != "" && in.FilterType != t {
			continue
		}
		entries := catalogByType[t]
		s := TypeSummary{Type: t, TotalInCode: len(entries)}

		for _, e := range entries {
			if optInTypes[t] && !e.Instrumented {
				s.NotTracked++
				continue
			}
			s.Eligible++
			if hits[t][e.Key] > 0 {
				s.Used++
			} else {
				s.Unused++
				s.UnusedKeys = append(s.UnusedKeys, e)
			}
		}

		// Reverse diff: runtime keys with no matching code unit (key-contract drift / dynamic routes).
		for k := range runtimeKeys[t] {
			if !catalogKeys[t][k] {
				s.ReverseDiff = append(s.ReverseDiff, k)
			}
		}
		sort.Strings(s.ReverseDiff)
		if n := len(runtimeKeys[t]); n > 0 && float64(len(s.ReverseDiff))/float64(n) > 0.02 {
			s.Unreliable = true
		}

		sort.Slice(s.UnusedKeys, func(i, j int) bool { return s.UnusedKeys[i].Key < s.UnusedKeys[j].Key })
		out = append(out, s)
	}
	return out
}

// WriteMarkdown renders the cleanup report to outPath.
func WriteMarkdown(in ReportInput, summaries []TypeSummary, outPath string) error {
	var b strings.Builder
	b.WriteString("# Cleanup Report\n\n")
	fmt.Fprintf(&b, "- Environment: **%s**\n", in.Env)
	fmt.Fprintf(&b, "- Generated: %s\n", in.Now.UTC().Format(time.RFC3339))

	windowOK := observationWindow(&b, in)

	if len(in.Catalog.Warnings) > 0 {
		b.WriteString("\n> ⚠️ Scanner warnings:\n")
		for _, w := range in.Catalog.Warnings {
			fmt.Fprintf(&b, "> - %s\n", w)
		}
	}

	b.WriteString("\n## Summary\n\n")
	b.WriteString("| Type | In code | Tracked | Used | Unused | % unused | Not tracked |\n")
	b.WriteString("|------|--------:|--------:|-----:|-------:|---------:|------------:|\n")
	for _, s := range summaries {
		pct := 0
		if s.Eligible > 0 {
			pct = int(float64(s.Unused) / float64(s.Eligible) * 100)
		}
		flag := ""
		if s.Unreliable {
			flag = " ⚠️"
		}
		fmt.Fprintf(&b, "| %s%s | %d | %d | %d | %d | %d%% | %d |\n",
			s.Type, flag, s.TotalInCode, s.Eligible, s.Used, s.Unused, pct, s.NotTracked)
	}

	if !windowOK {
		b.WriteString("\n> ⚠️ Observation window is shorter than the configured minimum — treat \"unused\" as **not yet trustworthy**.\n")
	}

	for _, s := range summaries {
		b.WriteString("\n## " + s.Type + "\n\n")
		if s.Unreliable {
			fmt.Fprintf(&b, "> ⚠️ %d runtime key(s) had no matching code unit (key-contract drift) — results for this type may be unreliable.\n\n", len(s.ReverseDiff))
		}
		if optInTypes[s.Type] && s.NotTracked > 0 {
			fmt.Fprintf(&b, "_%d %s(s) are not instrumented (no `v-usage` / tracker call) and are excluded from the unused count._\n\n", s.NotTracked, s.Type)
		}
		if len(s.UnusedKeys) == 0 {
			b.WriteString("_No unused units._\n")
		} else {
			b.WriteString("Unused (exists in code, never exercised):\n\n")
			for _, e := range s.UnusedKeys {
				fmt.Fprintf(&b, "- `%s` — %s\n", e.Key, e.File)
			}
		}
		if len(s.ReverseDiff) > 0 {
			b.WriteString("\nRuntime keys with no matching code unit (drift / dynamic):\n\n")
			for _, k := range s.ReverseDiff {
				fmt.Fprintf(&b, "- `%s`\n", k)
			}
		}
	}

	return os.WriteFile(outPath, []byte(b.String()), 0o644)
}

func observationWindow(b *strings.Builder, in ReportInput) bool {
	if in.EnabledSince == nil {
		b.WriteString("- Observation window: **tracking has never been enabled** — no usage data to trust yet.\n")
		return false
	}
	days := int(in.Now.Sub(*in.EnabledSince).Hours() / 24)
	fmt.Fprintf(b, "- Tracking enabled since: %s (%d day window)\n", in.EnabledSince.UTC().Format("2006-01-02"), days)
	return days >= in.MinDays
}
