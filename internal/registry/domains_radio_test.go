package registry

import (
	"strings"
	"testing"
)

func TestRadioPilotAdministrationUsesExplicitSelectedTenantBoundary(t *testing.T) {
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name != "radio-admin" {
			continue
		}
		if domain.APIPath != "/api/admin/radio" {
			t.Fatal("pilot administration requires its platform-only boundary")
		}
		expected := map[string]string{"status": "environment", "grant": "complimentary", "revoke": "complimentary/revoke"}
		for _, action := range domain.Actions {
			if action.RESTPath != expected[action.Name] {
				t.Fatalf("wrong admin route for %s", action.Name)
			}
			if len(action.Args) != 0 {
				t.Fatal("pilot commands must use the explicitly selected tenant")
			}
			if action.Name == "status" {
				if action.HTTPMethod != "GET" {
					t.Fatal("status is read-only")
				}
			} else if action.HTTPMethod != "POST" || len(action.Flags) != 1 || !action.Flags[0].Required || !action.Flags[0].RootJSONObjectFile {
				t.Fatal("pilot changes require an explicit reviewed JSON request")
			}
			delete(expected, action.Name)
		}
		if len(expected) != 0 {
			t.Fatalf("missing admin actions: %v", expected)
		}
		return
	}
	t.Fatal("radio-admin domain missing")
}

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
	expected := map[string]string{"stop-pilot": "complimentary/revoke", "transcript": "channels/{channelGuid}/transmissions/{transmissionGuid}/transcript", "transmissions": "channels/{channelGuid}/transmissions", "status": "environment", "policy": "policy", "channels": "history-channels", "recordings": "recordings", "get": "recordings/{recordingGuid}", "playback": "recordings/{recordingGuid}/playback"}
	for _, action := range domain.Actions {
		if action.RESTPath != expected[action.Name] {
			t.Errorf("wrong route for %s: %s", action.Name, action.RESTPath)
		}
		if !strings.HasPrefix(action.ToolName, "UteamupRadio") {
			t.Errorf("missing MCP equivalent for %s", action.Name)
		}
		for _, arg := range action.Args {
			if (arg.Name != "recordingGuid" && arg.Name != "channelGuid" && arg.Name != "transmissionGuid") || arg.Type != "non-empty-uuid" {
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
		if action.Name == "stop-pilot" && (action.HTTPMethod != "POST" || len(action.Flags) != 1 || !action.Flags[0].Required || !action.Flags[0].RootJSONObjectFile) {
			t.Error("stopping a pilot requires an explicit reviewed JSON request")
		}
		delete(expected, action.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing radio actions: %v", expected)
	}
}
