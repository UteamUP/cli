package registry

import (
	"testing"
)

// --- Construction module domains ---

func TestConstructionDomainsRegistered(t *testing.T) {
	for _, name := range []string{"construction-issue", "rfi", "submittal", "dailylog", "construction-sheet"} {
		if findDomain(name) == nil {
			t.Errorf("expected %q domain to be registered", name)
		}
	}
}

func TestConstructionDomainAPIPaths(t *testing.T) {
	// Every construction domain pins APIPath to its REAL controller route.
	// Registry ToolName is metadata — the runtime calls REST, so a wrong or
	// missing path means every command silently 404s.
	cases := map[string]string{
		"construction-issue": "/api/constructionissue",
		"rfi":                "/api/rfi",
		"submittal":          "/api/submittal",
		"dailylog":           "/api/dailylog",
		"construction-sheet": "/api/constructionsheet",
	}
	for name, apiPath := range cases {
		d := findDomain(name)
		if d == nil {
			t.Errorf("expected %q domain to be registered", name)
			continue
		}
		if d.APIPath != apiPath {
			t.Errorf("%s: expected APIPath %q, got %q", name, apiPath, d.APIPath)
		}
	}
}

func TestConstructionDomainAliases(t *testing.T) {
	cases := map[string][]string{
		"construction-issue": {"construction-issues"},
		"rfi":                {"rfis"},
		"submittal":          {"submittals"},
		"dailylog":           {"dailylogs", "daily-log"},
		"construction-sheet": {"construction-sheets"},
	}
	for name, aliases := range cases {
		d := findDomain(name)
		if d == nil {
			t.Errorf("expected %q domain to be registered", name)
			continue
		}
		expected := make(map[string]bool, len(aliases))
		for _, alias := range aliases {
			expected[alias] = true
		}
		for _, alias := range d.Aliases {
			delete(expected, alias)
		}
		if len(expected) > 0 {
			t.Errorf("%s: missing aliases: %v", name, expected)
		}
	}
}

func TestConstructionDomainActions(t *testing.T) {
	// Every action maps to its exact backend MCP tool.
	cases := map[string]map[string]string{
		"construction-issue": {
			"list":       "UteamupConstructionIssueList",
			"get":        "UteamupConstructionIssueGet",
			"create":     "UteamupConstructionIssueCreate",
			"set-status": "UteamupConstructionIssueUpdateStatus",
		},
		"rfi": {
			"list":    "UteamupRfiList",
			"get":     "UteamupRfiGet",
			"respond": "UteamupRfiRespond",
		},
		"submittal": {
			"list":   "UteamupSubmittalList",
			"get":    "UteamupSubmittalGet",
			"review": "UteamupSubmittalReview",
		},
		"dailylog": {
			"list":   "UteamupDailylogList",
			"get":    "UteamupDailylogGet",
			"create": "UteamupDailylogCreate",
		},
		"construction-sheet": {
			"list": "UteamupConstructionSheetList",
		},
	}

	for domainName, expected := range cases {
		d := findDomain(domainName)
		if d == nil {
			t.Errorf("expected %q domain to be registered", domainName)
			continue
		}

		actionMap := make(map[string]string)
		for _, a := range d.Actions {
			actionMap[a.Name] = a.ToolName
		}

		if len(actionMap) != len(expected) {
			t.Errorf("%s: expected exactly %d actions, got %d (%v)", domainName, len(expected), len(actionMap), actionMap)
		}
		for name, tool := range expected {
			if actual, ok := actionMap[name]; !ok {
				t.Errorf("%s: missing action %q", domainName, name)
			} else if actual != tool {
				t.Errorf("%s action %q: expected tool %q, got %q", domainName, name, tool, actual)
			}
		}
	}
}

func TestConstructionGetActionsAreGuidFirst(t *testing.T) {
	// get must take the record's public GUID positional arg (string), never a
	// legacy integer id — GUIDs In, Integer IDs Out.
	for _, domainName := range []string{"construction-issue", "rfi", "submittal", "dailylog"} {
		a := findDomainAction(t, domainName, "get")
		if len(a.Args) != 1 {
			t.Errorf("%s get: expected 1 positional arg, got %d", domainName, len(a.Args))
			continue
		}
		if a.Args[0].Name != "externalGuid" {
			t.Errorf("%s get: expected positional arg %q, got %q", domainName, "externalGuid", a.Args[0].Name)
		}
		if a.Args[0].Type != "string" {
			t.Errorf("%s get: identity arg must be string (guid), got %q", domainName, a.Args[0].Type)
		}
		if !a.Args[0].Required {
			t.Errorf("%s get: identity arg must be required", domainName)
		}
	}
}

