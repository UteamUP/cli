package registry

import "testing"

func TestFleetDashboardExposesGuidFirstReadParity(t *testing.T) {
	domain := findDomain("fleet-dashboard")
	if domain == nil {
		t.Fatal("expected fleet-dashboard domain")
	}

	actions := make(map[string]Action, len(domain.Actions))
	for _, action := range domain.Actions {
		actions[action.Name] = action
	}
	if actions["utilization"].ToolName != "UteamupFleetDashboardGetUtilization" {
		t.Fatalf("utilization action = %+v", actions["utilization"])
	}
	if actions["compliance"].ToolName != "UteamupFleetDashboardGetCompliance" {
		t.Fatalf("compliance action = %+v", actions["compliance"])
	}
}

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

func TestFuelTransactionDomainUsesGuidArguments(t *testing.T) {
	domain, ok := Get("fuel-transaction")
	if !ok {
		t.Fatal("fuel-transaction domain not registered")
	}

	actions := make(map[string]Action, len(domain.Actions))
	for _, action := range domain.Actions {
		actions[action.Name] = action
	}

	for _, name := range []string{"get", "update", "delete"} {
		action := actions[name]
		if len(action.Args) != 1 || action.Args[0].Name != "transactionGuid" || action.Args[0].Type != "string" {
			t.Fatalf("%s must require one string transactionGuid argument, got %+v", name, action.Args)
		}
	}

	for _, name := range []string{"summary", "efficiency"} {
		action := actions[name]
		if len(action.Args) != 1 || action.Args[0].Name != "assetGuid" || action.Args[0].Type != "string" {
			t.Fatalf("%s must require one string assetGuid argument, got %+v", name, action.Args)
		}
	}

	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Name == "id" || arg.Type == "int" {
				t.Fatalf("fuel action %s leaks an integer identifier argument: %+v", action.Name, arg)
			}
		}
	}
}

func TestVehicleInspectionDomainUsesGuidArguments(t *testing.T) {
	domain, ok := Get("vehicle-inspection")
	if !ok {
		t.Fatal("vehicle-inspection domain not registered")
	}

	actions := make(map[string]Action, len(domain.Actions))
	for _, action := range domain.Actions {
		actions[action.Name] = action
	}

	for _, name := range []string{"get", "update", "delete", "submit-items", "complete"} {
		action := actions[name]
		if len(action.Args) != 1 || action.Args[0].Name != "inspectionGuid" || action.Args[0].Type != "string" {
			t.Fatalf("%s must require one string inspectionGuid argument, got %+v", name, action.Args)
		}
	}
	if actions["submit-items"].RESTPath != "by-guid/{inspectionGuid}/items" || actions["submit-items"].HTTPMethod != "POST" {
		t.Fatalf("submit-items route mismatch: %+v", actions["submit-items"])
	}
	if actions["complete"].RESTPath != "by-guid/{inspectionGuid}/complete" || actions["complete"].HTTPMethod != "POST" {
		t.Fatalf("complete route mismatch: %+v", actions["complete"])
	}
	var correctiveFlag *FlagDef
	for index := range actions["complete"].Flags {
		if actions["complete"].Flags[index].Name == "create-corrective-workorder" {
			correctiveFlag = &actions["complete"].Flags[index]
			break
		}
	}
	if correctiveFlag == nil || correctiveFlag.BodyName != "createCorrectiveWorkorder" || correctiveFlag.Type != "bool" {
		t.Fatalf("complete must expose explicit corrective-workorder confirmation: %+v", actions["complete"].Flags)
	}
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Name == "id" || arg.Type == "int" {
				t.Fatalf("vehicle inspection action %s leaks integer identifiers: %+v", action.Name, arg)
			}
		}
	}
}

func TestDriverAssignmentDomainUsesGuidArguments(t *testing.T) {
	domain, ok := Get("driver-assignment")
	if !ok {
		t.Fatal("driver-assignment domain not registered")
	}
	actions := make(map[string]Action, len(domain.Actions))
	for _, action := range domain.Actions {
		actions[action.Name] = action
	}
	for _, name := range []string{"get", "update", "delete", "end"} {
		action := actions[name]
		if len(action.Args) != 1 || action.Args[0].Name != "assignmentGuid" || action.Args[0].Type != "string" {
			t.Fatalf("%s must require one string assignmentGuid argument, got %+v", name, action.Args)
		}
	}
	current := actions["current"]
	if len(current.Args) != 1 || current.Args[0].Name != "assetGuid" || current.Args[0].Type != "string" {
		t.Fatalf("current must require one string assetGuid argument, got %+v", current.Args)
	}
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Name == "id" || arg.Type == "int" {
				t.Fatalf("driver assignment action %s leaks integer identifiers: %+v", action.Name, arg)
			}
		}
	}
}

func TestDriverDomainUsesGuidArguments(t *testing.T) {
	domain, ok := Get("driver")
	if !ok {
		t.Fatal("driver domain not registered")
	}

	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if arg.Name == "id" || arg.Name == "driverId" || arg.Name == "licenseId" || arg.Type == "int" {
				t.Fatalf("driver action %s leaks an integer identifier: %+v", action.Name, arg)
			}
		}
	}
}
