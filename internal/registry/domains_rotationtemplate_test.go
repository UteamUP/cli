package registry

import "testing"

func findRotationTemplateDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "rotationtemplate" {
			return d
		}
	}
	t.Fatal("expected rotationtemplate domain to be registered")
	return nil
}

func TestRotationTemplateDomainRegistered(t *testing.T) {
	d := findRotationTemplateDomain(t)
	if d.APIPath != "/api/rotationtemplate" {
		t.Errorf("rotationtemplate APIPath = %q, want %q", d.APIPath, "/api/rotationtemplate")
	}
}

func TestRotationTemplateActionsWired(t *testing.T) {
	d := findRotationTemplateDomain(t)
	byName := map[string]*Action{}
	for i := range d.Actions {
		byName[d.Actions[i].Name] = &d.Actions[i]
	}

	list, ok := byName["list"]
	if !ok || list.HTTPMethod != "GET" {
		t.Errorf("list must be a GET action, got %+v", list)
	}

	build, ok := byName["build"]
	if !ok || build.HTTPMethod != "POST" || build.RESTPath != "{key}/build" {
		t.Fatalf("build must be POST \"{key}/build\", got %+v", build)
	}
	if len(build.Args) != 1 || build.Args[0].Name != "key" || !build.Args[0].Required {
		t.Errorf("build must take a required 'key' arg, got %+v", build.Args)
	}
	byFlag := map[string]*FlagDef{}
	for i := range build.Flags {
		byFlag[build.Flags[i].Name] = &build.Flags[i]
	}
	if a, ok := byFlag["anchor"]; !ok || a.BodyName != "anchorDate" || !a.Required {
		t.Errorf("build 'anchor' must be required → anchorDate, got %+v", a)
	}
	if d, ok := byFlag["day-shift"]; !ok || d.BodyName != "dayShiftGuid" {
		t.Errorf("build 'day-shift' must map to dayShiftGuid, got %+v", d)
	}
}
