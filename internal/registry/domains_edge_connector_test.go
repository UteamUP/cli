package registry

import "testing"

func TestEdgeConnectorAdministrationStaysReadOnly(t *testing.T) {
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name != "edge-connectors" {
			continue
		}
		if domain.APIPath != "/api/edge-connectors" || len(domain.Actions) != 2 {
			t.Fatal("unexpected connector surface")
		}
		for _, action := range domain.Actions {
			if action.HTTPMethod != "GET" || len(action.Flags) != 0 {
				t.Fatalf("connector action %s can mutate trust or override identity", action.Name)
			}
			if action.Name == "radio-status" && (action.RESTPath != "{connectorGuid}/radio" || len(action.Args) != 1 || action.Args[0].Type != "non-empty-uuid") {
				t.Fatal("radio status must use the public connector GUID")
			}
		}
		return
	}
	t.Fatal("connector domain missing")
}
