package registry

import "testing"

func TestAIProviderGovernanceSnapshotMirrorsReadOnlyMcpTool(t *testing.T) {
	domain := findRegisteredDomain(t, "ai-provider")
	var snapshot *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "governance-snapshot" {
			snapshot = &domain.Actions[index]
			break
		}
	}
	if snapshot == nil {
		t.Fatal("governance-snapshot action is not registered")
	}

	if snapshot.ToolName != "UteamupGetTenantAIControlPlaneSnapshot" {
		t.Fatalf("unexpected MCP tool: %q", snapshot.ToolName)
	}
	if snapshot.HTTPMethod != "GET" ||
		snapshot.RESTBasePath != "/api/tenant-ai-governance" ||
		snapshot.RESTPath != "snapshot" {
		t.Fatalf("unexpected REST contract: %#v", snapshot)
	}
	if len(snapshot.Args) != 0 || len(snapshot.Flags) != 0 {
		t.Fatalf("snapshot must be a no-argument read: %#v", snapshot)
	}
}

func TestAIProviderGovernanceSnapshotResolvesWithoutIdentifiers(t *testing.T) {
	domain := findRegisteredDomain(t, "ai-provider")
	var action Action
	for _, candidate := range domain.Actions {
		if candidate.Name == "governance-snapshot" {
			action = candidate
			break
		}
	}

	path, consumed := buildRESTPath(domain, action, map[string]any{})
	if path != "/api/tenant-ai-governance/snapshot" {
		t.Fatalf("path = %q", path)
	}
	if len(consumed) != 0 {
		t.Fatalf("snapshot unexpectedly consumed identifiers: %v", consumed)
	}
}
