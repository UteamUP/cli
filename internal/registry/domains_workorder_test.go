package registry

import (
	"reflect"
	"testing"
)

func TestWorkorderGuidAndLookupRoutes(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("expected workorder domain to be registered")
	}

	tests := []struct {
		action string
		args   map[string]any
		want   string
		used   []string
	}{
		{
			action: "get",
			args:   map[string]any{"workorderGuid": "11111111-1111-4111-8111-111111111111"},
			want:   "/api/workorder/11111111-1111-4111-8111-111111111111",
			used:   []string{"workorderGuid"},
		},
		{
			action: "update",
			args:   map[string]any{"workorderGuid": "22222222-2222-4222-8222-222222222222"},
			want:   "/api/workorder/by-guid/22222222-2222-4222-8222-222222222222",
			used:   []string{"workorderGuid"},
		},
		{
			action: "delete",
			args:   map[string]any{"workorderGuid": "33333333-3333-4333-8333-333333333333"},
			want:   "/api/workorder/by-guid/33333333-3333-4333-8333-333333333333",
			used:   []string{"workorderGuid"},
		},
		{
			action: "complete",
			args:   map[string]any{"workorderGuid": "11111111-1111-4111-8111-111111111111"},
			want:   "/api/workorder/by-guid/11111111-1111-4111-8111-111111111111/complete",
			used:   []string{"workorderGuid"},
		},
		{
			action: "search",
			args:   map[string]any{"query": "pump"},
			want:   "/api/workorder/search",
		},
		{
			action: "by-code",
			args:   map[string]any{"codeBranch": "1-HLA"},
			want:   "/api/workorder/by-code/1-HLA",
			used:   []string{"codeBranch"},
		},
	}

	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			var action *Action
			for index := range d.Actions {
				if d.Actions[index].Name == test.action {
					action = &d.Actions[index]
					break
				}
			}
			if action == nil {
				t.Fatalf("expected %s action", test.action)
			}

			path, consumed := buildRESTPath(d, *action, test.args)
			if path != test.want {
				t.Fatalf("path = %q, want %q", path, test.want)
			}
			if !reflect.DeepEqual(consumed, test.used) {
				t.Fatalf("consumed = %v, want %v", consumed, test.used)
			}
		})
	}
}

func TestWorkorderGetUpdateDeleteMatchMCPGuidContracts(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("expected workorder domain to be registered")
	}

	expectedTools := map[string]string{
		"get":    "UteamupWorkorderGet",
		"update": "UteamupWorkorderUpdate",
		"delete": "UteamupWorkorderDelete",
	}

	for actionName, toolName := range expectedTools {
		var action *Action
		for i := range d.Actions {
			if d.Actions[i].Name == actionName {
				action = &d.Actions[i]
				break
			}
		}
		if action == nil {
			t.Fatalf("expected %s action", actionName)
		}
		if action.ToolName != toolName {
			t.Errorf("%s tool = %q, want %q", actionName, action.ToolName, toolName)
		}
		wantPath := "by-guid/{workorderGuid}"
		if actionName == "get" {
			wantPath = "{workorderGuid}"
		}
		if action.RESTPath != wantPath {
			t.Errorf("%s REST path = %q, want %q", actionName, action.RESTPath, wantPath)
		}
		if len(action.Args) != 1 || action.Args[0].Name != "workorderGuid" || action.Args[0].Type != "uuid" {
			t.Errorf("%s args = %+v, want one workorderGuid uuid", actionName, action.Args)
		}
	}
}

func TestWorkorderCompleteUsesGuidOnlyStatusFreeContract(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("expected workorder domain to be registered")
	}

	for i := range d.Actions {
		action := d.Actions[i]
		if action.Name != "complete" {
			continue
		}
		if action.ToolName != "UteamupWorkorderComplete" {
			t.Fatalf("complete tool = %q", action.ToolName)
		}
		if action.HTTPMethod != "POST" {
			t.Fatalf("complete method = %q, want POST", action.HTTPMethod)
		}
		if len(action.Args) != 1 || action.Args[0].Name != "workorderGuid" || action.Args[0].Type != "uuid" {
			t.Fatalf("complete args = %+v, want one workorderGuid uuid", action.Args)
		}
		if len(action.Flags) != 0 {
			t.Fatalf("complete flags = %+v, want no caller-selected status", action.Flags)
		}
		return
	}

	t.Fatal("expected complete action")
}

