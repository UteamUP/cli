package registry

import (
	"testing"
)

func findStockDomain(t *testing.T) *Domain {
	t.Helper()
	for _, dom := range DefaultRegistry.Domains() {
		if dom.Name == "stock" {
			return dom
		}
	}
	t.Fatal("expected stock domain to be registered")
	return nil
}

func findStockAction(t *testing.T, name string) *Action {
	t.Helper()
	d := findStockDomain(t)
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected `%s` action on stock domain", name)
	return nil
}

func TestStockDomainRegistered(t *testing.T) {
	d := findStockDomain(t)
	if len(d.Aliases) != 1 || d.Aliases[0] != "stocks" {
		t.Errorf("stock domain aliases = %+v, want [stocks]", d.Aliases)
	}
}

func TestStockSearchActionWired(t *testing.T) {
	action := findStockAction(t, "search")

	if action.ToolName != "UteamupStockSearchItems" {
		t.Errorf("search ToolName = %q, want %q", action.ToolName, "UteamupStockSearchItems")
	}
	if action.RESTPath != "items/search" {
		t.Errorf("search RESTPath = %q, want %q (the backend route is items/search)", action.RESTPath, "items/search")
	}
	if len(action.Args) != 0 {
		t.Errorf("search expected no positional args (q is an optional flag), got %+v", action.Args)
	}

	expectedFlags := map[string]string{
		"q":          "string",
		"type":       "string",
		"stock-guid": "string",
		"page":       "int",
		"page-size":  "int",
	}
	gotFlags := make(map[string]string)
	for _, f := range action.Flags {
		gotFlags[f.Name] = f.Type
	}
	for name, ty := range expectedFlags {
		got, ok := gotFlags[name]
		if !ok {
			t.Errorf("search action missing expected flag %q", name)
			continue
		}
		if got != ty {
			t.Errorf("search action flag %q type = %q, want %q", name, got, ty)
		}
	}
}

func TestStockAlertsActionWired(t *testing.T) {
	action := findStockAction(t, "alerts")

	if action.ToolName != "UteamupStockListAlerts" {
		t.Errorf("alerts ToolName = %q, want %q", action.ToolName, "UteamupStockListAlerts")
	}
	if action.RESTPath != "locations/{stockGuid}/alerts" {
		t.Errorf("alerts RESTPath = %q, want %q", action.RESTPath, "locations/{stockGuid}/alerts")
	}

	// stockGuid arrives via a required flag whose camelCase body name feeds the
	// {stockGuid} path placeholder.
	var stockGuid *FlagDef
	for i := range action.Flags {
		if action.Flags[i].Name == "stock-guid" {
			stockGuid = &action.Flags[i]
			break
		}
	}
	if stockGuid == nil {
		t.Fatal("alerts action must expose a `stock-guid` flag")
	}
	if !stockGuid.Required {
		t.Error("stock-guid flag must be Required (it fills the path placeholder)")
	}
	if stockGuid.Type != "string" {
		t.Errorf("stock-guid flag type = %q, want string (Guids are strings)", stockGuid.Type)
	}
}

