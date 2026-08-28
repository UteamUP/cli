package registry

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

func TestQualityDomainMirrorsGovernanceAndCleanupTools(t *testing.T) {
	t.Parallel()
	domain := findQualityDomain(t)
	if domain.APIPath != "/api/quality/governance" {
		t.Fatalf("APIPath = %q, want /api/quality/governance", domain.APIPath)
	}
	if len(domain.Actions) != 5 {
		t.Fatalf("actions = %d, want two governance and three cleanup actions", len(domain.Actions))
	}

	list := findQualityAction(t, domain, "list")
	if list.ToolName != "UteamupQualityRecordsSearch" || list.HTTPMethod != "GET" {
		t.Errorf("list action = %+v, want the Quality search GET tool", list)
	}
	get := findQualityAction(t, domain, "get")
	if get.ToolName != "UteamupQualityRecordLedgerGet" || get.HTTPMethod != "GET" {
		t.Errorf("get action = %+v, want the Quality ledger GET tool", get)
	}

	wantCleanupTools := map[string]string{
		"cleanup-deadletters":        "UteamupQualityCleanupDeadLettersList",
		"cleanup-deadletter-get":     "UteamupQualityCleanupDeadLetterGet",
		"cleanup-deadletter-redrive": "UteamupQualityCleanupDeadLetterRedrive",
	}
	for actionName, toolName := range wantCleanupTools {
		action := findQualityAction(t, domain, actionName)
		if action.ToolName != toolName {
			t.Errorf("%s ToolName = %q, want %q", actionName, action.ToolName, toolName)
		}
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

func TestQualityCleanupRoutesMatchTheOperationsController(t *testing.T) {
	t.Parallel()
	domain := findQualityDomain(t)
	const cleanupJobGUID = "11111111-2222-4333-8444-555555555555"

	list := findQualityAction(t, domain, "cleanup-deadletters")
	listPath, consumed := buildRESTPath(domain, list, nil)
	if listPath != "/api/quality/operations/storage-cleanup/dead-letters" || len(consumed) != 0 {
		t.Errorf("cleanup list route = %q consumed=%v", listPath, consumed)
	}

	get := findQualityAction(t, domain, "cleanup-deadletter-get")
	getPath, consumed := buildRESTPath(domain, get, map[string]any{"cleanupJobGuid": cleanupJobGUID})
	if getPath != "/api/quality/operations/storage-cleanup/dead-letters/"+cleanupJobGUID {
		t.Errorf("cleanup get route = %q", getPath)
	}
	if len(consumed) != 1 || consumed[0] != "cleanupJobGuid" {
		t.Errorf("cleanup get consumed args = %v", consumed)
	}

	redrive := findQualityAction(t, domain, "cleanup-deadletter-redrive")
	redrivePath, consumed := buildRESTPath(domain, redrive, map[string]any{"cleanupJobGuid": cleanupJobGUID})
	if redrivePath != "/api/quality/operations/storage-cleanup/dead-letters/"+cleanupJobGUID+"/redrive" {
		t.Errorf("cleanup redrive route = %q", redrivePath)
	}
	if redrive.HTTPMethod != "POST" || len(consumed) != 1 || consumed[0] != "cleanupJobGuid" {
		t.Errorf("cleanup redrive contract = method %q consumed=%v", redrive.HTTPMethod, consumed)
	}
}

func TestQualityCleanupListPaginationUsesQueryParameters(t *testing.T) {
	t.Parallel()
	action := findQualityAction(t, findQualityDomain(t), "cleanup-deadletters")
	want := map[string]string{"page": "page", "page-size": "pageSize"}
	for _, flag := range action.Flags {
		if flag.QueryName != want[flag.Name] {
			t.Errorf("%s QueryName = %q, want %q", flag.Name, flag.QueryName, want[flag.Name])
		}
		delete(want, flag.Name)
	}
	if len(want) != 0 {
		t.Errorf("missing cleanup pagination flags: %v", want)
	}
}

func TestQualityCleanupRedriveRequiresReviewedBodyAndHeaders(t *testing.T) {
	t.Parallel()
	action := findQualityAction(t, findQualityDomain(t), "cleanup-deadletter-redrive")
	want := map[string]FlagDef{
		"preview-token":     {BodyName: "previewToken", Required: true, Sensitive: true, Type: "string"},
		"reason":            {BodyName: "reason", Required: true, Type: "string"},
		"confirm":           {BodyName: "confirmed", Required: true, Type: "bool", MustBeTrue: true},
		"idempotency-key":   {HeaderName: "Idempotency-Key", Required: true, Type: "uuid"},
		"concurrency-token": {HeaderName: "If-Match", Required: true, Sensitive: true, Type: "string"},
	}
	for _, flag := range action.Flags {
		expected, ok := want[flag.Name]
		if !ok {
			t.Errorf("unexpected redrive flag %q", flag.Name)
			continue
		}
		if flag.BodyName != expected.BodyName || flag.HeaderName != expected.HeaderName ||
			flag.Required != expected.Required || flag.Sensitive != expected.Sensitive ||
			flag.Type != expected.Type || flag.MustBeTrue != expected.MustBeTrue {
			t.Errorf("redrive --%s = %+v, want contract %+v", flag.Name, flag, expected)
		}
		delete(want, flag.Name)
	}
	if len(want) != 0 {
		t.Errorf("missing redrive flags: %v", want)
	}
}

func TestQualityCleanupRedriveValidationFailsClosed(t *testing.T) {
	t.Parallel()
	domain := findQualityDomain(t)
	action := findQualityAction(t, domain, "cleanup-deadletter-redrive")
	const validGUID = "11111111-2222-4333-8444-555555555555"

	newValidCommand := func(t *testing.T) *cobra.Command {
		t.Helper()
		command := buildActionCommand(domain, action, nil, nil, nil, nil)
		validFlags := map[string]string{
			"preview-token":     "preview-token",
			"reason":            "Transient Azure provider outage",
			"confirm":           "true",
			"idempotency-key":   validGUID,
			"concurrency-token": "concurrency-token",
		}
		for name, value := range validFlags {
			if err := command.Flags().Set(name, value); err != nil {
				t.Fatalf("set --%s: %v", name, err)
			}
		}
		return command
	}

	tests := []struct {
		name      string
		argument  string
		flag      string
		flagValue string
		wantError string
	}{
		{name: "invalid cleanup job GUID", argument: "not-a-guid", wantError: "cleanupJobGuid"},
		{name: "blank preview token", argument: validGUID, flag: "preview-token", flagValue: " ", wantError: "--preview-token"},
		{name: "blank reason", argument: validGUID, flag: "reason", flagValue: " ", wantError: "--reason"},
		{name: "confirmation disabled", argument: validGUID, flag: "confirm", flagValue: "false", wantError: "--confirm"},
		{name: "invalid idempotency GUID", argument: validGUID, flag: "idempotency-key", flagValue: "retry-1", wantError: "--idempotency-key"},
		{name: "blank concurrency token", argument: validGUID, flag: "concurrency-token", flagValue: " ", wantError: "--concurrency-token"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := newValidCommand(t)
			if test.flag != "" {
				if err := command.Flags().Set(test.flag, test.flagValue); err != nil {
					t.Fatalf("set --%s: %v", test.flag, err)
				}
			}
			err := validateActionInput(command, []string{test.argument}, action)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("validation error = %v, want %q", err, test.wantError)
			}
		})
	}

	if err := validateActionInput(newValidCommand(t), []string{validGUID}, action); err != nil {
		t.Fatalf("valid reviewed redrive rejected: %v", err)
	}
}

func TestQualityBoundaryContainsNoIntegerIdentifiers(t *testing.T) {
	t.Parallel()
	domain := findQualityDomain(t)
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Type == "int" && strings.HasSuffix(strings.ToLower(arg.Name), "id") {
				t.Errorf("%s exposes integer-style identifier %s", action.Name, arg.Name)
			}
		}
	}
}
