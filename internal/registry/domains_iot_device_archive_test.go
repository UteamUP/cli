package registry

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

// runIoTAction executes one iot action against a TLS test server and returns the request it received (nil when
// the command failed before sending), plus the command error.
func runIoTAction(t *testing.T, actionName string, cliArgs ...string) (*http.Request, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var received *http.Request
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Clone(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deviceGuid":"11111111-1111-4111-8111-111111111111","isActive":false,"updatedAt":"2026-11-06T12:00:00Z","items":[]}`))
	}))
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "iot-archive-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	logger := logging.New(logging.LevelError)
	format := "json"
	factory := func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
	}
	command := buildActionCommand(findDomain("iot"), findIoTAction(t, actionName), factory, logger, &format, nil)
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(cliArgs)
	return received, command.Execute()
}

func TestIoTDeviceArchiveAndRestorePostToTheGuidRoutes(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for action, suffix := range map[string]string{"device-archive": "archive", "device-restore": "restore"} {
		request, err := runIoTAction(t, action, guid)
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if request.Method != http.MethodPost || request.URL.Path != "/api/iot/devices/by-guid/"+guid+"/"+suffix {
			t.Fatalf("%s sent %s %s", action, request.Method, request.URL)
		}
		if request.URL.RawQuery != "" {
			t.Fatalf("%s leaked arguments into the query: %s", action, request.URL)
		}
	}
}

func TestIoTDeviceArchiveRefusesEmptyAndMalformedGuidsBeforeSending(t *testing.T) {
	for _, action := range []string{"device-archive", "device-restore"} {
		for _, guid := range []string{"00000000-0000-0000-0000-000000000000", "not-a-guid"} {
			request, err := runIoTAction(t, action, guid)
			if err == nil || request != nil {
				t.Fatalf("%s %s must be refused before any request; err=%v request=%v", action, guid, err, request)
			}
		}
		if request, err := runIoTAction(t, action, "11111111-1111-4111-8111-111111111111", "extra"); err == nil || request != nil {
			t.Fatalf("%s must refuse an extra argument; err=%v request=%v", action, err, request)
		}
	}
}

func TestIoTDevicesListHidesInactiveUnlessAsked(t *testing.T) {
	request, err := runIoTAction(t, "devices")
	if err != nil {
		t.Fatal(err)
	}
	if request.URL.Query().Get("isActive") != "true" {
		t.Fatalf("default list must ask for active devices, got %s", request.URL)
	}
	request, err = runIoTAction(t, "devices", "--is-active=false")
	if err != nil {
		t.Fatal(err)
	}
	if request.URL.Path != "/api/iot/devices" || request.URL.Query().Get("isActive") != "false" {
		t.Fatalf("inactive list sent %s", request.URL)
	}
}

func TestIoTSensorAndDeviceTypeArchiveAndRestorePostToTheirGuidRoutes(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	routes := map[string]string{
		"sensor-archive":      "/api/iot/sensors/" + guid + "/archive",
		"sensor-restore":      "/api/iot/sensors/" + guid + "/restore",
		"device-type-archive": "/api/iot/device-types/" + guid + "/archive",
		"device-type-restore": "/api/iot/device-types/" + guid + "/restore",
	}
	for action, path := range routes {
		request, err := runIoTAction(t, action, guid)
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if request.Method != http.MethodPost || request.URL.Path != path {
			t.Fatalf("%s sent %s %s", action, request.Method, request.URL)
		}
		if request.URL.RawQuery != "" {
			t.Fatalf("%s leaked arguments into the query: %s", action, request.URL)
		}
	}
}

func TestIoTSensorArchiveRefusesAMalformedGuidBeforeSending(t *testing.T) {
	request, err := runIoTAction(t, "sensor-archive", "not-a-guid")
	if err == nil || request != nil {
		t.Fatalf("expected a refusal before any request, got %v %v", request, err)
	}
}
