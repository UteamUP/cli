package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestCustomerGuidTransportAndConfirmation(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	for _, scenario := range []string{"get", "create", "update", "delete", "search", "unconfirmed", "integer", "empty-guid"} {
		t.Run(scenario, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			name := scenario
			if scenario == "unconfirmed" {
				name = "delete"
			}
			if scenario == "integer" || scenario == "empty-guid" {
				name = "get"
			}
			seen := false
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = true
				want := "/api/customer"
				if name == "get" || name == "update" || name == "delete" {
					want += "/" + guid
				}
				if r.URL.Path != want {
					t.Errorf("path = %s, want %s", r.URL.Path, want)
				}
				method := map[string]string{"get": "GET", "create": "POST", "update": "PUT", "delete": "DELETE", "search": "GET"}[name]
				if r.Method != method {
					t.Errorf("method = %s, want %s", r.Method, method)
				}
				if name == "search" && r.URL.Query().Get("nameFilter") != "Ship customer" {
					t.Error("search filter lost")
				}
				if r.URL.Query().Has("confirm") || r.URL.Query().Has("customerGuid") {
					t.Error("local or route field leaked")
				}
				if name == "create" || name == "update" {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
						return
					}
					if len(body) != 3 || body["name"] != "Reviewed customer" || body["currencyGuid"] != guid {
						t.Errorf("body changed: %v", body)
					}
					contacts, ok := body["contactGuids"].([]any)
					if !ok || len(contacts) != 1 || contacts[0] != guid {
						t.Error("contact GUID lost")
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"guid":"` + guid + `"}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "isolated-test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			logger := logging.New(logging.LevelError)
			format := "json"
			factory := func() (*client.APIClient, error) {
				return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
			}
			action := findDomainAction(t, "customer", name)
			command := buildActionCommand(findDomain("customer"), *action, factory, logger, &format, nil)
			args := []string{}
			if name == "get" || name == "update" || name == "delete" {
				args = append(args, guid)
			}
			if name == "search" {
				args = append(args, "Ship customer")
			}
			if name == "create" || name == "update" {
				file := filepath.Join(home, "customer.json")
				if err := os.WriteFile(file, []byte(`{"name":"Reviewed customer","currencyGuid":"`+guid+`","contactGuids":["`+guid+`"]}`), 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--from-json", file)
			}
			if name == "create" || name == "update" || name == "delete" {
				if scenario != "unconfirmed" {
					args = append(args, "--confirm")
				}
			}
			if scenario == "integer" {
				args[0] = "7"
			}
			if scenario == "empty-guid" {
				args[0] = "00000000-0000-0000-0000-000000000000"
			}
			command.SetArgs(args)
			err := command.Execute()
			denied := scenario == "unconfirmed" || scenario == "integer" || scenario == "empty-guid"
			if denied {
				if err == nil || seen {
					t.Fatal("invalid command must fail before transport")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !seen {
				t.Fatal("request not dispatched")
			}
		})
	}
}
