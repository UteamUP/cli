package registry

import (
	"testing"
)

// --- Journal domain ---

func TestJournalDomainRegistered(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}
}

func TestJournalDomainAliases(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}
	expected := map[string]bool{"journals": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestJournalDomainActions(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}

	expected := map[string]string{
		"by-code":           "UteamupJournalByCode",
		"by-asset":          "UteamupJournalByAsset",
		"import":            "UteamupJournalImport",
		"create-from-image": "UteamupJournalCreateFromImage",
		"search-assets":     "UteamupAssetMentionSearch",
		"search-workorders": "UteamupWorkorderMentionSearch",
	}

	actionMap := make(map[string]string)
	for _, a := range d.Actions {
		actionMap[a.Name] = a.ToolName
	}

	for name, tool := range expected {
		if actual, ok := actionMap[name]; !ok {
			t.Errorf("missing action %q", name)
		} else if actual != tool {
			t.Errorf("action %q: expected tool %q, got %q", name, tool, actual)
		}
	}
}

// Guards the required positional args on the new import action so that
// accidentally dropping file-name or file-content-base64 becomes a test
// failure rather than a silent CLI regression.
func TestJournalImportArgs(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}
	action := findAction(d, "import")
	if action == nil {
		t.Fatal("expected import action on journal domain")
	}
	if len(action.Args) < 2 {
		t.Fatalf("import action should have 2 positional args, got %d", len(action.Args))
	}
	if action.Args[0].Name != "file-name" || !action.Args[0].Required {
		t.Errorf("first arg should be required file-name, got %+v", action.Args[0])
	}
	if action.Args[1].Name != "file-content-base64" || !action.Args[1].Required {
		t.Errorf("second arg should be required file-content-base64, got %+v", action.Args[1])
	}
}

func TestJournalSearchAssetsAction(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}
	action := findAction(d, "search-assets")
	if action == nil {
		t.Fatal("expected search-assets action on journal domain")
	}
	if len(action.Args) == 0 || action.Args[0].Name != "query" || !action.Args[0].Required {
		t.Errorf("search-assets should take a required query arg, got %+v", action.Args)
	}
	flagMap := flagsToMap(action.Flags)
	limitFlag, ok := flagMap["limit"]
	if !ok {
		t.Fatal("search-assets missing limit flag")
	}
	if limitFlag.Default != 8 {
		t.Errorf("limit default should be 8, got %v", limitFlag.Default)
	}
}

func TestJournalByCodeArgsAndFlags(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}

	action := findAction(d, "by-code")
	if action == nil {
		t.Fatal("expected by-code action on journal domain")
	}

	// Must have required positional arg: code-catalog-entry-id
	if len(action.Args) == 0 {
		t.Fatal("by-code action should have at least one positional arg")
	}
	if action.Args[0].Name != "code-catalog-entry-id" {
		t.Errorf("expected arg name code-catalog-entry-id, got %q", action.Args[0].Name)
	}
	if !action.Args[0].Required {
		t.Error("code-catalog-entry-id arg should be required")
	}

	// Must have page and page-size flags
	flagMap := flagsToMap(action.Flags)
	if _, ok := flagMap["page"]; !ok {
		t.Error("by-code action missing page flag")
	}
	if _, ok := flagMap["page-size"]; !ok {
		t.Error("by-code action missing page-size flag")
	}
}

func TestJournalByAssetArgsAndFlags(t *testing.T) {
	d := findDomain("journal")
	if d == nil {
		t.Fatal("expected journal domain to be registered")
	}

	action := findAction(d, "by-asset")
	if action == nil {
		t.Fatal("expected by-asset action on journal domain")
	}

	// Must have required positional arg: asset-id
	if len(action.Args) == 0 {
		t.Fatal("by-asset action should have at least one positional arg")
	}
	if action.Args[0].Name != "asset-id" {
		t.Errorf("expected arg name asset-id, got %q", action.Args[0].Name)
	}
	if !action.Args[0].Required {
		t.Error("asset-id arg should be required")
	}

	// Must have page and page-size flags
	flagMap := flagsToMap(action.Flags)
	if _, ok := flagMap["page"]; !ok {
		t.Error("by-asset action missing page flag")
	}
	if _, ok := flagMap["page-size"]; !ok {
		t.Error("by-asset action missing page-size flag")
	}
}

// --- CodeCatalog domain ---

func TestCodeCatalogDomainRegistered(t *testing.T) {
	d := findDomain("codecatalog")
	if d == nil {
		t.Fatal("expected codecatalog domain to be registered")
	}
}

func TestCodeCatalogDomainAliases(t *testing.T) {
	d := findDomain("codecatalog")
	if d == nil {
		t.Fatal("expected codecatalog domain to be registered")
	}
	expected := map[string]bool{"cc": true, "codes": true}
	for _, alias := range d.Aliases {
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing aliases: %v", expected)
	}
}

func TestCodeCatalogSearchAction(t *testing.T) {
	d := findDomain("codecatalog")
	if d == nil {
		t.Fatal("expected codecatalog domain to be registered")
	}

	action := findAction(d, "search")
	if action == nil {
		t.Fatal("expected search action on codecatalog domain")
	}
	if action.ToolName != "UteamupCodeCatalogSearch" {
		t.Errorf("expected tool UteamupCodeCatalogSearch, got %q", action.ToolName)
	}
}

func TestCodeCatalogSearchArgAndFlags(t *testing.T) {
	d := findDomain("codecatalog")
	if d == nil {
		t.Fatal("expected codecatalog domain to be registered")
	}

	action := findAction(d, "search")
	if action == nil {
		t.Fatal("expected search action on codecatalog domain")
	}

	// Required positional arg: query
	if len(action.Args) == 0 {
		t.Fatal("search action should have at least one positional arg")
	}
	if action.Args[0].Name != "query" {
		t.Errorf("expected arg name query, got %q", action.Args[0].Name)
	}
	if !action.Args[0].Required {
		t.Error("query arg should be required")
	}

	// limit flag
	flagMap := flagsToMap(action.Flags)
	limitFlag, ok := flagMap["limit"]
	if !ok {
		t.Fatal("search action missing limit flag")
	}
	if limitFlag.Default != 10 {
		t.Errorf("limit default should be 10, got %v", limitFlag.Default)
	}
	if limitFlag.Required {
		t.Error("limit flag should not be required")
	}
}

// --- helpers ---

func findDomain(name string) *Domain {
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == name {
			return d
		}
	}
	return nil
}

func findAction(d *Domain, name string) *Action {
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	return nil
}

func flagsToMap(flags []FlagDef) map[string]FlagDef {
	m := make(map[string]FlagDef, len(flags))
	for _, f := range flags {
		m[f.Name] = f
	}
	return m
}
