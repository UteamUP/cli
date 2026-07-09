package registry

import (
	"os"
	"path/filepath"
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

func findChemicalAction(t *testing.T, name string) *Action {
	t.Helper()
	for _, dom := range DefaultRegistry.Domains() {
		if dom.Name != "chemical" {
			continue
		}
		for i := range dom.Actions {
			if dom.Actions[i].Name == name {
				return &dom.Actions[i]
			}
		}
		t.Fatalf("expected `%s` action on chemical domain", name)
	}
	t.Fatal("expected chemical domain to be registered")
	return nil
}

// Chemical is GUID-first. crudActions() would give it the legacy integer `id` arg and route to the
// now-[Obsolete] /api/chemical/{id}. And because Chemical's GUID routes are prefixed `by-guid/`, each
// identified action MUST declare an explicit RESTPath — without it buildRESTPath falls back to
// /api/chemical/{guid}, a route Chemical does not expose (codes does, chemical doesn't).
func TestChemicalDomainIsGuidFirst(t *testing.T) {
	for _, name := range []string{"get", "update", "delete"} {
		action := findChemicalAction(t, name)

		if len(action.Args) != 1 || action.Args[0].Name != "externalGuid" {
			t.Errorf("chemical %s args = %+v, want a single externalGuid positional (GUIDs-in rule)", name, action.Args)
		}
		if action.Args[0].Type != "string" {
			t.Errorf("chemical %s arg type = %q, want string (a GUID, never an int id)", name, action.Args[0].Type)
		}
		if action.RESTPath != "by-guid/{externalGuid}" {
			t.Errorf("chemical %s RESTPath = %q, want %q (backend exposes by-guid/, not /{guid})", name, action.RESTPath, "by-guid/{externalGuid}")
		}
	}

	// list/create take no identifier, so they must not have picked up a positional arg.
	for _, name := range []string{"list", "create"} {
		if action := findChemicalAction(t, name); len(action.Args) != 0 {
			t.Errorf("chemical %s expected no positional args, got %+v", name, action.Args)
		}
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

func stockActionFlag(t *testing.T, actionName, flagName string) *FlagDef {
	t.Helper()
	action := findStockAction(t, actionName)
	for i := range action.Flags {
		if action.Flags[i].Name == flagName {
			return &action.Flags[i]
		}
	}
	t.Fatalf("%s action must expose a `%s` flag", actionName, flagName)
	return nil
}

func TestStockTransferActionWired(t *testing.T) {
	action := findStockAction(t, "transfer")

	if action.ToolName != "TransferInventory" {
		t.Errorf("transfer ToolName = %q, want %q", action.ToolName, "TransferInventory")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("transfer HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "transfers" {
		t.Errorf("transfer RESTPath = %q, want %q", action.RESTPath, "transfers")
	}

	// GUIDs only at the boundary: every identifier flag is a string Guid.
	for _, name := range []string{"stock-item-guid", "destination-stock-guid"} {
		f := stockActionFlag(t, "transfer", name)
		if !f.Required || f.Type != "string" {
			t.Errorf("transfer flag %q must be a Required string Guid, got %+v", name, f)
		}
	}
	if qty := stockActionFlag(t, "transfer", "quantity"); !qty.Required || qty.Type != "int" {
		t.Errorf("transfer quantity must be a Required int flag, got %+v", qty)
	}
	for _, name := range []string{"destination-bin-guid", "reason", "reference"} {
		if f := stockActionFlag(t, "transfer", name); f.Required {
			t.Errorf("transfer flag %q must be optional", name)
		}
	}
}

func TestStockTransfersListActionWired(t *testing.T) {
	action := findStockAction(t, "transfers")

	if action.ToolName != "UteamupStockListTransfers" {
		t.Errorf("transfers ToolName = %q, want %q", action.ToolName, "UteamupStockListTransfers")
	}
	if action.HTTPMethod != "" {
		t.Errorf("transfers HTTPMethod = %q, want \"\" (defaults to GET)", action.HTTPMethod)
	}
	if action.RESTPath != "transfers" {
		t.Errorf("transfers RESTPath = %q, want %q", action.RESTPath, "transfers")
	}

	gotFlags := make(map[string]string)
	for _, f := range action.Flags {
		gotFlags[f.Name] = f.Type
	}
	for name, ty := range map[string]string{"stock-guid": "string", "page": "int", "page-size": "int"} {
		got, ok := gotFlags[name]
		if !ok {
			t.Errorf("transfers missing expected flag %q", name)
			continue
		}
		if got != ty {
			t.Errorf("transfers flag %q type = %q, want %q", name, got, ty)
		}
	}
}

func TestStockPoReceiveActionWired(t *testing.T) {
	action := findStockAction(t, "po-receive")

	if action.ToolName != "UteamupStockReceivePurchaseOrder" {
		t.Errorf("po-receive ToolName = %q, want %q", action.ToolName, "UteamupStockReceivePurchaseOrder")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("po-receive HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "purchase-orders/{guid}/receive" {
		t.Errorf("po-receive RESTPath = %q, want %q", action.RESTPath, "purchase-orders/{guid}/receive")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "guid" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Fatalf("po-receive expected single required string positional arg 'guid', got %+v", action.Args)
	}

	file := stockActionFlag(t, "po-receive", "file")
	if !file.Required || !file.JSONFile || file.Short != "f" {
		t.Errorf("po-receive file flag must be Required JSONFile with -f short, got %+v", file)
	}
	if file.BodyName != "receivedItems" {
		t.Errorf("po-receive file BodyName = %q, want receivedItems (backend binds ReceivePurchaseOrderRequestModel.ReceivedItems)", file.BodyName)
	}
}

func TestStockBulkAdjustActionWired(t *testing.T) {
	action := findStockAction(t, "bulk-adjust")

	if action.ToolName != "UteamupStockBulkAdjust" {
		t.Errorf("bulk-adjust ToolName = %q, want %q", action.ToolName, "UteamupStockBulkAdjust")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("bulk-adjust HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "transactions/bulk" {
		t.Errorf("bulk-adjust RESTPath = %q, want %q", action.RESTPath, "transactions/bulk")
	}

	file := stockActionFlag(t, "bulk-adjust", "file")
	if !file.Required || !file.JSONFile || file.Short != "f" {
		t.Errorf("bulk-adjust file flag must be Required JSONFile with -f short, got %+v", file)
	}
	if file.BodyName != "operations" {
		t.Errorf("bulk-adjust file BodyName = %q, want operations (backend binds BulkStockTransactionsRequestModel.Operations)", file.BodyName)
	}
}

func TestStockExportImportActionsWired(t *testing.T) {
	export := findStockAction(t, "export")
	if export.ToolName != "UteamupStockExportItems" {
		t.Errorf("export ToolName = %q, want %q", export.ToolName, "UteamupStockExportItems")
	}
	if export.HTTPMethod != "" {
		t.Errorf("export HTTPMethod = %q, want \"\" (defaults to GET)", export.HTTPMethod)
	}
	if export.RESTPath != "items/export" {
		t.Errorf("export RESTPath = %q, want %q", export.RESTPath, "items/export")
	}

	imp := findStockAction(t, "import")
	if imp.ToolName != "UteamupStockImportItems" {
		t.Errorf("import ToolName = %q, want %q", imp.ToolName, "UteamupStockImportItems")
	}
	if imp.HTTPMethod != "POST" {
		t.Errorf("import HTTPMethod = %q, want POST", imp.HTTPMethod)
	}
	if imp.RESTPath != "items/import" {
		t.Errorf("import RESTPath = %q, want %q", imp.RESTPath, "items/import")
	}

	file := stockActionFlag(t, "import", "file")
	if !file.Required || !file.UploadFile || file.Short != "f" {
		t.Errorf("import file flag must be Required UploadFile with -f short, got %+v", file)
	}

	dryRun := stockActionFlag(t, "import", "dry-run")
	if dryRun.Type != "bool" || dryRun.Default != false {
		t.Errorf("import dry-run must be a bool flag defaulting to false, got %+v", dryRun)
	}
	// The backend binds [FromQuery] bool dryrun (lowercase) — BodyName carries
	// the divergent name onto the query string of the multipart POST.
	if dryRun.BodyName != "dryrun" {
		t.Errorf("import dry-run BodyName = %q, want dryrun", dryRun.BodyName)
	}
}

func TestStockBinsActionsWired(t *testing.T) {
	bins := findStockAction(t, "bins")
	if bins.ToolName != "UteamupStockListBins" {
		t.Errorf("bins ToolName = %q, want %q", bins.ToolName, "UteamupStockListBins")
	}
	if bins.RESTPath != "locations/{stockGuid}/bins" {
		t.Errorf("bins RESTPath = %q, want %q", bins.RESTPath, "locations/{stockGuid}/bins")
	}
	if sg := stockActionFlag(t, "bins", "stock-guid"); !sg.Required || sg.Type != "string" {
		t.Errorf("bins stock-guid flag must be a Required string Guid (fills the path placeholder), got %+v", sg)
	}

	create := findStockAction(t, "bins-create")
	if create.ToolName != "UteamupStockUpsertBin" {
		t.Errorf("bins-create ToolName = %q, want %q", create.ToolName, "UteamupStockUpsertBin")
	}
	if create.HTTPMethod != "POST" {
		t.Errorf("bins-create HTTPMethod = %q, want POST", create.HTTPMethod)
	}
	if create.RESTPath != "bins" {
		t.Errorf("bins-create RESTPath = %q, want %q", create.RESTPath, "bins")
	}
	for _, name := range []string{"stock-guid", "code"} {
		if f := stockActionFlag(t, "bins-create", name); !f.Required || f.Type != "string" {
			t.Errorf("bins-create flag %q must be a Required string, got %+v", name, f)
		}
	}
	if bt := stockActionFlag(t, "bins-create", "bin-type"); bt.Default != "Bin" {
		t.Errorf("bins-create bin-type Default = %v, want \"Bin\" (matches backend default)", bt.Default)
	}
	for _, name := range []string{"name", "parent-bin-guid"} {
		if f := stockActionFlag(t, "bins-create", name); f.Required {
			t.Errorf("bins-create flag %q must be optional", name)
		}
	}
}

func TestStockUnitsActionWired(t *testing.T) {
	action := findStockAction(t, "units")

	if action.ToolName != "UteamupStockListUnits" {
		t.Errorf("units ToolName = %q, want %q", action.ToolName, "UteamupStockListUnits")
	}
	if action.HTTPMethod != "" {
		t.Errorf("units HTTPMethod = %q, want \"\" (defaults to GET)", action.HTTPMethod)
	}
	if action.RESTPath != "items/{itemGuid}/units" {
		t.Errorf("units RESTPath = %q, want %q", action.RESTPath, "items/{itemGuid}/units")
	}

	// itemGuid fills the path placeholder via the required camelCased flag.
	if ig := stockActionFlag(t, "units", "item-guid"); !ig.Required || ig.Type != "string" {
		t.Errorf("units item-guid flag must be a Required string Guid, got %+v", ig)
	}

	gotFlags := make(map[string]string)
	for _, f := range action.Flags {
		gotFlags[f.Name] = f.Type
	}
	for name, ty := range map[string]string{"status": "string", "serial": "string", "page": "int", "page-size": "int"} {
		got, ok := gotFlags[name]
		if !ok {
			t.Errorf("units missing expected flag %q", name)
			continue
		}
		if got != ty {
			t.Errorf("units flag %q type = %q, want %q", name, got, ty)
		}
	}
}

func TestStockUnitsLookupActionWired(t *testing.T) {
	action := findStockAction(t, "units-lookup")

	if action.ToolName != "UteamupStockLookupUnit" {
		t.Errorf("units-lookup ToolName = %q, want %q", action.ToolName, "UteamupStockLookupUnit")
	}
	if action.HTTPMethod != "" {
		t.Errorf("units-lookup HTTPMethod = %q, want \"\" (defaults to GET)", action.HTTPMethod)
	}
	if action.RESTPath != "units/lookup/{serial}" {
		t.Errorf("units-lookup RESTPath = %q, want %q", action.RESTPath, "units/lookup/{serial}")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "serial" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Errorf("units-lookup expected single required string positional arg 'serial', got %+v", action.Args)
	}
}

func TestStockUnitTransitionActionWired(t *testing.T) {
	action := findStockAction(t, "unit-transition")

	if action.ToolName != "UteamupStockTransitionUnit" {
		t.Errorf("unit-transition ToolName = %q, want %q", action.ToolName, "UteamupStockTransitionUnit")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("unit-transition HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "units/{unitGuid}/transition" {
		t.Errorf("unit-transition RESTPath = %q, want %q", action.RESTPath, "units/{unitGuid}/transition")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "unitGuid" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Fatalf("unit-transition expected single required string positional arg 'unitGuid', got %+v", action.Args)
	}

	if ts := stockActionFlag(t, "unit-transition", "target-status"); !ts.Required || ts.Type != "string" {
		t.Errorf("unit-transition target-status must be a Required string flag, got %+v", ts)
	}
	for _, name := range []string{"asset-guid", "workorder-guid", "reason"} {
		if f := stockActionFlag(t, "unit-transition", name); f.Required || f.Type != "string" {
			t.Errorf("unit-transition flag %q must be an optional string, got %+v", name, f)
		}
	}
}

func TestStockReserveActionWired(t *testing.T) {
	action := findStockAction(t, "reserve")

	if action.ToolName != "UteamupStockCreateReservation" {
		t.Errorf("reserve ToolName = %q, want %q", action.ToolName, "UteamupStockCreateReservation")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("reserve HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "reservations" {
		t.Errorf("reserve RESTPath = %q, want %q", action.RESTPath, "reservations")
	}

	if sig := stockActionFlag(t, "reserve", "stock-item-guid"); !sig.Required || sig.Type != "string" {
		t.Errorf("reserve stock-item-guid must be a Required string Guid, got %+v", sig)
	}
	if qty := stockActionFlag(t, "reserve", "quantity"); !qty.Required || qty.Type != "int" {
		t.Errorf("reserve quantity must be a Required int flag, got %+v", qty)
	}
	for _, name := range []string{"workorder-guid", "project-guid", "unit-guid", "reserved-until"} {
		if f := stockActionFlag(t, "reserve", name); f.Required || f.Type != "string" {
			t.Errorf("reserve flag %q must be an optional string, got %+v", name, f)
		}
	}
}

func TestStockReleaseActionWired(t *testing.T) {
	action := findStockAction(t, "release")

	if action.ToolName != "UteamupStockReleaseReservation" {
		t.Errorf("release ToolName = %q, want %q", action.ToolName, "UteamupStockReleaseReservation")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("release HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "reservations/{reservationGuid}/release" {
		t.Errorf("release RESTPath = %q, want %q", action.RESTPath, "reservations/{reservationGuid}/release")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "reservationGuid" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Errorf("release expected single required string positional arg 'reservationGuid', got %+v", action.Args)
	}
}

func TestStockAtpActionWired(t *testing.T) {
	action := findStockAction(t, "atp")

	if action.ToolName != "UteamupStockGetAtp" {
		t.Errorf("atp ToolName = %q, want %q", action.ToolName, "UteamupStockGetAtp")
	}
	if action.HTTPMethod != "" {
		t.Errorf("atp HTTPMethod = %q, want \"\" (defaults to GET)", action.HTTPMethod)
	}
	if action.RESTPath != "items/{itemGuid}/atp" {
		t.Errorf("atp RESTPath = %q, want %q", action.RESTPath, "items/{itemGuid}/atp")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "itemGuid" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Errorf("atp expected single required string positional arg 'itemGuid', got %+v", action.Args)
	}
}

func TestStockDuplicateActionWired(t *testing.T) {
	action := findStockAction(t, "duplicate")

	if action.ToolName != "UteamupStockItemDuplicate" {
		t.Errorf("duplicate ToolName = %q, want %q", action.ToolName, "UteamupStockItemDuplicate")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("duplicate HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "items/{itemGuid}/duplicate" {
		t.Errorf("duplicate RESTPath = %q, want %q", action.RESTPath, "items/{itemGuid}/duplicate")
	}

	// Single Guid (string) positional arg feeding the {itemGuid} path placeholder.
	if len(action.Args) != 1 || action.Args[0].Name != "itemGuid" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Errorf("duplicate expected single required string positional arg 'itemGuid', got %+v", action.Args)
	}

	// target-stock-guid is a required Guid string → body targetStockGuid.
	target := stockActionFlag(t, "duplicate", "target-stock-guid")
	if !target.Required || target.Type != "string" {
		t.Errorf("target-stock-guid must be a Required string Guid, got %+v", target)
	}
	if target.BodyName != "targetStockGuid" {
		t.Errorf("target-stock-guid BodyName = %q, want targetStockGuid", target.BodyName)
	}

	// name is an optional string → body name.
	name := stockActionFlag(t, "duplicate", "name")
	if name.Required {
		t.Error("name flag must be optional")
	}
	if name.Type != "string" || name.BodyName != "name" {
		t.Errorf("name flag = %+v, want optional string → body name", name)
	}
}

func TestReadJSONFileFlagParsesArray(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ops.json")
	if err := os.WriteFile(path, []byte(`[{"stockItemGuid":"11111111-1111-1111-1111-111111111111","action":"Add","quantity":3}]`), 0o600); err != nil {
		t.Fatal(err)
	}

	parsed, err := readJSONFileFlag(path)
	if err != nil {
		t.Fatalf("readJSONFileFlag returned error: %v", err)
	}
	arr, ok := parsed.([]any)
	if !ok {
		t.Fatalf("parsed value type = %T, want []any", parsed)
	}
	if len(arr) != 1 {
		t.Fatalf("parsed array length = %d, want 1", len(arr))
	}
	op, ok := arr[0].(map[string]any)
	if !ok {
		t.Fatalf("element type = %T, want map[string]any", arr[0])
	}
	if op["action"] != "Add" || op["quantity"] != float64(3) {
		t.Errorf("parsed element = %+v, want action=Add quantity=3", op)
	}
}

func TestReadJSONFileFlagRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.json")
	if err := os.WriteFile(path, []byte(`{not json`), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := readJSONFileFlag(path); err == nil {
		t.Error("expected error for invalid JSON content, got nil")
	}
}

func TestReadJSONFileFlagMissingFile(t *testing.T) {
	if _, err := readJSONFileFlag(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestStockReservationsListActionWired(t *testing.T) {
	action := findStockAction(t, "reservations")

	if action.ToolName != "UteamupStockListReservations" {
		t.Errorf("reservations ToolName = %q, want %q", action.ToolName, "UteamupStockListReservations")
	}
	if action.RESTPath != "reservations" {
		t.Errorf("reservations RESTPath = %q, want %q", action.RESTPath, "reservations")
	}

	want := map[string]bool{"item-guid": false, "workorder-guid": false, "project-guid": false}
	for _, f := range action.Flags {
		if _, ok := want[f.Name]; ok {
			want[f.Name] = true
			if f.Required {
				t.Errorf("reservations flag %q should be optional", f.Name)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("reservations missing flag %q", name)
		}
	}
}

func TestStockQuarantineReleaseActionWired(t *testing.T) {
	action := findStockAction(t, "quarantine-release")

	if action.ToolName != "UteamupStockReleaseQuarantine" {
		t.Errorf("quarantine-release ToolName = %q, want %q", action.ToolName, "UteamupStockReleaseQuarantine")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("quarantine-release HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "items/{itemGuid}/quarantine/release" {
		t.Errorf("quarantine-release RESTPath = %q, want %q", action.RESTPath, "items/{itemGuid}/quarantine/release")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "itemGuid" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Fatalf("quarantine-release expected single required string positional arg 'itemGuid', got %+v", action.Args)
	}

	if q := stockActionFlag(t, "quarantine-release", "quantity"); !q.Required || q.Type != "int" {
		t.Errorf("quarantine-release quantity must be a Required int flag, got %+v", q)
	}
	if rg := stockActionFlag(t, "quarantine-release", "reason-guid"); rg.Required || rg.Type != "string" {
		t.Errorf("quarantine-release reason-guid must be an optional string flag, got %+v", rg)
	}
	if n := stockActionFlag(t, "quarantine-release", "notes"); n.Required || n.Type != "string" {
		t.Errorf("quarantine-release notes must be an optional string flag, got %+v", n)
	}
	if ug := stockActionFlag(t, "quarantine-release", "unit-guids"); ug.Required || ug.Type != "stringSlice" {
		t.Errorf("quarantine-release unit-guids must be an optional stringSlice flag, got %+v", ug)
	}
}

func TestStockQuarantineRejectActionWired(t *testing.T) {
	action := findStockAction(t, "quarantine-reject")

	if action.ToolName != "UteamupStockRejectQuarantine" {
		t.Errorf("quarantine-reject ToolName = %q, want %q", action.ToolName, "UteamupStockRejectQuarantine")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("quarantine-reject HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "items/{itemGuid}/quarantine/reject" {
		t.Errorf("quarantine-reject RESTPath = %q, want %q", action.RESTPath, "items/{itemGuid}/quarantine/reject")
	}

	if rg := stockActionFlag(t, "quarantine-reject", "reason-guid"); !rg.Required || rg.Type != "string" {
		t.Errorf("quarantine-reject reason-guid must be a Required string flag, got %+v", rg)
	}
	if q := stockActionFlag(t, "quarantine-reject", "quantity"); !q.Required || q.Type != "int" {
		t.Errorf("quarantine-reject quantity must be a Required int flag, got %+v", q)
	}
}

func TestStockAdjustmentReasonsActionWired(t *testing.T) {
	action := findStockAction(t, "reasons")

	if action.ToolName != "UteamupStockListAdjustmentReasons" {
		t.Errorf("reasons ToolName = %q, want %q", action.ToolName, "UteamupStockListAdjustmentReasons")
	}
	if action.HTTPMethod != "" {
		t.Errorf("reasons HTTPMethod = %q, want empty (GET default)", action.HTTPMethod)
	}
	if action.RESTPath != "adjustment-reasons" {
		t.Errorf("reasons RESTPath = %q, want %q", action.RESTPath, "adjustment-reasons")
	}
}

func TestStockApprovalsListActionWired(t *testing.T) {
	action := findStockAction(t, "approvals")

	if action.ToolName != "UteamupStockListApprovals" {
		t.Errorf("approvals ToolName = %q, want %q", action.ToolName, "UteamupStockListApprovals")
	}
	if action.RESTPath != "approvals" {
		t.Errorf("approvals RESTPath = %q, want %q", action.RESTPath, "approvals")
	}
	for _, name := range []string{"page", "page-size"} {
		if f := stockActionFlag(t, "approvals", name); f.Type != "int" {
			t.Errorf("approvals flag %q must be an int pagination flag, got %+v", name, f)
		}
	}
}

func TestStockApproveTransactionActionWired(t *testing.T) {
	action := findStockAction(t, "approve")

	if action.ToolName != "UteamupStockApproveTransaction" {
		t.Errorf("approve ToolName = %q, want %q", action.ToolName, "UteamupStockApproveTransaction")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("approve HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "approvals/{transactionGuid}/approve" {
		t.Errorf("approve RESTPath = %q, want %q", action.RESTPath, "approvals/{transactionGuid}/approve")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "transactionGuid" || !action.Args[0].Required {
		t.Fatalf("approve expected single required positional arg 'transactionGuid', got %+v", action.Args)
	}
}

func TestStockRejectApprovalActionWired(t *testing.T) {
	action := findStockAction(t, "reject-approval")

	if action.ToolName != "UteamupStockRejectTransaction" {
		t.Errorf("reject-approval ToolName = %q, want %q", action.ToolName, "UteamupStockRejectTransaction")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("reject-approval HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "approvals/{transactionGuid}/reject" {
		t.Errorf("reject-approval RESTPath = %q, want %q", action.RESTPath, "approvals/{transactionGuid}/reject")
	}
	// Default "" keeps the JSON body present so [FromBody] binding succeeds.
	if n := stockActionFlag(t, "reject-approval", "notes"); n.Default != "" || n.Type != "string" {
		t.Errorf("reject-approval notes must be a string flag with empty-string default, got %+v", n)
	}
}

func TestStockSettingsActionsWired(t *testing.T) {
	get := findStockAction(t, "settings")
	if get.ToolName != "UteamupStockGetSettings" || get.RESTPath != "settings" || get.HTTPMethod != "" {
		t.Errorf("settings action miswired: %+v", get)
	}

	update := findStockAction(t, "settings-update")
	if update.ToolName != "UteamupStockUpdateSettings" {
		t.Errorf("settings-update ToolName = %q, want %q", update.ToolName, "UteamupStockUpdateSettings")
	}
	if update.HTTPMethod != "PUT" || update.RESTPath != "settings" {
		t.Errorf("settings-update must be PUT settings, got %s %s", update.HTTPMethod, update.RESTPath)
	}

	if vm := stockActionFlag(t, "settings-update", "valuation-method"); vm.Default != "WeightedAverage" {
		t.Errorf("settings-update valuation-method default = %v, want WeightedAverage", vm.Default)
	}
	abc := map[string]struct {
		bodyName string
		def      int
	}{
		"abc-class-a-days": {"abcClassACountDays", 30},
		"abc-class-b-days": {"abcClassBCountDays", 90},
		"abc-class-c-days": {"abcClassCCountDays", 180},
	}
	for name, want := range abc {
		f := stockActionFlag(t, "settings-update", name)
		if f.BodyName != want.bodyName || f.Default != want.def || f.Type != "int" {
			t.Errorf("settings-update flag %q = %+v, want BodyName=%q Default=%d int", name, f, want.bodyName, want.def)
		}
	}
	if at := stockActionFlag(t, "settings-update", "approval-threshold"); at.BodyName != "approvalThresholdQuantity" || at.Default != nil {
		t.Errorf("settings-update approval-threshold must map to approvalThresholdQuantity with no default, got %+v", at)
	}
}

func TestStockGrantActionsWired(t *testing.T) {
	list := findStockAction(t, "grants")
	if list.ToolName != "UteamupStockListGrants" || list.RESTPath != "grants" || list.HTTPMethod != "" {
		t.Errorf("grants action miswired: %+v", list)
	}

	upsert := findStockAction(t, "grant-upsert")
	if upsert.ToolName != "UteamupStockUpsertGrant" || upsert.HTTPMethod != "POST" || upsert.RESTPath != "grants" {
		t.Errorf("grant-upsert miswired: %+v", upsert)
	}
	if f := stockActionFlag(t, "grant-upsert", "user-id"); !f.Required || f.Type != "string" {
		t.Errorf("grant-upsert user-id must be a Required string flag, got %+v", f)
	}
	if f := stockActionFlag(t, "grant-upsert", "stock-guid"); !f.Required || f.Type != "string" {
		t.Errorf("grant-upsert stock-guid must be a Required string flag, got %+v", f)
	}
	if f := stockActionFlag(t, "grant-upsert", "can-view"); f.Type != "bool" || f.Default != true {
		t.Errorf("grant-upsert can-view must be a bool flag defaulting to true, got %+v", f)
	}
	if f := stockActionFlag(t, "grant-upsert", "can-mutate"); f.Type != "bool" || f.Default != false {
		t.Errorf("grant-upsert can-mutate must be a bool flag defaulting to false, got %+v", f)
	}

	del := findStockAction(t, "grant-delete")
	if del.ToolName != "UteamupStockDeleteGrant" || del.HTTPMethod != "DELETE" || del.RESTPath != "grants/{grantGuid}" {
		t.Errorf("grant-delete miswired: %+v", del)
	}
	if len(del.Args) != 1 || del.Args[0].Name != "grantGuid" || !del.Args[0].Required {
		t.Fatalf("grant-delete expected single required positional arg 'grantGuid', got %+v", del.Args)
	}
}

func TestStockDueCountsActionWired(t *testing.T) {
	action := findStockAction(t, "due-counts")

	if action.ToolName != "UteamupStockDueCounts" {
		t.Errorf("due-counts ToolName = %q, want %q", action.ToolName, "UteamupStockDueCounts")
	}
	if action.RESTPath != "counts/due" || action.HTTPMethod != "" {
		t.Errorf("due-counts must be GET counts/due, got %s %s", action.HTTPMethod, action.RESTPath)
	}
	for _, name := range []string{"page", "page-size"} {
		if f := stockActionFlag(t, "due-counts", name); f.Type != "int" {
			t.Errorf("due-counts flag %q must be an int pagination flag, got %+v", name, f)
		}
	}
}

func TestStockTakeVarianceActionWired(t *testing.T) {
	action := findStockAction(t, "variance")

	if action.ToolName != "UteamupStockTakeVariance" {
		t.Errorf("variance ToolName = %q, want %q", action.ToolName, "UteamupStockTakeVariance")
	}
	if action.RESTPath != "takes/{takeGuid}/variance" || action.HTTPMethod != "" {
		t.Errorf("variance must be GET takes/{takeGuid}/variance, got %s %s", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "takeGuid" || !action.Args[0].Required {
		t.Fatalf("variance expected single required positional arg 'takeGuid', got %+v", action.Args)
	}
}

func TestStockReportActionsWired(t *testing.T) {
	aging := findStockAction(t, "report-aging")
	if aging.ToolName != "UteamupStockReportAging" || aging.RESTPath != "reports/aging" || aging.HTTPMethod != "" {
		t.Errorf("report-aging miswired: %+v", aging)
	}

	turnover := findStockAction(t, "report-turnover")
	if turnover.ToolName != "UteamupStockReportTurnover" || turnover.RESTPath != "reports/turnover" {
		t.Errorf("report-turnover miswired: %+v", turnover)
	}
	if f := stockActionFlag(t, "report-turnover", "period-days"); f.Type != "int" || f.Default != 365 {
		t.Errorf("report-turnover period-days must be an int flag defaulting to 365, got %+v", f)
	}

	dead := findStockAction(t, "report-dead-stock")
	if dead.ToolName != "UteamupStockReportDeadStock" || dead.RESTPath != "reports/dead-stock" {
		t.Errorf("report-dead-stock miswired: %+v", dead)
	}
	if f := stockActionFlag(t, "report-dead-stock", "since-days"); f.Type != "int" || f.Default != 180 {
		t.Errorf("report-dead-stock since-days must be an int flag defaulting to 180, got %+v", f)
	}
	for _, actionName := range []string{"report-aging", "report-turnover", "report-dead-stock"} {
		for _, name := range []string{"page", "page-size"} {
			if f := stockActionFlag(t, actionName, name); f.Type != "int" {
				t.Errorf("%s flag %q must be an int pagination flag, got %+v", actionName, name, f)
			}
		}
	}
}

func TestStockForecastActionsWired(t *testing.T) {
	forecast := findStockAction(t, "forecast")
	if forecast.ToolName != "UteamupStockGetForecast" || forecast.RESTPath != "items/{itemGuid}/forecast" || forecast.HTTPMethod != "" {
		t.Errorf("forecast must be GET items/{itemGuid}/forecast, got %+v", forecast)
	}
	if len(forecast.Args) != 1 || forecast.Args[0].Name != "itemGuid" || !forecast.Args[0].Required || forecast.Args[0].Type != "string" {
		t.Fatalf("forecast expected single required string positional arg 'itemGuid', got %+v", forecast.Args)
	}

	report := findStockAction(t, "forecast-report")
	if report.ToolName != "UteamupStockReportForecast" || report.RESTPath != "reports/forecast" || report.HTTPMethod != "" {
		t.Errorf("forecast-report must be GET reports/forecast, got %+v", report)
	}
	for _, name := range []string{"page", "page-size"} {
		if f := stockActionFlag(t, "forecast-report", name); f.Type != "int" {
			t.Errorf("forecast-report flag %q must be an int pagination flag, got %+v", name, f)
		}
	}

	apply := findStockAction(t, "forecast-apply")
	if apply.ToolName != "UteamupStockApplyForecast" || apply.HTTPMethod != "POST" || apply.RESTPath != "items/{itemGuid}/forecast/apply" {
		t.Errorf("forecast-apply must be POST items/{itemGuid}/forecast/apply, got %+v", apply)
	}
	if len(apply.Args) != 1 || apply.Args[0].Name != "itemGuid" || !apply.Args[0].Required {
		t.Fatalf("forecast-apply expected single required positional arg 'itemGuid', got %+v", apply.Args)
	}
	if f := stockActionFlag(t, "forecast-apply", "fields"); !f.Required || f.Type != "stringSlice" {
		t.Errorf("forecast-apply fields flag must be a required stringSlice (backend binds ApplyForecastSuggestionRequestModel.Fields), got %+v", f)
	}
}

func TestStockPoFromReceiptActionWired(t *testing.T) {
	action := findStockAction(t, "po-from-receipt")

	if action.ToolName != "UteamupStockCreatePoFromReceipt" {
		t.Errorf("po-from-receipt ToolName = %q, want %q", action.ToolName, "UteamupStockCreatePoFromReceipt")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("po-from-receipt HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "purchase-orders/from-receipt" {
		t.Errorf("po-from-receipt RESTPath = %q, want %q", action.RESTPath, "purchase-orders/from-receipt")
	}
	if len(action.Args) != 0 {
		t.Errorf("po-from-receipt should take no positional args, got %+v", action.Args)
	}

	file := stockActionFlag(t, "po-from-receipt", "file")
	if !file.Required || !file.JSONFile || file.Short != "f" {
		t.Errorf("po-from-receipt file flag must be Required JSONFile with -f short, got %+v", file)
	}
	if file.BodyName != "lines" {
		t.Errorf("po-from-receipt file BodyName = %q, want lines (backend binds CreatePurchaseOrderFromReceiptRequestModel.Lines)", file.BodyName)
	}

	for _, name := range []string{"vendor-guid", "currency-guid"} {
		if f := stockActionFlag(t, "po-from-receipt", name); f.Required || f.Type != "string" {
			t.Errorf("po-from-receipt flag %q must be an optional string, got %+v", name, f)
		}
	}
}

func TestStockCountFromPhotoActionWired(t *testing.T) {
	action := findStockAction(t, "count-from-photo")

	if action.ToolName != "UteamupStockCountFromPhoto" {
		t.Errorf("count-from-photo ToolName = %q, want %q", action.ToolName, "UteamupStockCountFromPhoto")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("count-from-photo HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "count-from-photo" {
		t.Errorf("count-from-photo RESTPath = %q, want %q (the backend route is POST api/stock/count-from-photo)", action.RESTPath, "count-from-photo")
	}
	if len(action.Args) != 0 {
		t.Errorf("count-from-photo should take no positional args, got %+v", action.Args)
	}

	// The photo is sent as a multipart IFormFile (matches the backend's [FromForm] IFormFile file).
	file := stockActionFlag(t, "count-from-photo", "file")
	if !file.Required || !file.UploadFile || file.Short != "f" {
		t.Errorf("count-from-photo file flag must be Required UploadFile with -f short, got %+v", file)
	}

	// GUIDs at the boundary: the optional stock item is a string Guid, never an int id.
	if sig := stockActionFlag(t, "count-from-photo", "stock-item-guid"); sig.Required || sig.Type != "string" {
		t.Errorf("count-from-photo stock-item-guid must be an optional string Guid, got %+v", sig)
	}
	if c := stockActionFlag(t, "count-from-photo", "context"); c.Required || c.Type != "string" {
		t.Errorf("count-from-photo context must be an optional string flag, got %+v", c)
	}
}

func TestStockOpsBatchActionWired(t *testing.T) {
	action := findStockAction(t, "ops-batch")

	if action.ToolName != "UteamupStockOpsBatch" {
		t.Errorf("ops-batch ToolName = %q, want %q", action.ToolName, "UteamupStockOpsBatch")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("ops-batch HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "ops/batch" {
		t.Errorf("ops-batch RESTPath = %q, want %q (the backend route is POST api/stock/ops/batch)", action.RESTPath, "ops/batch")
	}
	if len(action.Args) != 0 {
		t.Errorf("ops-batch should take no positional args, got %+v", action.Args)
	}

	file := stockActionFlag(t, "ops-batch", "file")
	if !file.Required || !file.JSONFile || file.Short != "f" {
		t.Errorf("ops-batch file flag must be Required JSONFile with -f short, got %+v", file)
	}
	if file.BodyName != "operations" {
		t.Errorf("ops-batch file BodyName = %q, want operations (backend binds StockOpsBatchRequestModel.Operations)", file.BodyName)
	}
}

func findDevicetokenDomain(t *testing.T) *Domain {
	t.Helper()
	for _, dom := range DefaultRegistry.Domains() {
		if dom.Name == "devicetoken" {
			return dom
		}
	}
	t.Fatal("expected devicetoken domain to be registered")
	return nil
}

func TestDevicetokenDomainRegistered(t *testing.T) {
	d := findDevicetokenDomain(t)
	if len(d.Aliases) != 1 || d.Aliases[0] != "devicetokens" {
		t.Errorf("devicetoken domain aliases = %+v, want [devicetokens]", d.Aliases)
	}
	// DeviceTokensController routes at api/devicetokens (plural) — the
	// auto-derived "/api/devicetoken" base would 404.
	if d.APIPath != "/api/devicetokens" {
		t.Errorf("devicetoken APIPath = %q, want %q", d.APIPath, "/api/devicetokens")
	}
}

func TestDevicetokenRegisterActionWired(t *testing.T) {
	d := findDevicetokenDomain(t)
	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "register" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected `register` action on devicetoken domain")
	}

	if action.ToolName != "UteamupDevicetokenRegister" {
		t.Errorf("register ToolName = %q, want %q", action.ToolName, "UteamupDevicetokenRegister")
	}
	if action.HTTPMethod != "POST" {
		t.Errorf("register HTTPMethod = %q, want POST", action.HTTPMethod)
	}
	if action.RESTPath != "register" {
		t.Errorf("register RESTPath = %q, want %q", action.RESTPath, "register")
	}

	gotFlags := make(map[string]*FlagDef)
	for i := range action.Flags {
		gotFlags[action.Flags[i].Name] = &action.Flags[i]
	}
	for _, name := range []string{"token", "platform"} {
		f, ok := gotFlags[name]
		if !ok {
			t.Errorf("register missing expected flag %q", name)
			continue
		}
		if !f.Required || f.Type != "string" {
			t.Errorf("register flag %q must be a Required string, got %+v", name, f)
		}
	}
}

func TestDevicetokenDeleteActionWired(t *testing.T) {
	d := findDevicetokenDomain(t)
	var action *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "delete" {
			action = &d.Actions[i]
			break
		}
	}
	if action == nil {
		t.Fatal("expected `delete` action on devicetoken domain")
	}

	if action.ToolName != "UteamupDevicetokenDelete" {
		t.Errorf("delete ToolName = %q, want %q", action.ToolName, "UteamupDevicetokenDelete")
	}
	// The action-name HTTPMethod map already routes `delete` as DELETE.
	if action.HTTPMethod != "" {
		t.Errorf("delete HTTPMethod = %q, want \"\" (action name maps to DELETE)", action.HTTPMethod)
	}
	if action.RESTPath != "{token}" {
		t.Errorf("delete RESTPath = %q, want %q (DELETE api/devicetokens/{token})", action.RESTPath, "{token}")
	}
	if len(action.Args) != 1 || action.Args[0].Name != "token" || !action.Args[0].Required || action.Args[0].Type != "string" {
		t.Fatalf("delete expected single required string positional arg 'token', got %+v", action.Args)
	}
}

// --- Batch 7: marketplace bridge, warranty, rental, vendor/TCO, lifecycle, unified search ---

func stockFlagByName(action *Action, name string) *FlagDef {
	for i := range action.Flags {
		if action.Flags[i].Name == name {
			return &action.Flags[i]
		}
	}
	return nil
}

func assertStockActionRoute(t *testing.T, name, tool, method, restPath string) *Action {
	t.Helper()
	action := findStockAction(t, name)
	if action.ToolName != tool {
		t.Errorf("%s ToolName = %q, want %q", name, action.ToolName, tool)
	}
	if action.HTTPMethod != method {
		t.Errorf("%s HTTPMethod = %q, want %q", name, action.HTTPMethod, method)
	}
	if action.RESTPath != restPath {
		t.Errorf("%s RESTPath = %q, want %q", name, action.RESTPath, restPath)
	}
	return action
}

func TestStockListOnMarketplaceActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "list-on-marketplace", "UteamupStockListOnMarketplace", "POST", "items/{itemGuid}/marketplace/list")
	if len(action.Args) != 1 || action.Args[0].Name != "itemGuid" || !action.Args[0].Required {
		t.Fatalf("list-on-marketplace expected required itemGuid arg, got %+v", action.Args)
	}
	price := stockFlagByName(action, "price")
	if price == nil || !price.Required || price.Type != "float" || price.BodyName != "price" {
		t.Errorf("price flag must be a required float bound to body `price`, got %+v", price)
	}
	if lt := stockFlagByName(action, "listing-type"); lt == nil || lt.BodyName != "listingType" || lt.Default != "Sale" {
		t.Errorf("listing-type flag must default Sale and bind body `listingType`, got %+v", lt)
	}
	if ug := stockFlagByName(action, "unit-guids"); ug == nil || ug.Type != "stringSlice" || ug.BodyName != "unitGuids" {
		t.Errorf("unit-guids flag must be a stringSlice bound to body `unitGuids`, got %+v", ug)
	}
}

func TestStockDelistActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "delist", "UteamupStockDelistFromMarketplace", "POST", "items/{itemGuid}/marketplace/delist/{marketplaceItemGuid}")
	if len(action.Args) != 2 || action.Args[0].Name != "itemGuid" || action.Args[1].Name != "marketplaceItemGuid" {
		t.Fatalf("delist expected itemGuid + marketplaceItemGuid args, got %+v", action.Args)
	}
}

func TestStockReceiveMarketplacePurchaseActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "receive-marketplace-purchase", "UteamupStockReceiveMarketplacePurchase", "POST", "purchase-orders/marketplace-receive")
	if tx := stockFlagByName(action, "transaction-guid"); tx == nil || !tx.Required || tx.BodyName != "marketplaceTransactionGuid" {
		t.Errorf("transaction-guid must be required and bind body `marketplaceTransactionGuid`, got %+v", tx)
	}
	if sg := stockFlagByName(action, "stock-guid"); sg == nil || !sg.Required || sg.BodyName != "stockGuid" {
		t.Errorf("stock-guid must be required and bind body `stockGuid`, got %+v", sg)
	}
	if cn := stockFlagByName(action, "create-new-item"); cn == nil || cn.Type != "bool" || cn.BodyName != "createNewStockItem" {
		t.Errorf("create-new-item must be a bool bound to body `createNewStockItem`, got %+v", cn)
	}
}

func TestStockWarrantyClaimsActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "warranty-claims", "UteamupStockListWarrantyClaims", "", "warranty-claims")
	if stockFlagByName(action, "status") == nil {
		t.Error("warranty-claims expected a status filter flag")
	}
	if stockFlagByName(action, "page") == nil || stockFlagByName(action, "page-size") == nil {
		t.Error("warranty-claims expected pagination flags")
	}
}

func TestStockWarrantyClaimCreateActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "warranty-claim-create", "UteamupStockCreateWarrantyClaim", "POST", "warranty-claims")
	if uf := stockFlagByName(action, "unit-guid"); uf == nil || !uf.Required || uf.BodyName != "stockItemUnitGuid" {
		t.Errorf("unit-guid must be required and bind body `stockItemUnitGuid`, got %+v", uf)
	}
}

func TestStockWarrantyClaimTransitionActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "warranty-claim-transition", "UteamupStockTransitionWarrantyClaim", "POST", "warranty-claims/{guid}/transition")
	if len(action.Args) != 1 || action.Args[0].Name != "guid" {
		t.Fatalf("warranty-claim-transition expected guid arg, got %+v", action.Args)
	}
	if s := stockFlagByName(action, "status"); s == nil || !s.Required || s.BodyName != "status" {
		t.Errorf("status must be required and bind body `status`, got %+v", s)
	}
}

func TestStockWarrantyCoverageDomainWired(t *testing.T) {
	var dom *Domain
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "warranty-coverage" {
			dom = d
			break
		}
	}
	if dom == nil {
		t.Fatal("expected warranty-coverage domain to be registered")
	}
	if dom.APIPath != "/api/asset" {
		t.Errorf("warranty-coverage APIPath = %q, want /api/asset (route is ~/api/asset/{guid}/warranty-coverage)", dom.APIPath)
	}
	if len(dom.Aliases) != 1 || dom.Aliases[0] != "warranty" {
		t.Errorf("warranty-coverage aliases = %+v, want [warranty]", dom.Aliases)
	}
	if len(dom.Actions) != 1 {
		t.Fatalf("warranty-coverage expected exactly 1 action, got %d", len(dom.Actions))
	}
	a := dom.Actions[0]
	if a.ToolName != "UteamupStockAssetWarrantyCoverage" || a.RESTPath != "{guid}/warranty-coverage" {
		t.Errorf("warranty-coverage action wired wrong: tool=%q path=%q", a.ToolName, a.RESTPath)
	}
}

func TestStockRentalCheckoutActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "rental-checkout", "UteamupStockRentalCheckout", "POST", "units/{guid}/rental/checkout")
	if len(action.Args) != 1 || action.Args[0].Name != "guid" {
		t.Fatalf("rental-checkout expected guid arg, got %+v", action.Args)
	}
	if d := stockFlagByName(action, "due-at"); d == nil || !d.Required || d.BodyName != "dueAt" {
		t.Errorf("due-at must be required and bind body `dueAt`, got %+v", d)
	}
	if dr := stockFlagByName(action, "daily-rate"); dr == nil || dr.Type != "float" || dr.BodyName != "dailyRate" {
		t.Errorf("daily-rate must be a float bound to body `dailyRate`, got %+v", dr)
	}
}

func TestStockRentalReturnActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "rental-return", "UteamupStockRentalReturn", "POST", "rentals/{guid}/return")
	if len(action.Args) != 1 || action.Args[0].Name != "guid" {
		t.Fatalf("rental-return expected guid arg, got %+v", action.Args)
	}
}

func TestStockRentalsActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "rentals", "UteamupStockListRentals", "", "rentals")
	if stockFlagByName(action, "status") == nil || stockFlagByName(action, "page") == nil {
		t.Error("rentals expected status filter + pagination flags")
	}
}

func TestStockVendorScoreActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "vendor-score", "UteamupStockVendorScore", "", "vendors/{guid}/score")
	if len(action.Args) != 1 || action.Args[0].Name != "guid" {
		t.Fatalf("vendor-score expected guid arg, got %+v", action.Args)
	}
}

func TestStockVendorRankingActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "vendor-ranking", "UteamupStockVendorRanking", "", "reports/vendor-ranking")
	if stockFlagByName(action, "page") == nil {
		t.Error("vendor-ranking expected pagination flags")
	}
}

func TestStockTcoActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "tco", "UteamupStockItemTco", "", "items/{guid}/tco")
	if len(action.Args) != 1 || action.Args[0].Name != "guid" {
		t.Fatalf("tco expected guid arg, got %+v", action.Args)
	}
}

func TestStockPartEffectivenessActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "part-effectiveness", "UteamupStockPartEffectiveness", "", "reports/part-effectiveness")
	if n := stockFlagByName(action, "name"); n == nil || !n.Required {
		t.Errorf("part-effectiveness expected a required name flag, got %+v", n)
	}
	if stockFlagByName(action, "page") == nil {
		t.Error("part-effectiveness expected pagination flags")
	}
}

func TestStockLifecycleRulesActionWired(t *testing.T) {
	assertStockActionRoute(t, "lifecycle-rules", "UteamupStockListLifecycleRules", "", "lifecycle-rules")
}

func TestStockLifecycleRuleCreateActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "lifecycle-rule-create", "UteamupStockCreateLifecycleRule", "POST", "lifecycle-rules")
	for _, name := range []string{"name", "trigger", "action-type"} {
		f := stockFlagByName(action, name)
		if f == nil || !f.Required {
			t.Errorf("lifecycle-rule-create flag %q must be required, got %+v", name, f)
		}
	}
	if at := stockFlagByName(action, "action-type"); at == nil || at.BodyName != "action" {
		t.Errorf("action-type must bind body field `action`, got %+v", at)
	}
}

func TestStockLifecycleRuleUpdateActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "lifecycle-rule-update", "UteamupStockUpdateLifecycleRule", "PUT", "lifecycle-rules/{guid}")
	if len(action.Args) != 1 || action.Args[0].Name != "guid" {
		t.Fatalf("lifecycle-rule-update expected guid arg, got %+v", action.Args)
	}
}

func TestStockLifecycleRuleDeleteActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "lifecycle-rule-delete", "UteamupStockDeleteLifecycleRule", "DELETE", "lifecycle-rules/{guid}")
	if len(action.Args) != 1 || action.Args[0].Name != "guid" {
		t.Fatalf("lifecycle-rule-delete expected guid arg, got %+v", action.Args)
	}
}

func TestStockUnifiedSearchActionWired(t *testing.T) {
	action := assertStockActionRoute(t, "unified-search", "UteamupStockUnifiedSearch", "", "search/unified")
	if q := stockFlagByName(action, "q"); q == nil || !q.Required {
		t.Errorf("unified-search expected a required q flag, got %+v", q)
	}
}
