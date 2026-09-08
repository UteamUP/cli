package registry

import "testing"

func TestTeamsIntegrationDomainIsReadOnly(t *testing.T) {
	domain := findDomain("teamsintegration")
	if domain == nil {
		t.Fatal("expected teamsintegration domain")
	}
	if domain.APIPath != "/api/teamsintegration" {
		t.Fatalf("API path = %q", domain.APIPath)
	}
	if len(domain.Actions) != 4 {
		t.Fatalf("actions = %d, want 4", len(domain.Actions))
	}

	// Linking and mirroring both need a per-person delegated Microsoft token obtained through a
	// browser. The CLI cannot acquire one, so an action offering them would always fail - and a
	// command that always fails is worse than a missing command.
	for _, action := range domain.Actions {
		switch action.Name {
		case "link", "unlink", "bind", "unbind", "mirror":
			t.Fatalf("action %q requires a delegated browser sign-in the CLI cannot perform", action.Name)
		}
	}

	byName := map[string]Action{}
	for _, action := range domain.Actions {
		byName[action.Name] = action
	}

	if got := byName["config"]; got.ToolName != "UteamupTeamsConfigGet" || got.HTTPMethod != "GET" {
		t.Fatalf("config action = %+v", got)
	}
	if got := byName["test"]; got.ToolName != "UteamupTeamsConfigTest" || got.HTTPMethod != "POST" {
		t.Fatalf("test action = %+v", got)
	}
	if got := byName["me"]; got.ToolName != "UteamupTeamsMyIdentityGet" || got.HTTPMethod != "GET" {
		t.Fatalf("me action = %+v", got)
	}

	// The caller's identity comes from the bearer token. A user or tenant argument here would
	// invite reading somebody else's link state by guessing an id.
	for _, action := range domain.Actions {
		for _, flag := range action.Flags {
			switch flag.Name {
			case "user-id", "userId", "user-guid", "tenant-id", "tenant-guid", "email":
				t.Fatalf("forbidden identity flag %q on %s", flag.Name, action.Name)
			}
		}
	}

	binding := byName["binding"]
	if len(binding.Args) != 1 || binding.Args[0].Name != "guid" {
		t.Fatalf("binding args = %+v", binding.Args)
	}
	path, consumed := buildRESTPath(domain, binding, map[string]any{
		"guid": "6f1f2b9c-0000-4000-8000-000000000001",
	})
	if path != "/api/teamsintegration/conversations/6f1f2b9c-0000-4000-8000-000000000001/binding" {
		t.Fatalf("path = %q", path)
	}
	// The guid must be consumed by the path template, not left over to be appended as a query
	// parameter as well.
	if len(consumed) != 1 || consumed[0] != "guid" {
		t.Fatalf("consumed = %v, want [guid]", consumed)
	}
}