func TestConstructionSubRouteActionsWired(t *testing.T) {
	// Sub-route writes mirror the real controller routes:
	//   POST /api/constructionissue/{guid}/status
	//   POST /api/rfi/{guid}/respond
	//   POST /api/submittal/{guid}/review
	// The GUID rides the URL via the {externalGuid} placeholder; the reviewed
	// payload rides the JSON body via flags.
	cases := []struct {
		domain       string
		action       string
		tool         string
		restPath     string
		requiredFlag string
	}{
		{"construction-issue", "set-status", "UteamupConstructionIssueUpdateStatus", "{externalGuid}/status", "status"},
		{"rfi", "respond", "UteamupRfiRespond", "{externalGuid}/respond", "official-response"},
		{"submittal", "review", "UteamupSubmittalReview", "{externalGuid}/review", "review-status"},
	}

	for _, c := range cases {
		a := findDomainAction(t, c.domain, c.action)
		if a.ToolName != c.tool || a.HTTPMethod != "POST" || a.RESTPath != c.restPath {
			t.Errorf("%s %s: want tool=%s method=POST path=%s, got tool=%s method=%q path=%s",
				c.domain, c.action, c.tool, c.restPath, a.ToolName, a.HTTPMethod, a.RESTPath)
		}
		if len(a.Args) != 1 || a.Args[0].Name != "externalGuid" || !a.Args[0].Required || a.Args[0].Type != "string" {
			t.Errorf("%s %s: first arg must be required string 'externalGuid' (matching the RESTPath placeholder), got %+v",
				c.domain, c.action, a.Args)
		}
		found := false
		for _, f := range a.Flags {
			if f.Name == c.requiredFlag {
				found = true
				if !f.Required {
					t.Errorf("%s %s: flag %q must be required", c.domain, c.action, c.requiredFlag)
				}
			}
		}
		if !found {
			t.Errorf("%s %s: missing required flag %q", c.domain, c.action, c.requiredFlag)
		}
	}
}

func TestConstructionDailylogCreateRoutesToGetOrCreate(t *testing.T) {
	// create maps to the idempotent backend route POST /api/dailylog/get-or-create,
	// expressed as a literal RESTPath suffix (no placeholders, so no positional args).
	a := findDomainAction(t, "dailylog", "create")
	if a.RESTPath != "get-or-create" {
		t.Errorf("dailylog create: expected RESTPath %q, got %q", "get-or-create", a.RESTPath)
	}
	if a.HTTPMethod != "" {
		// Action name "create" already maps to POST via the HTTPMethod table;
		// an override here would hide a route change.
		t.Errorf("dailylog create: expected empty HTTPMethod (derived POST), got %q", a.HTTPMethod)
	}
	if len(a.Args) != 0 {
		t.Errorf("dailylog create: expected no positional args, got %d", len(a.Args))
	}

	required := map[string]bool{"project-guid": false, "log-date": false}
	for _, f := range a.Flags {
		if _, ok := required[f.Name]; ok {
			required[f.Name] = f.Required
		}
	}
	for name, ok := range required {
		if !ok {
			t.Errorf("dailylog create: flag %q must exist and be required", name)
		}
	}
}

func TestConstructionSheetDomainIsListOnly(t *testing.T) {
	// The sheet register is read-only from the CLI: sheets are created by the
	// review-first import/OCR flows, not by ad-hoc terminal writes.
	d := findDomain("construction-sheet")
	if d == nil {
		t.Fatal("expected construction-sheet domain to be registered")
	}
	if len(d.Actions) != 1 || d.Actions[0].Name != "list" {
		t.Errorf("construction-sheet should expose exactly one action (list), got %+v", d.Actions)
	}
}

func TestConstructionActionsCarryNoIdentityFlags(t *testing.T) {
	// The backend resolves the acting user from the API key. No construction
	// action may accept a user or tenant identity flag — adding one would make
	// the command misleading and invite spoofing.
	for _, domainName := range []string{"construction-issue", "rfi", "submittal", "dailylog", "construction-sheet"} {
		d := findDomain(domainName)
		if d == nil {
			t.Errorf("expected %q domain to be registered", domainName)
			continue
		}
		for _, a := range d.Actions {
			for _, f := range a.Flags {
				switch f.Name {
				case "user-id", "user-guid", "tenant-id", "tenant-guid":
					t.Errorf("%s %s: identity flag %q is forbidden (identity comes from the API key)",
						domainName, a.Name, f.Name)
				}
			}
		}
	}
}
