package registry

import "testing"

func findGeofenceZoneAction(t *testing.T, name string) *Action {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name != "geofence-zone" {
			continue
		}
		for i := range domain.Actions {
			if domain.Actions[i].Name == name {
				return &domain.Actions[i]
			}
		}
		t.Fatalf("expected %q action on geofence-zone domain", name)
	}
	t.Fatal("expected geofence-zone domain to be registered")
	return nil
}

func TestGeofenceZoneCrudIsGuidFirst(t *testing.T) {
	for _, name := range []string{"get", "update", "delete"} {
		action := findGeofenceZoneAction(t, name)
		if len(action.Args) != 1 {
			t.Errorf("%s args = %+v, want one GUID argument", name, action.Args)
			continue
		}
		argument := action.Args[0]
		if argument.Name != "externalGuid" || argument.Type != "string" {
			t.Errorf("%s identity = %+v, want externalGuid string", name, argument)
		}
		if action.RESTPath != "by-guid/{externalGuid}" {
			t.Errorf("%s RESTPath = %q, want by-guid/{externalGuid}", name, action.RESTPath)
		}
	}
}

func TestGeofenceZoneListAndCreateHaveNoIdentityArgument(t *testing.T) {
	for _, name := range []string{"list", "create"} {
		if action := findGeofenceZoneAction(t, name); len(action.Args) != 0 {
			t.Errorf("%s args = %+v, want none", name, action.Args)
		}
	}
}
