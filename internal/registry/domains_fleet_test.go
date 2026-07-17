package registry

import "testing"

func TestFleetMaintenanceProposalUsesGovernedIdempotentRoute(t *testing.T) {
	domain := findDomain("fleet-dashboard")
	if domain == nil {
		t.Fatal("expected fleet-dashboard domain")
	}

	var proposal *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "propose-maintenance" {
			proposal = &domain.Actions[index]
			break
		}
	}
	if proposal == nil {
		t.Fatal("expected propose-maintenance action")
	}

	if proposal.ToolName != "UteamupFleetMaintenancePropose" {
		t.Fatalf("tool = %q", proposal.ToolName)
	}
	if proposal.HTTPMethod != "POST" {
		t.Fatalf("method = %q, want POST", proposal.HTTPMethod)
	}
	path, consumed := buildRESTPath(domain, *proposal, map[string]any{})
	if path != "/api/upmateassistant/fleet/maintenance-proposals" {
		t.Fatalf("path = %q", path)
	}
	if len(consumed) != 0 {
		t.Fatalf("unexpected consumed args: %v", consumed)
	}

	flags := map[string]FlagDef{}
	for _, flag := range proposal.Flags {
		flags[flag.Name] = flag
	}
	if !flags["source-type"].Required || flags["source-type"].BodyName != "sourceType" {
		t.Fatalf("source-type flag = %+v", flags["source-type"])
	}
	if !flags["source-guid"].Required || flags["source-guid"].BodyName != "sourceGuid" {
		t.Fatalf("source-guid flag = %+v", flags["source-guid"])
	}
	if !flags["idempotency-key"].Required ||
		flags["idempotency-key"].HeaderName != "Idempotency-Key" {
		t.Fatalf("idempotency flag = %+v", flags["idempotency-key"])
	}
}

func TestFleetContactDomainMirrorsGuidFirstMcpTools(t *testing.T) {
	domain := findDomain("fleet-contact")
	if domain == nil {
		t.Fatal("expected fleet-contact domain")
	}

	want := map[string]struct {
		tool string
		args []string
	}{
		"list":   {tool: "UteamupFleetAssetContactList", args: []string{"assetExternalGuid"}},
		"add":    {tool: "UteamupFleetAssetContactAdd", args: []string{"assetExternalGuid"}},
		"delete": {tool: "UteamupFleetAssetContactDelete", args: []string{"assetExternalGuid", "associationExternalGuid"}},
	}
	if len(domain.Actions) != len(want) {
		t.Fatalf("fleet-contact actions = %d, want %d", len(domain.Actions), len(want))
	}

	for _, action := range domain.Actions {
		expected, ok := want[action.Name]
		if !ok {
			t.Fatalf("unexpected action %q", action.Name)
		}
		if action.ToolName != expected.tool {
			t.Fatalf("%s tool = %q, want %q", action.Name, action.ToolName, expected.tool)
		}
		if len(action.Args) != len(expected.args) {
			t.Fatalf("%s args = %d, want %d", action.Name, len(action.Args), len(expected.args))
		}
		for index, arg := range action.Args {
			if arg.Name != expected.args[index] || arg.Type != "string" || !arg.Required {
				t.Fatalf("%s arg[%d] = %+v", action.Name, index, arg)
			}
			if arg.Name == "id" || arg.Name == "assetId" || arg.Name == "contactId" {
				t.Fatalf("%s exposes integer-style argument %q", action.Name, arg.Name)
			}
		}
	}
}

func TestVehicleInspectionOverdueActionMirrorsAssistantSafeMCPRead(t *testing.T) {
	domain := findDomain("vehicle-inspection")
	if domain == nil {
		t.Fatal("vehicle-inspection domain is not registered")
	}

	var action *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "overdue" {
			action = &domain.Actions[index]
			break
		}
	}
	if action == nil {
		t.Fatal("vehicle-inspection overdue action is not registered")
	}
	if action.ToolName != "UteamupVehicleInspectionGetOverdue" ||
		action.HTTPMethod != "GET" || action.RESTPath != "overdue" {
		t.Fatalf(
			"vehicle-inspection overdue action = tool %q, method %q, path %q",
			action.ToolName,
			action.HTTPMethod,
			action.RESTPath,
		)
	}
	if len(action.Args) != 0 || len(action.Flags) != 0 {
		t.Fatalf("vehicle-inspection overdue action unexpectedly accepts identifiers: %+v", action)
	}
}
