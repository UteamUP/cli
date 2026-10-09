package registry

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestListTemplateRequestsKeepWorkflowAndPolicyAtRootWithGuidRoute(t *testing.T) {
	const guid = "11111111-1111-4111-8111-111111111111"
	const sourceGuid = "22222222-2222-4222-8222-222222222222"
	const branchGuid = "33333333-3333-4333-8333-333333333333"
	for _, domainName := range []string{"checklist", "tasklist"} {
		t.Run(domainName, func(t *testing.T) {
			action := findDomainAction(t, domainName, "update")
			model := map[string]any{
				"name":                      "Preventive inspection",
				"workorderMaintenanceTypes": []any{float64(10), float64(12)},
			}
			if domainName == "checklist" {
				model["items"] = []any{
					map[string]any{"externalGuid": sourceGuid, "content": "Record pressure", "order": float64(1), "itemType": "Measurement"},
					map[string]any{"externalGuid": branchGuid, "content": "Inspect seal", "order": float64(2), "condition": map[string]any{
						"dependsOnItemGuid": sourceGuid, "operator": "GreaterThan", "value": "10", "branch": "Then",
					}},
				}
			} else {
				model["tasks"] = []any{map[string]any{"externalGuid": sourceGuid, "name": "Inspect seal", "order": float64(1)}}
			}
			raw, err := json.Marshal(model)
			if err != nil {
				t.Fatal(err)
			}
			file := writeRegistryJSONFixture(t, string(raw))
			request, err := readRootJSONObjectFile(file)
			if err != nil {
				t.Fatal(err)
			}
			args := map[string]any{"externalGuid": guid}
			if err := mergeRootJSONObject(args, request); err != nil {
				t.Fatal(err)
			}
			if err := validateRouteValues(*action, args); err != nil {
				t.Fatal(err)
			}
			path, consumed := buildRESTPath(findDomainByName(t, domainName), *action, args)
			if path != "/api/"+domainName+"/"+guid || action.HTTPMethod != "PUT" {
				t.Fatalf("incorrect update route: %s %s", action.HTTPMethod, path)
			}
			for _, key := range consumed {
				delete(args, key)
			}
			encoded, err := json.Marshal(args)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(encoded, raw) {
				t.Fatalf("template request changed: got %s, want %s", encoded, raw)
			}
			fileFlag := actionFlagByName(t, action, "from-json")
			if !fileFlag.Required || !fileFlag.RootJSONObjectFile {
				t.Fatal("template mutations must accept a required unwrapped typed JSON model")
			}
		})
	}
}

func TestListTemplateInvalidIdentifiersFailBeforeClientCreation(t *testing.T) {
	logger := logging.New(logging.LevelError)
	format := "json"
	for _, domainName := range []string{"checklist", "tasklist"} {
		for _, verb := range []string{"get", "update", "delete"} {
			t.Run(domainName+"/"+verb, func(t *testing.T) {
				action := findDomainAction(t, domainName, verb)
				var created bool
				command := buildActionCommand(findDomainByName(t, domainName), *action, func() (*client.APIClient, error) {
					created = true
					return nil, nil
				}, logger, &format, nil)
				quietTemplateCommand(command)
				args := []string{"5"}
				if verb == "update" {
					args = append(args, "--from-json", writeRegistryJSONFixture(t, `{"name":"Safety","workorderMaintenanceTypes":[]}`))
				}
				command.SetArgs(args)
				if err := command.Execute(); err == nil || created {
					t.Fatal("integer identifiers must fail locally before any transport")
				}
			})
		}
	}
}

func quietTemplateCommand(command *cobra.Command) {
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
}
