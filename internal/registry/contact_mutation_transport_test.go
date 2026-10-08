package registry

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestContactReviewedMutationMetadataReachesNormalTransportExactly(t *testing.T) {
	for _, name := range []string{"update", "delete", "update-conflicting-file", "update-v2"} {
		t.Run(name, func(t *testing.T) {
			actionName := name
			if name == "update-conflicting-file" || name == "update-v2" {
				actionName = "update"
			}
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			const guid = "11111111-1111-4111-8111-111111111111"
			const operation = "22222222-2222-4222-8222-222222222222"
			const revision = "2026-10-07T12:00:00.1234567Z"
			seen := false
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = true
				if r.URL.Path != "/api/contact/"+guid {
					t.Errorf("wrong route %s", r.URL.Path)
				}
				fields := map[string]any{}
				if actionName == "update" {
					if r.Method != http.MethodPut {
						t.Errorf("wrong method %s", r.Method)
					}
					raw, _ := io.ReadAll(r.Body)
					if err := json.Unmarshal(raw, &fields); err != nil {
						t.Error(err)
					}
					version := float64(1)
					if name == "update-v2" {
						version = 2
						parents, ok := fields["contactTypeCreateOperationGuids"].([]any)
						if !ok || len(parents) != 1 || parents[0] != guid {
							t.Error("original parent operation changed")
						}
					}
					if fields["mutationOutcomeVersion"] != version {
						t.Error("wrong request version")
					}
				} else {
					if r.Method != http.MethodDelete {
						t.Errorf("wrong method %s", r.Method)
					}
					for key, values := range r.URL.Query() {
						if len(values) != 1 {
							t.Error("duplicate query value")
						}
						fields[key] = values[0]
					}
					if fields["mutationOutcomeVersion"] != "1" {
						t.Error("missing v1")
					}
				}
				if fields["idempotencyKey"] != operation || fields["expectedUpdatedAt"] != revision {
					t.Error("original key/revision changed")
				}
				if _, exists := fields["confirm"]; exists {
					t.Error("local confirmation leaked")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"outcome":"applied"}`))
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
			domain := findDomain("contact")
			var action Action
			for _, candidate := range domain.Actions {
				if candidate.Name == actionName {
					action = candidate
					break
				}
			}
			command := buildActionCommand(domain, action, factory, logger, &format, nil)
			args := []string{guid, "--idempotency-key", operation, "--expected-updated-at", revision, "--confirm"}
			if actionName == "update" {
				file := filepath.Join(home, "contact.json")
				fields := []byte(`{"firstName":"Reviewed","lastName":"Contact","email":"contact@example.test"}`)
				if name == "update-conflicting-file" {
					fields = []byte(`{"firstName":"Reviewed","lastName":"Contact","email":"contact@example.test","idempotencyKey":"33333333-3333-4333-8333-333333333333"}`)
				}
				if err := os.WriteFile(file, fields, 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--from-json", file)
			}
			if name == "update-v2" {
				args = append(args, "--mutation-outcome-version", "2", "--contact-type-create-operation-guids", guid)
			}
			command.SetArgs(args)
			err := command.Execute()
			if name == "update-conflicting-file" {
				if err == nil || seen {
					t.Fatal("conflicting original metadata must fail before transport")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !seen {
				t.Fatal("normal request not dispatched")
			}
		})
	}
}

func TestContactTypeOriginalIntentReachesNormalGuidTransport(t *testing.T) {
	for _, name := range []string{"create", "update", "delete"} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			const target = "11111111-1111-4111-8111-111111111111"
			const operation = "22222222-2222-4222-8222-222222222222"
			const revision = "2026-10-07T12:00:00.1234567Z"
			seen := false
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = true
				path := "/api/contacttype"
				if name != "create" {
					path += "/" + target
				}
				if r.URL.Path != path || r.Method != map[string]string{"create": "POST", "update": "PUT", "delete": "DELETE"}[name] {
					t.Errorf("wrong normal route/method: %s %s", r.Method, r.URL.Path)
				}
				fields := map[string]any{}
				if name == "delete" {
					for key, values := range r.URL.Query() {
						if len(values) != 1 {
							t.Error("duplicate original query metadata")
						}
						fields[key] = values[0]
					}
					if fields["mutationOutcomeVersion"] != "1" {
						t.Error("missing negotiated query version")
					}
				} else {
					raw, _ := io.ReadAll(r.Body)
					if err := json.Unmarshal(raw, &fields); err != nil {
						t.Error(err)
					}
					if fields["name"] != "Reviewed type" || fields["mutationOutcomeVersion"] != float64(1) {
						t.Error("original type fields/version changed")
					}
				}
				if fields["idempotencyKey"] != operation || name != "create" && fields["expectedUpdatedAt"] != revision {
					t.Error("original operation or exact seven-tick revision changed")
				}
				if _, exists := fields["confirm"]; exists {
					t.Error("local confirmation leaked into public request")
				}
				if _, exists := fields["contactTypeGuid"]; exists {
					t.Error("route identity leaked into body/query")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"schemaVersion":1,"outcome":"applied"}`))
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
			domain := findDomain("contact-type")
			var action Action
			for _, candidate := range domain.Actions {
				if candidate.Name == name {
					action = candidate
					break
				}
			}
			args := []string{"--idempotency-key", operation, "--confirm"}
			if name != "create" {
				args = append([]string{target}, args...)
				args = append(args, "--expected-updated-at", revision)
			}
			if name != "delete" {
				args = append(args, "--name", "Reviewed type")
			}
			command := buildActionCommand(domain, action, factory, logger, &format, nil)
			command.SetArgs(args)
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if !seen {
				t.Fatal("reviewed normal ContactType request was not dispatched")
			}
		})
	}
}

func TestContactLogicalParentsAreExplicitBoundedCanonicalAndConfirmed(t *testing.T) {
	const parent = "11111111-1111-4111-8111-111111111111"
	for _, action := range contactActions() {
		if action.Name != "create" && action.Name != "update" {
			continue
		}
		for _, value := range []string{parent, "", parent + "," + parent, strings.Repeat(parent+",", 128) + parent} {
			cmd := &cobra.Command{}
			cmd.Flags().Int("mutation-outcome-version", 2, "")
			cmd.Flags().StringSlice("contact-type-create-operation-guids", nil, "")
			if err := cmd.Flags().Set("contact-type-create-operation-guids", value); err != nil {
				t.Fatal(err)
			}
			err := validateContactLogicalParents(cmd, action)
			if (value == parent) != (err == nil) {
				t.Fatalf("unexpected logical validation: %v", err)
			}
		}
		confirmed := false
		for _, flag := range action.Flags {
			if flag.Name == "confirm" {
				confirmed = flag.Required && flag.MustBeTrue && flag.LocalOnly
			}
		}
		if !confirmed {
			t.Fatal("logical dependency must preserve explicit confirmation")
		}
	}
}
