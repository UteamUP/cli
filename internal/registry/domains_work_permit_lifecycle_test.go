package registry

import "testing"

func TestWorkPermitLifecycleDomainRoutesThroughMcpTools(t *testing.T) {
	t.Parallel()
	domain := findDomain("work-permit-lifecycle")
	if domain == nil {
		t.Fatal("work-permit-lifecycle domain is not registered")
	}
	want := map[string]string{
		"list":      "UteamupWorkPermitList",
		"get":       "UteamupWorkPermitGet",
		"templates": "UteamupWorkPermitTemplateList",
		"template":  "UteamupWorkPermitTemplateGet",
		"create":    "UteamupWorkPermitCreate",
		"update":    "UteamupWorkPermitUpdate",
		"submit":    "UteamupWorkPermitSubmit",
	}
	if len(domain.Actions) != len(want) {
		t.Fatalf("actions = %d, want %d", len(domain.Actions), len(want))
	}
	for _, action := range domain.Actions {
		if want[action.Name] != action.ToolName || !action.MCPOnly {
			t.Fatalf("action contract = %+v", action)
		}
	}
}

func TestWorkPermitLifecycleExposesNoHumanDecisionActions(t *testing.T) {
	t.Parallel()
	domain := findDomain("work-permit-lifecycle")
	if domain == nil {
		t.Fatal("work-permit-lifecycle domain is not registered")
	}
	for _, action := range domain.Actions {
		switch action.Name {
		case "approve", "reject", "sign", "complete", "revoke", "delete":
			t.Fatalf("%s must stay a human decision in the app", action.Name)
		}
	}
}

func TestWorkPermitLifecycleArgumentsAreGuidsWithToolNames(t *testing.T) {
	t.Parallel()
	domain := findDomain("work-permit-lifecycle")
	if domain == nil {
		t.Fatal("work-permit-lifecycle domain is not registered")
	}
	wantArg := map[string]string{"get": "workPermitGuid", "update": "workPermitGuid", "submit": "workPermitGuid", "template": "templateGuid"}
	for _, action := range domain.Actions {
		body, ok := wantArg[action.Name]
		if !ok {
			if len(action.Args) != 0 {
				t.Fatalf("%s takes no positional arguments: %+v", action.Name, action.Args)
			}
			continue
		}
		if len(action.Args) != 1 || action.Args[0].BodyName != body || action.Args[0].Type != "non-empty-uuid" || !action.Args[0].Required {
			t.Fatalf("%s must require a GUID sent as %s: %+v", action.Name, body, action.Args)
		}
	}
}

func TestWorkPermitLifecycleCreateRequiresTitleAndTakesNoTenantOrUser(t *testing.T) {
	t.Parallel()
	domain := findDomain("work-permit-lifecycle")
	if domain == nil {
		t.Fatal("work-permit-lifecycle domain is not registered")
	}
	for _, action := range domain.Actions {
		for _, flag := range action.Flags {
			switch flag.Name {
			case "tenant", "tenant-guid", "tenant-id", "user", "user-id", "user-guid":
				t.Fatalf("%s must not take --%s; identity comes from the session", action.Name, flag.Name)
			}
			if action.Name == "create" && flag.Name == "title" && !flag.Required {
				t.Fatal("create must require --title")
			}
			if action.Name == "update" && flag.Name == "title" && flag.Required {
				t.Fatal("update must leave --title optional")
			}
		}
	}
}

func TestWorkPermitLifecycleListStatusIsConstrained(t *testing.T) {
	t.Parallel()
	domain := findDomain("work-permit-lifecycle")
	if domain == nil {
		t.Fatal("work-permit-lifecycle domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name != "list" {
			continue
		}
		for _, flag := range action.Flags {
			if flag.Name == "status" && len(flag.AllowedValues) != 8 {
				t.Fatalf("status must allow the eight permit statuses: %v", flag.AllowedValues)
			}
		}
		return
	}
	t.Fatal("list action not found")
}