func TestStockAckAlertActionWired(t *testing.T) {
	action := findStockAction(t, "ack-alert")

	if action.ToolName != "UteamupStockAcknowledgeAlert" {
		t.Errorf("ack-alert ToolName = %q, want %q", action.ToolName, "UteamupStockAcknowledgeAlert")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("ack-alert HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "alerts/{alertGuid}/acknowledge" {
		t.Errorf("ack-alert RESTPath = %q, want %q", action.RESTPath, "alerts/{alertGuid}/acknowledge")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "alertGuid" {
		t.Fatalf("ack-alert expected single positional arg 'alertGuid', got %+v", action.Args)
	}
	if action.Args[0].Type != "string" || !action.Args[0].Required {
		t.Errorf("alertGuid arg must be a Required string Guid, got %+v", action.Args[0])
	}
}

func TestStockPurchaseOrderActionsWired(t *testing.T) {
	cases := []struct {
		name       string
		toolName   string
		httpMethod string
		restPath   string
		wantGuid   bool
	}{
		{"po-list", "UteamupStockListPurchaseOrders", "", "purchase-orders", false},
		{"po-get", "UteamupStockGetPurchaseOrder", "", "purchase-orders/{guid}", true},
		{"po-submit", "UteamupStockSubmitPurchaseOrder", "POST", "purchase-orders/{guid}/submit", true},
		{"po-approve", "UteamupStockApprovePurchaseOrder", "POST", "purchase-orders/{guid}/approve", true},
		{"po-cancel", "UteamupStockCancelPurchaseOrder", "POST", "purchase-orders/{guid}/cancel", true},
	}

	for _, tc := range cases {
		action := findStockAction(t, tc.name)
		if action.ToolName != tc.toolName {
			t.Errorf("%s ToolName = %q, want %q", tc.name, action.ToolName, tc.toolName)
		}
		if action.HTTPMethod != tc.httpMethod {
			t.Errorf("%s HTTPMethod = %q, want %q", tc.name, action.HTTPMethod, tc.httpMethod)
		}
		if action.RESTPath != tc.restPath {
			t.Errorf("%s RESTPath = %q, want %q", tc.name, action.RESTPath, tc.restPath)
		}
		if tc.wantGuid {
			if len(action.Args) != 1 || action.Args[0].Name != "guid" || !action.Args[0].Required || action.Args[0].Type != "string" {
				t.Errorf("%s expected single required string positional arg 'guid', got %+v", tc.name, action.Args)
			}
		}
	}
}

func TestStockPoListStatusFlag(t *testing.T) {
	action := findStockAction(t, "po-list")
	gotFlags := make(map[string]string)
	for _, f := range action.Flags {
		gotFlags[f.Name] = f.Type
	}
	for name, ty := range map[string]string{"status": "string", "page": "int", "page-size": "int"} {
		got, ok := gotFlags[name]
		if !ok {
			t.Errorf("po-list missing expected flag %q", name)
			continue
		}
		if got != ty {
			t.Errorf("po-list flag %q type = %q, want %q", name, got, ty)
		}
	}
}

func TestStockPoCancelReasonFlagAlwaysSendsBody(t *testing.T) {
	action := findStockAction(t, "po-cancel")
	var reason *FlagDef
	for i := range action.Flags {
		if action.Flags[i].Name == "reason" {
			reason = &action.Flags[i]
			break
		}
	}
	if reason == nil {
		t.Fatal("po-cancel must expose a `reason` flag")
	}
	// Default "" keeps a JSON body on the POST — the backend binds
	// [FromBody] CancelPurchaseOrderRequestModel and rejects an empty body.
	if reason.Default != "" {
		t.Errorf("reason Default = %v, want \"\" (always send a body)", reason.Default)
	}
}

func TestStockReorderPolicyActionsWired(t *testing.T) {
	get := findStockAction(t, "reorder-policy-get")
	if get.ToolName != "UteamupStockGetReorderPolicy" {
		t.Errorf("reorder-policy-get ToolName = %q, want %q", get.ToolName, "UteamupStockGetReorderPolicy")
	}
	if get.RESTPath != "reorder-policy" {
		t.Errorf("reorder-policy-get RESTPath = %q, want %q", get.RESTPath, "reorder-policy")
	}

	set := findStockAction(t, "reorder-policy-set")
	if set.ToolName != "UteamupStockUpdateReorderPolicy" {
		t.Errorf("reorder-policy-set ToolName = %q, want %q", set.ToolName, "UteamupStockUpdateReorderPolicy")
	}
	if set.HTTPMethod != "PUT" {
		t.Errorf("reorder-policy-set HTTPMethod = %q, want PUT", set.HTTPMethod)
	}
	if set.RESTPath != "reorder-policy" {
		t.Errorf("reorder-policy-set RESTPath = %q, want %q", set.RESTPath, "reorder-policy")
	}

	// Flag → backend DTO field mapping. BodyName carries the divergent names;
	// cron-schedule camelCases to cronSchedule without an override.
	wantBodyNames := map[string]string{
		"enabled":       "isEnabled",
		"auto-submit":   "autoSubmitEnabled",
		"auto-deduct":   "autoDeductOnWorkorderCompletionEnabled",
		"cron-schedule": "",
	}
	gotFlags := make(map[string]*FlagDef)
	for i := range set.Flags {
		gotFlags[set.Flags[i].Name] = &set.Flags[i]
	}
	for name, bodyName := range wantBodyNames {
		f, ok := gotFlags[name]
		if !ok {
			t.Errorf("reorder-policy-set missing expected flag %q", name)
			continue
		}
		if f.BodyName != bodyName {
			t.Errorf("reorder-policy-set flag %q BodyName = %q, want %q", name, f.BodyName, bodyName)
		}
		// Every flag carries a Default so the PUT always sends the full policy
		// payload (deterministic upsert; bools absent from JSON would read false
		// server-side anyway, but an explicit body is self-documenting).
		if f.Default == nil {
			t.Errorf("reorder-policy-set flag %q must declare a Default", name)
		}
	}
	if cron := gotFlags["cron-schedule"]; cron != nil && cron.Default != "0 3 * * *" {
		t.Errorf("cron-schedule Default = %v, want \"0 3 * * *\" (matches backend default)", cron.Default)
	}
}
