package registry

import (
	"testing"
)

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
