package registry

import "testing"

func TestShiftHandoverGenerateSummaryContract(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "shift-handover" {
			domain = candidate
			break
		}
	}
	if domain == nil {
		t.Fatal("shift-handover domain not registered")
	}
	var action *Action
	for index := range domain.Actions {
		if domain.Actions[index].Name == "generate-summary" {
			action = &domain.Actions[index]
			break
		}
	}
	if action == nil {
		t.Fatal("generate-summary action missing")
	}
	if action.HTTPMethod != "POST" || action.RESTPath != "by-guid/{handoverGuid}/generate-summary" {
		t.Fatalf("route = %s %s", action.HTTPMethod, action.RESTPath)
	}
	if len(action.Args) != 1 || action.Args[0].Name != "handoverGuid" || action.Args[0].Type != "string" {
		t.Fatalf("args = %+v", action.Args)
	}
}
