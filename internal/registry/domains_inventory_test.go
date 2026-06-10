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
