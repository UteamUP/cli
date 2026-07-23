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
