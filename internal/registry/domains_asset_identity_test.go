package registry

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAssetEquipmentIdentityPatchUsesOnlyExplicitFields(t *testing.T) {
	const guid = "9825f027-1b40-43fb-b224-1983849be941"
	request := runAssetCommand(t, "patch", guid,
		"--manufacturer", "IKEA of Sweden", "--ip-address", "192.168.68.86",
		"--mac-address", "AA:BB:CC:DD:EE:FF", "--host-name", "ha.local",
		"--web-url", "http://ha.local:8123/", "--expected-updated-at-utc", "2026-11-02T06:00:00Z")
	if request.Method != http.MethodPatch || request.URL.Path != "/api/asset/by-guid/"+guid {
		t.Fatalf("equipment patch route = %s %s", request.Method, request.URL.Path)
	}
	var body map[string]any
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"manufacturer": "IKEA of Sweden", "ipAddress": "192.168.68.86",
		"macAddress": "AA:BB:CC:DD:EE:FF", "hostName": "ha.local",
		"webUrl": "http://ha.local:8123/", "expectedUpdatedAtUtc": "2026-11-02T06:00:00Z",
	}
	if len(body) != len(want) {
		t.Fatalf("unexpected patch fields: %#v", body)
	}
	for key, value := range want {
		if body[key] != value {
			t.Errorf("%s = %#v, want %q", key, body[key], value)
		}
	}
}

func TestAssetEquipmentIdentityExplicitClearDoesNotSendOmittedFields(t *testing.T) {
	const guid = "9825f027-1b40-43fb-b224-1983849be941"
	request := runAssetCommand(t, "patch", guid, "--ip-address", "", "--web-url", "")
	var body map[string]any
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 2 || body["ipAddress"] != "" || body["webUrl"] != "" {
		t.Fatalf("explicit clears must stay distinct from omission: %#v", body)
	}
}

func TestAssetEquipmentIdentityFullUpdatePreservesFlatContract(t *testing.T) {
	request := runAssetCommand(t, "update", "9825f027-1b40-43fb-b224-1983849be941",
		"--name", "HA", "--manufacturer", "IKEA of Sweden", "--host-name", "ha.local")
	if request.Method != http.MethodPut {
		t.Fatalf("update method = %s", request.Method)
	}
	var body map[string]any
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["manufacturer"] != "IKEA of Sweden" || body["hostName"] != "ha.local" {
		t.Fatalf("full update dropped equipment identity: %#v", body)
	}
	if _, supplied := body["ipAddress"]; supplied {
		t.Fatalf("full update invented omitted IP: %#v", body)
	}
}
