package registry

import (
	"strings"
	"testing"
)

func TestRadioRoutesUseActiveTenantAndPublicGuids(t *testing.T) {
	var domain *Domain
	for _, candidate := range DefaultRegistry.Domains() {
		if candidate.Name == "radio" {
			domain = candidate
		}
	}
	if domain == nil {
		t.Fatal("radio domain missing")
	}
	if domain.APIPath != "/api/radio" {
		t.Fatal("radio must use its explicit REST boundary")
	}
	expected := map[string]string{"transmissions": "channels/{channelGuid}/transmissions", "status": "environment", "policy": "policy", "channels": "history-channels", "recordings": "recordings", "get": "recordings/{recordingGuid}", "playback": "recordings/{recordingGuid}/playback"}
	for _, action := range domain.Actions {
		if action.RESTPath != expected[action.Name] {
			t.Errorf("wrong route for %s: %s", action.Name, action.RESTPath)
		}
		if !strings.HasPrefix(action.ToolName, "UteamupRadio") {
			t.Errorf("missing MCP equivalent for %s", action.Name)
		}
		for _, arg := range action.Args {
			if (arg.Name != "recordingGuid" && arg.Name != "channelGuid") || arg.Type != "non-empty-uuid" {
				t.Errorf("unexpected public argument %s", arg.Name)
			}
		}
		for _, flag := range action.Flags {
			if strings.Contains(normalize(flag.Name), "userid") || strings.Contains(normalize(flag.Name), "tenantid") {
				t.Errorf("identity override %s", flag.Name)
			}
		}
		if action.Name == "playback" && !action.DisableResponseExport {
			t.Error("playback URLs must not enter automatic exports")
		}
		if action.Name == "policy" && (action.HTTPMethod != "PUT" || !action.Flags[0].RootJSONObjectFile) {
			t.Error("complete reviewed policy must use a root JSON object")
		}
		delete(expected, action.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing radio actions: %v", expected)
	}
}
