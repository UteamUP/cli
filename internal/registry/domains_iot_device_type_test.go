package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestIoTDeviceTypeMutationsSendTheExactRESTBodyAndGuid(t *testing.T) {
	const guid = "55555555-5555-4555-8555-555555555555"
	for _, operation := range []struct{ name, method string }{
		{"device-type-create", "POST"}, {"device-type-update", "PUT"}, {"device-type-delete", "DELETE"},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			var body map[string]any
			requests := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				path := "/api/iot/device-types"
				if operation.method != "POST" {
					path += "/" + guid
				}
				if r.Method != operation.method || r.URL.Path != path {
					t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, operation.method, path)
				}
				if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
					t.Error("mutation omitted the CSRF header")
				}
				if operation.method == "DELETE" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			logger := logging.New(logging.LevelError)
			format := "json"
			factory := func() (*client.APIClient, error) {
				return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
			}
			action := findIoTAction(t, operation.name)
			command := buildActionCommand(findDomain("iot"), action, factory, logger, &format, nil)
			args := []string{}
			if operation.method != "POST" {
				args = append(args, guid)
			}
			if operation.method != "DELETE" {
				args = append(args, "--name", "IKEA KAJPLATS", "--manufacturer", "IKEA of Sweden", "--manufacturer-model", "E27", "--configuration-schema", `{"type":"object"}`)
			}
			command.SetArgs(args)
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Fatalf("sent %d requests, want1", requests)
			}
			if operation.method != "DELETE" {
				for key, value := range map[string]any{"name": "IKEA KAJPLATS", "manufacturer": "IKEA of Sweden", "description": "", "model": "E27", "configurationSchema": `{"type":"object"}`} {
					if body[key] != value {
						t.Errorf("body[%s] = %v, want %v", key, body[key], value)
					}
				}
				if _, leaked := body["deviceTypeGuid"]; leaked {
					t.Error("route identity leaked into the model")
				}
			}
		})
	}
}

func TestIoTDeviceTypeMutationsRequireNonEmptyGuidAndName(t *testing.T) {
	for _, name := range []string{"device-type-update", "device-type-delete"} {
		action := findIoTAction(t, name)
		if len(action.Args) != 1 || action.Args[0].Name != "deviceTypeGuid" || action.Args[0].Type != "non-empty-uuid" || !action.Args[0].Required {
			t.Fatalf("%s must require a non-empty public GUID", name)
		}
	}
	for _, name := range []string{"device-type-create", "device-type-update"} {
		action := findIoTAction(t, name)
		if len(action.Flags) != 5 || action.Flags[0].Name != "name" || !action.Flags[0].Required {
			t.Fatalf("%s must expose the complete type model and require its name", name)
		}
	}
}
