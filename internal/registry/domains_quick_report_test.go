package registry

import "testing"

func TestQuickReportActionsUseGuidOnlyMCPContracts(t *testing.T) {
	domain := findDomain("quick-report")
	if domain == nil {
		t.Fatal("quick-report domain is not registered")
	}

	create := findAction(domain, "create")
	if create == nil || !create.MCPOnly || create.ToolName != "UteamupQuickreportCreate" {
		t.Fatalf("quick-report create action = %+v", create)
	}
	createFlags := flagsToMap(create.Flags)
	assertQuickReportIdempotencyFlag(t, createFlags)
	if flag, ok := createFlags["asset-guid"]; !ok || flag.Type != "uuid" {
		t.Fatalf("asset-guid flag = %+v", flag)
	}
	if _, exists := createFlags["asset-id"]; exists {
		t.Fatal("quick-report create exposes legacy asset-id")
	}

	complete := findAction(domain, "complete")
	if complete == nil || !complete.MCPOnly || complete.ToolName != "UteamupQuickreportComplete" {
		t.Fatalf("quick-report complete action = %+v", complete)
	}
	completeFlags := flagsToMap(complete.Flags)
	assertQuickReportIdempotencyFlag(t, completeFlags)
	if flag, ok := completeFlags["workorder-guid"]; !ok || flag.Type != "uuid" || !flag.Required {
		t.Fatalf("workorder-guid flag = %+v", flag)
	}
	if _, exists := completeFlags["workorder-id"]; exists {
		t.Fatal("quick-report complete exposes legacy workorder-id")
	}
}

func assertQuickReportIdempotencyFlag(t *testing.T, flags map[string]FlagDef) {
	t.Helper()

	flag, ok := flags["idempotency-key"]
	if !ok || flag.Type != "uuid" || !flag.Required || flag.BodyName != "idempotencyKey" {
		t.Fatalf("idempotency-key flag = %+v", flag)
	}
	if flag.HeaderName != "" {
		t.Fatalf("MCP-only idempotency key must be sent in the tool body, got header %q", flag.HeaderName)
	}
}
