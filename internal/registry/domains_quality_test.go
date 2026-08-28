package registry

import (
	"strings"
	"testing"
)

func findQualityDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "quality" {
			return domain
		}
	}
	t.Fatal("quality domain is not registered")
	return nil
}

func findQualityAction(t *testing.T, domain *Domain, name string) Action {
	t.Helper()
	for _, action := range domain.Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("quality action %q is not registered", name)
	return Action{}
}

func TestQualityDomainMirrorsTheTwoGovernanceReadTools(t *testing.T) {
	t.Parallel()
	domain := findQualityDomain(t)
	if domain.APIPath != "/api/quality/governance" {
		t.Fatalf("APIPath = %q, want /api/quality/governance", domain.APIPath)
	}
	if len(domain.Actions) != 2 {
		t.Fatalf("actions = %d, want exactly two read-only actions", len(domain.Actions))
	}

	list := findQualityAction(t, domain, "list")
	if list.ToolName != "UteamupQualityRecordsSearch" || list.HTTPMethod != "GET" {
		t.Errorf("list action = %+v, want the Quality search GET tool", list)
	}
	get := findQualityAction(t, domain, "get")
	if get.ToolName != "UteamupQualityRecordLedgerGet" || get.HTTPMethod != "GET" {
		t.Errorf("get action = %+v, want the Quality ledger GET tool", get)
	}
}

func TestQualityRoutesMatchTheGovernanceController(t *testing.T) {
	t.Parallel()
	domain := findQualityDomain(t)
	list := findQualityAction(t, domain, "list")
	listPath, _ := buildRESTPath(domain, list, map[string]any{})
	if listPath != "/api/quality/governance/records" {
		t.Errorf("list path = %q", listPath)
	}

	get := findQualityAction(t, domain, "get")
	getPath, consumed := buildRESTPath(domain, get, map[string]any{
		"recordKind":       "NonConformance",
		"domainRecordGuid": "6f83f5f3-75fd-4d74-a086-9b8427ff5c34",
	})
	if getPath != "/api/quality/governance/records/NonConformance/6f83f5f3-75fd-4d74-a086-9b8427ff5c34" {
		t.Errorf("get path = %q", getPath)
	}
	if len(consumed) != 2 {
		t.Errorf("consumed args = %v, want record kind and public GUID", consumed)
	}
	if strings.Contains(getPath, "{") {
		t.Errorf("get path left an unexpanded placeholder: %s", getPath)
	}
}

func TestQualityListFiltersTravelOnlyAsQueryParameters(t *testing.T) {
	t.Parallel()
	domain := findQualityDomain(t)
	list := findQualityAction(t, domain, "list")
	wantQueries := map[string]string{
		"record-kind":      "recordKind",
		"site-guid":        "siteGuid",
		"retention-status": "retentionDispositionStatus",
		"integrity-status": "historyIntegrityStatus",
		"search":           "search",
		"page":             "page",
		"page-size":        "pageSize",
	}
	for _, flag := range list.Flags {
		want, ok := wantQueries[flag.Name]
		if !ok {
			t.Errorf("unexpected list flag %q", flag.Name)
			continue
		}
		if flag.QueryName != want {
			t.Errorf("%s QueryName = %q, want %q", flag.Name, flag.QueryName, want)
		}
		delete(wantQueries, flag.Name)
	}
	if len(wantQueries) != 0 {
		t.Errorf("missing list query flags: %v", wantQueries)
	}
}

func TestQualityBoundaryContainsNoIntegerIdentifiers(t *testing.T) {
	t.Parallel()
	domain := findQualityDomain(t)
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if strings.HasSuffix(strings.ToLower(arg.Name), "id") {
				t.Errorf("%s exposes integer-style identifier %s", action.Name, arg.Name)
			}
		}
	}
}
