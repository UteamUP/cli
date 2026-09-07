package registry

import "testing"

func TestFleetTransportGuidAndReviewedPayloadContracts(t *testing.T) {
	t.Parallel()
	domain := findDomain("fleet-transport")
	if domain == nil || domain.APIPath != "/api/fleet/transport" || len(domain.Actions) != 5 {
		t.Fatalf("missing transport domain: %+v", domain)
	}
	for _, action := range domain.Actions {
		for _, argument := range action.Args {
			if argument.Name != "assignmentGuid" || argument.Type != "uuid" || !argument.Required {
				t.Fatalf("private or optional identity: %+v", argument)
			}
		}
		if action.Name == "create" || action.Name == "transition" || action.Name == "manifest" {
			if len(action.Flags) != 1 || !action.Flags[0].JSONFile || !action.Flags[0].Required || action.Flags[0].BodyName != "model" {
				t.Fatalf("%s must accept one explicit reviewed JSON model", action.Name)
			}
		}
		for _, flag := range action.Flags {
			if flag.Name == "tenant" || flag.Name == "tenant-id" || flag.Name == "user-id" {
				t.Fatalf("caller-controlled scope: %+v", flag)
			}
		}
	}
}
