package registry

import "testing"

// The asset-dossier domain mirrors AssetDossierController (/api/assetdossier/*).
// Generation is asynchronous and spends AI credits, so two things are load-bearing:
// requestGuid is the caller's idempotency key, and --wait must poll the job the request
// returned rather than re-posting it.

func assetDossierDomain(t *testing.T) *Domain {
	t.Helper()
	d := findDomain("asset-dossier")
	if d == nil {
		t.Fatal("expected asset-dossier domain to be registered")
	}
	return d
}

func TestAssetDossierDomainRoutesUnderAssetdossier(t *testing.T) {
	d := assetDossierDomain(t)
	if d.APIPath != "/api/assetdossier" {
		t.Errorf("APIPath = %q, want /api/assetdossier", d.APIPath)
	}
}

func TestAssetDossierActionsMatchBackendContract(t *testing.T) {
	d := assetDossierDomain(t)
	cases := []struct{ name, tool, method, path string }{
		{"request", "UteamupAssetDossierRequest", "POST", "jobs"},
		{"get", "UteamupAssetDossierGet", "", "jobs/{jobGuid}"},
		{"list", "UteamupAssetDossierListByAsset", "", "by-asset/{assetGuid}"},
		{"cancel", "UteamupAssetDossierCancel", "POST", "jobs/{jobGuid}/cancel"},
	}

	for _, c := range cases {
		a := findAction(d, c.name)
		if a == nil {
			t.Errorf("expected %s action on asset-dossier", c.name)
			continue
		}
		if a.ToolName != c.tool {
			t.Errorf("%s: tool = %q, want %q", c.name, a.ToolName, c.tool)
		}
		if a.HTTPMethod != c.method || a.RESTPath != c.path {
			t.Errorf("%s: route = %s %q, want %s %q", c.name, a.HTTPMethod, a.RESTPath, c.method, c.path)
		}
	}
	if len(d.Actions) != len(cases) {
		t.Errorf("asset-dossier has %d actions, want %d", len(d.Actions), len(cases))
	}
}

func TestAssetDossierRoutesExpandFromTheirArgs(t *testing.T) {
	d := assetDossierDomain(t)
	cases := []struct {
		action string
		args   map[string]any
		want   string
	}{
		{"get", map[string]any{"jobGuid": "j-1"}, "/api/assetdossier/jobs/j-1"},
		{"list", map[string]any{"assetGuid": "a-1"}, "/api/assetdossier/by-asset/a-1"},
		{"cancel", map[string]any{"jobGuid": "j-1"}, "/api/assetdossier/jobs/j-1/cancel"},
	}

	for _, c := range cases {
		action := findAction(d, c.action)
		if action == nil {
			t.Fatalf("missing asset-dossier action %q", c.action)
		}
		got, consumed := buildRESTPath(d, *action, c.args)
		if got != c.want {
			t.Errorf("%s path = %q, want %q", c.action, got, c.want)
		}
		if len(consumed) != 1 {
			t.Errorf("%s consumed = %v, want one path arg", c.action, consumed)
		}
	}
}

// Re-posting the same requestGuid returns the existing job instead of charging again.
// A missing or optional flag here turns a retried command into a second charge.
func TestAssetDossierRequestRequiresItsIdempotencyKey(t *testing.T) {
	d := assetDossierDomain(t)
	request := findAction(d, "request")
	if request == nil {
		t.Fatal("expected request action on asset-dossier")
	}

	flags := flagsToMap(request.Flags)
	asset, ok := flags["asset-guid"]
	if !ok || !asset.Required || asset.BodyName != "assetGuid" || asset.Type != "non-empty-uuid" {
		t.Errorf("--asset-guid must be a required non-empty GUID mapped to assetGuid, got %+v", asset)
	}
	key, ok := flags["request-guid"]
	if !ok || !key.Required || key.BodyName != "requestGuid" || key.Type != "non-empty-uuid" {
		t.Errorf("--request-guid must be a required non-empty GUID mapped to requestGuid, got %+v", key)
	}
	options, ok := flags["options"]
	if !ok || !options.JSONFile {
		t.Errorf("--options must be a JSON file flag, got %+v", options)
	}
}

// --wait blocks on the job the POST returned. It has to be LocalOnly (never a request
// field), poll the GET route, and stop on every terminal status — a missing `cancelled`
// or `failed` would hang the terminal until the timeout.
func TestAssetDossierRequestWaitPollsTheJobItQueued(t *testing.T) {
	d := assetDossierDomain(t)
	request := findAction(d, "request")
	if request == nil {
		t.Fatal("expected request action on asset-dossier")
	}

	wait, ok := flagsToMap(request.Flags)["wait"]
	if !ok {
		t.Fatal("expected a --wait flag on asset-dossier request")
	}
	if wait.Type != "bool" || !wait.LocalOnly {
		t.Errorf("--wait must be a local-only boolean, got %+v", wait)
	}
	if request.PollWaitFlag != "wait" {
		t.Errorf("PollWaitFlag = %q, want wait", request.PollWaitFlag)
	}
	if request.PollPath != "jobs/{jobGuid}" {
		t.Errorf("PollPath = %q, want jobs/{jobGuid}", request.PollPath)
	}
	if request.PollIdentifierField != "jobGuid" || request.PollStatusField != "status" {
		t.Errorf("poll fields = %q/%q, want jobGuid/status", request.PollIdentifierField, request.PollStatusField)
	}

	terminal := map[string]bool{}
	for _, status := range request.PollTerminalStatuses {
		terminal[status] = true
	}
	for _, status := range []string{"completed", "failed", "cancelled"} {
		if !terminal[status] {
			t.Errorf("PollTerminalStatuses is missing %q: %v", status, request.PollTerminalStatuses)
		}
	}
	if len(request.PollTerminalStatuses) != 3 {
		t.Errorf("PollTerminalStatuses = %v, want exactly the three terminal statuses", request.PollTerminalStatuses)
	}

	if err := validateActionDefinition(*request); err != nil {
		t.Errorf("request definition is invalid: %v", err)
	}
}

// isTerminalStatus is what stops the poll loop. It compares case-insensitively because
// the backend serialises the status lowercase while the registry spells it out.
func TestIsTerminalStatusMatchesRegardlessOfCase(t *testing.T) {
	terminal := []string{"completed", "failed", "cancelled"}

	for _, status := range []any{"completed", "Completed", "CANCELLED"} {
		if !isTerminalStatus(status, terminal) {
			t.Errorf("expected %v to be terminal", status)
		}
	}
	for _, status := range []any{"queued", "processing", nil, 3} {
		if isTerminalStatus(status, terminal) {
			t.Errorf("expected %v not to be terminal", status)
		}
	}
}