func TestWorkorderCreateUsesOneGuidSafeIdempotentMCPContract(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("expected workorder domain to be registered")
	}

	var create *Action
	for i := range d.Actions {
		action := &d.Actions[i]
		if action.Name == "create-by-guid" {
			t.Fatal("create-by-guid alias must be removed in favor of the canonical create action")
		}
		if action.Name == "create" {
			if create != nil {
				t.Fatal("expected exactly one create action")
			}
			create = action
		}
	}
	if create == nil {
		t.Fatal("expected create action")
	}
	if create.ToolName != "UteamupWorkorderCreate" {
		t.Fatalf("create tool = %q, want UteamupWorkorderCreate", create.ToolName)
	}
	if !create.MCPOnly {
		t.Fatal("create must call the scalar MCP contract instead of the multipart REST model")
	}
	if len(create.Args) != 0 {
		t.Fatalf("create args = %+v, want scalar named flags only", create.Args)
	}

	flags := make(map[string]FlagDef, len(create.Flags))
	for _, flag := range create.Flags {
		flags[flag.Name] = flag
	}
	for _, name := range []string{"title", "description", "start-utc", "due-utc", "idempotency-key"} {
		flag, ok := flags[name]
		if !ok || !flag.Required {
			t.Errorf("create flag --%s must exist and be required", name)
		}
	}
	for _, name := range []string{"asset-guid", "primary-assignee-guid"} {
		flag, ok := flags[name]
		if !ok || flag.Type != "uuid" || flag.Required {
			t.Errorf("create flag --%s = %+v, want optional uuid", name, flag)
		}
	}
	for _, legacy := range []string{"asset-id", "assigned-to", "from-json"} {
		if _, ok := flags[legacy]; ok {
			t.Errorf("legacy integer/model flag --%s must not be public", legacy)
		}
	}
	if priority := flags["priority"]; priority.Type != "int" || priority.Default != 3 {
		t.Errorf("priority = %+v, want int default 3", priority)
	}
}

func TestWorkorderSearchUsesBackendQueryParameter(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("expected workorder domain to be registered")
	}

	for _, action := range d.Actions {
		if action.Name != "search" {
			continue
		}
		if len(action.Args) != 1 || action.Args[0].QueryName != "query" {
			t.Fatalf("search query mapping = %+v, want positional query -> query string field query", action.Args)
		}
		return
	}

	t.Fatal("expected workorder search action")
}

// --- Workorder Quick Close action ---

func TestWorkorderDomainHasQuickCloseAction(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("expected workorder domain to be registered")
	}
	var a *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "quick-close" {
			a = &d.Actions[i]
			break
		}
	}
	if a == nil {
		t.Fatal("expected quick-close action on the workorder domain")
	}
	if a.ToolName != "UteamupWorkorderQuickClose" {
		t.Errorf("quick-close: expected tool UteamupWorkorderQuickClose, got %q", a.ToolName)
	}
}

// The list action must expose the asset-guid filter (kebab → camelCase assetGuid)
// so `ut workorder list --asset-guid <guid>` scopes to one asset's work orders —
// the watch NFC → asset → its workorders flow.
func TestWorkorderListHasAssetGuidFlag(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("expected workorder domain to be registered")
	}
	var list *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "list" {
			list = &d.Actions[i]
			break
		}
	}
	if list == nil {
		t.Fatal("expected list action on the workorder domain")
	}

	var f *FlagDef
	for i := range list.Flags {
		if list.Flags[i].Name == "asset-guid" {
			f = &list.Flags[i]
			break
		}
	}
	if f == nil {
		t.Fatal("expected list action to carry the asset-guid flag")
	}
	if f.Type != "string" {
		t.Errorf("asset-guid: expected type string, got %q", f.Type)
	}
	if f.Required {
		t.Error("asset-guid must be optional (filter), not required")
	}
}

// Quick Close must carry ONE tenant-scoped target (template + asset) plus the
// resolution note. Losing any of these three required flags would ship a
// command that always errors server-side — test the contract.
func TestWorkorderQuickCloseRequiredFlags(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("expected workorder domain to be registered")
	}
	var qc *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "quick-close" {
			qc = &d.Actions[i]
			break
		}
	}
	if qc == nil {
		t.Fatal("expected quick-close action")
	}

	// Required flag names per the backend contract.
	required := map[string]bool{
		"template": false,
		"asset":    false,
		"note":     false,
	}
	for _, f := range qc.Flags {
		if _, ok := required[f.Name]; ok {
			if !f.Required {
				t.Errorf("flag --%s must be marked Required", f.Name)
			}
			required[f.Name] = true
		}
	}
	for name, seen := range required {
		if !seen {
			t.Errorf("missing required flag --%s", name)
		}
	}
}

// The idempotency-key, industry-code and performed-at flags are all optional
// by design. If they become required the CLI would force callers to generate
// a GUID themselves — which defeats the usability of the command.
func TestWorkorderQuickCloseOptionalFlagsAreNotRequired(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("expected workorder domain to be registered")
	}
	var qc *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "quick-close" {
			qc = &d.Actions[i]
			break
		}
	}
	if qc == nil {
		t.Fatal("expected quick-close action")
	}

	mustBeOptional := map[string]bool{
		"idempotency-key": true,
		"industry-code":   true,
		"performed-at":    true,
	}
	for _, f := range qc.Flags {
		if mustBeOptional[f.Name] && f.Required {
			t.Errorf("flag --%s must be optional", f.Name)
		}
	}
}

// Quick Close must never accept positional args — every identifier is a GUID
// and would be painful to position-order. If positional args are introduced
// later the help text and UX break.
func TestWorkorderQuickCloseHasNoPositionalArgs(t *testing.T) {
	d := findDomain("workorder")
	if d == nil {
		t.Fatal("expected workorder domain to be registered")
	}
	for _, a := range d.Actions {
		if a.Name != "quick-close" {
			continue
		}
		if len(a.Args) != 0 {
			t.Errorf("quick-close should take no positional args, got %d", len(a.Args))
		}
		return
	}
	t.Fatal("expected quick-close action on the workorder domain")
}
