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
