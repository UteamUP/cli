package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestTeamDeleteRejectsIncompleteOrInvalidOriginalMetadataBeforeTransport(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	const revision = "2026-10-16T10:00:00.1234567Z"
	cases := [][]string{
		{"--idempotency-key", id},
		{"--expected-updated-at", revision, "--mutation-outcome-version", "1"},
		{"--idempotency-key", emptyUUIDValue, "--expected-updated-at", revision, "--mutation-outcome-version", "1"},
		{"--idempotency-key", id, "--expected-updated-at", revision, "--mutation-outcome-version", "2"},
	}
	for _, bad := range []string{"2026-10-16T10:00:00.12345678Z", "2026-10-16T10:00:00.1234560Z", "2026-10-16T10:00:00.0Z", "0001-01-01T00:00:00Z", "2026-10-16T10:00:00.1234567+00:00", "2026-02-30T10:00:00.1234567Z", "0001-01-01T00:00:00.0000000Z"} {
		cases = append(cases, []string{"--idempotency-key", id, "--expected-updated-at", bad, "--mutation-outcome-version", "1"})
	}
	for _, flags := range cases {
		created := false
		format := "json"
		cmd := buildDomainCommand(findDomainByName(t, "team"), func() (*client.APIClient, error) { created = true; return nil, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		cmd.SetArgs(append([]string{"delete", id}, flags...))
		if err := cmd.Execute(); err == nil || created {
			t.Fatalf("invalid metadata reached transport: %v err=%v", flags, err)
		}
	}
}

func TestTeamDeleteAcceptsCanonicalBackendRevisions(t *testing.T) {
	for _, revision := range []string{"2026-10-16T10:00:00Z", "2026-10-16T10:00:00.1Z", "2026-10-16T10:00:00.123456Z", "2026-10-16T10:00:00.1234567Z"} {
		t.Run(revision, func(t *testing.T) {
			action := findDomainAction(t, "team", "delete")
			cmd := &cobra.Command{}
			cmd.Flags().String("idempotency-key", "", "")
			cmd.Flags().String("expected-updated-at", "", "")
			cmd.Flags().Int("mutation-outcome-version", 0, "")
			for name, value := range map[string]string{"idempotency-key": "11111111-1111-4111-8111-111111111111", "expected-updated-at": revision, "mutation-outcome-version": "1"} {
				if err := cmd.Flags().Set(name, value); err != nil {
					t.Fatal(err)
				}
			}
			if err := validateTeamDeleteMetadata(cmd, *action); err != nil {
				t.Fatalf("backend revision rejected: %v", err)
			}
		})
	}
}

func TestTeamMutationMetadataUsesNormalRoutesWithoutChangingOriginal(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	const operation = "22222222-2222-4222-8222-222222222222"
	const revision = "2026-10-16T10:00:00.1234567Z"
	for _, action := range []string{"create", "update", "delete", "legacy-delete"} {
		t.Run(action, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			body := map[string]any{"name": "Reviewed", "userIds": []any{id}, "idempotencyKey": operation, "mutationOutcomeVersion": float64(1)}
			if action == "update" {
				body["expectedUpdatedAt"] = revision
			}
			payload, _ := json.Marshal(body)
			file := filepath.Join(home, "original.json")
			if err := os.WriteFile(file, payload, 0600); err != nil {
				t.Fatal(err)
			}
			count := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				path := "/api/team"
				method := "POST"
				if action != "create" {
					path += "/" + id
					method = "PUT"
				}
				if action == "delete" || action == "legacy-delete" {
					method = "DELETE"
				}
				if r.Method != method || r.URL.Path != path {
					t.Errorf("wrong normal route: %s %s", r.Method, r.URL)
				}
				if action == "delete" {
					q := r.URL.Query()
					if len(q) != 3 || q.Get("idempotencyKey") != operation || q.Get("expectedUpdatedAt") != revision || q.Get("mutationOutcomeVersion") != "1" {
						t.Errorf("changed deletion metadata: %v", q)
					}
				} else if r.URL.RawQuery != "" {
					t.Errorf("unexpected query: %s", r.URL.RawQuery)
				}
				if action == "create" || action == "update" {
					var actual map[string]any
					if err := json.NewDecoder(r.Body).Decode(&actual); err != nil {
						t.Error(err)
					}
					if !reflect.DeepEqual(actual, body) {
						t.Errorf("original changed: %#v", actual)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"schemaVersion":1,"outcome":"applied"}`))
			}))
			defer server.Close()
			if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "test", ExpiresAt: time.Now().Add(time.Hour), TenantGUID: id}); err != nil {
				t.Fatal(err)
			}
			api := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
			format := "json"
			cmd := buildDomainCommand(findDomainByName(t, "team"), func() (*client.APIClient, error) { return api, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
			args := []string{action}
			if action == "legacy-delete" {
				args[0] = "delete"
			}
			if action != "create" {
				args = append(args, id)
			}
			if action == "create" || action == "update" {
				args = append(args, "--from-json", file)
			}
			if action == "delete" {
				args = append(args, "--idempotency-key", operation, "--expected-updated-at", revision, "--mutation-outcome-version", "1")
			}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil || count != 1 {
				t.Fatalf("request failed: %v count=%d", err, count)
			}
		})
	}
}

func TestTeamActionsUseGuidRoutesAndPreserveLegacyJsonWrites(t *testing.T) {
	domain := findDomainByName(t, "team")
	if domain.APIPath != "/api/team" {
		t.Fatal("Team API route changed")
	}
	for _, name := range []string{"get", "update", "delete"} {
		action := findDomainAction(t, "team", name)
		if action.RESTPath != "{externalGuid}" || len(action.Args) != 1 || action.Args[0].Type != "non-empty-uuid" {
			t.Fatalf("%s does not use the public GUID", name)
		}
	}
	for _, name := range []string{"create", "update"} {
		fields := actionFlagByName(t, findDomainAction(t, "team", name), "from-json")
		if !fields.RootJSONObjectFile || fields.Required {
			t.Fatalf("%s changed legacy JSON option", name)
		}
	}
}
