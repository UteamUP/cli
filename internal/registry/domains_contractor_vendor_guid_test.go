package registry

import "testing"

func TestContractorDomainsExposeOnlyRealGuidTools(t *testing.T) {
	tests := []struct {
		domain  string
		actions map[string]string
	}{
		{
			domain: "contractor",
			actions: map[string]string{
				"list": "UteamupContractorProfileList",
				"get":  "UteamupContractorProfileGet",
			},
		},
		{
			domain: "contractor-workorder",
			actions: map[string]string{
				"assign":      "UteamupContractorAssign",
				"assignments": "UteamupContractorAssignmentList",
				"bids":        "UteamupContractorBidList",
				"accept-bid":  "UteamupContractorBidAccept",
			},
		},
	}

	for _, test := range tests {
		domain := findDomain(test.domain)
		if domain == nil {
			t.Fatalf("expected %s domain", test.domain)
		}
		if len(domain.Actions) != len(test.actions) {
			t.Fatalf("%s action count = %d, want %d", test.domain, len(domain.Actions), len(test.actions))
		}
		for _, action := range domain.Actions {
			wantTool, ok := test.actions[action.Name]
			if !ok {
				t.Fatalf("%s exposes nonexistent action %q", test.domain, action.Name)
			}
			if action.ToolName != wantTool {
				t.Errorf("%s %s tool = %q, want %q", test.domain, action.Name, action.ToolName, wantTool)
			}
			assertNoIntegerIdentityArgs(t, test.domain, action)
		}
	}

	contractor := findDomain("contractor")
	if contractor.Actions[0].Name != "list" || !contractor.Actions[0].MCPOnly {
		t.Fatalf("contractor list must call the real MCP list tool: %+v", contractor.Actions[0])
	}
	workorders := findDomain("contractor-workorder")
	assertRequiredModelFile(t, "contractor-workorder", actionByName(t, workorders, "assign"))
}

func TestVendorInteractionDomainsExposeGuidTools(t *testing.T) {
	tests := []struct {
		domain  string
		actions map[string]string
		models  []string
	}{
		{
			domain: "vendor-match",
			actions: map[string]string{
				"match": "UteamupVendorMatchVendors",
				"list":  "UteamupVendorMatchGet",
			},
			models: []string{"match"},
		},
		{
			domain: "vendor-message",
			actions: map[string]string{
				"send":   "UteamupVendorMessageSend",
				"list":   "UteamupVendorMessageList",
				"unread": "UteamupVendorMessageUnreadCount",
			},
			models: []string{"send"},
		},
		{
			domain: "vendor-rating",
			actions: map[string]string{
				"submit":    "UteamupVendorRatingSubmit",
				"list":      "UteamupVendorRatingList",
				"aggregate": "UteamupVendorRatingAggregate",
				"flag":      "UteamupVendorRatingFlag",
			},
			models: []string{"submit", "flag"},
		},
	}

	for _, test := range tests {
		domain := findDomain(test.domain)
		if domain == nil {
			t.Fatalf("expected %s domain", test.domain)
		}
		if len(domain.Actions) != len(test.actions) {
			t.Fatalf("%s action count = %d, want %d", test.domain, len(domain.Actions), len(test.actions))
		}
		for _, action := range domain.Actions {
			wantTool, ok := test.actions[action.Name]
			if !ok {
				t.Fatalf("%s exposes nonexistent action %q", test.domain, action.Name)
			}
			if action.ToolName != wantTool {
				t.Errorf("%s %s tool = %q, want %q", test.domain, action.Name, action.ToolName, wantTool)
			}
			assertNoIntegerIdentityArgs(t, test.domain, action)
		}
		for _, actionName := range test.models {
			assertRequiredModelFile(t, test.domain, actionByName(t, domain, actionName))
		}
	}

	if findDomain("vendor-portal") != nil {
		t.Fatal("vendor-portal advertised generic CRUD tools that do not exist")
	}
}

func actionByName(t *testing.T, domain *Domain, name string) Action {
	t.Helper()
	for _, action := range domain.Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("%s is missing action %s", domain.Name, name)
	return Action{}
}

func assertNoIntegerIdentityArgs(t *testing.T, domain string, action Action) {
	t.Helper()
	for _, argument := range action.Args {
		if argument.Name == "id" || argument.Type == "int" {
			t.Errorf("%s %s leaks integer identity argument %+v", domain, action.Name, argument)
		}
	}
}

func assertRequiredModelFile(t *testing.T, domain string, action Action) {
	t.Helper()
	if !action.MCPOnly {
		t.Errorf("%s %s must use the MCP model contract", domain, action.Name)
	}
	if len(action.Flags) != 1 {
		t.Fatalf("%s %s flags = %+v, want one model file", domain, action.Name, action.Flags)
	}
	flag := action.Flags[0]
	if flag.Name != "from-json" || flag.BodyName != "model" || !flag.Required || !flag.JSONFile {
		t.Errorf("%s %s model flag = %+v", domain, action.Name, flag)
	}
}
