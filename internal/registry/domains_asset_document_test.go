package registry

import (
	"net/http"
	"testing"

	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestAssetRemoveDocumentUsesGuidDetachRoute(t *testing.T) {
	const assetGuid = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	const documentGuid = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	request := runAssetCommand(t, "remove-document", assetGuid, documentGuid)

	if request.Method != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", request.Method)
	}
	if request.URL.Path != "/api/asset/by-guid/"+assetGuid+"/documents/by-guid/"+documentGuid {
		t.Errorf("wrong detach path: %s", request.URL.Path)
	}
	if request.URL.RawQuery != "" {
		t.Errorf("GUIDs must stay in the route: %s", request.URL.RawQuery)
	}
	if request.Header.Get("X-Requested-With") != "XMLHttpRequest" {
		t.Error("detach must retain the mutation CSRF header")
	}
	if action := findAssetAction(t, "remove-document"); action.ToolName != "UteamupAssetRemoveDocument" {
		t.Errorf("tool = %q, want UteamupAssetRemoveDocument", action.ToolName)
	}
}

func TestAssetRemoveDocumentRejectsMissingInvalidOrEmptyGuidsBeforeRequest(t *testing.T) {
	const guid = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	const emptyGuid = "00000000-0000-0000-0000-000000000000"
	for _, args := range [][]string{
		{"remove-document"},
		{"remove-document", guid},
		{"remove-document", "42", guid},
		{"remove-document", guid, "42"},
		{"remove-document", emptyGuid, guid},
		{"remove-document", guid, emptyGuid},
	} {
		t.Run(args[len(args)-1], func(t *testing.T) {
			createdClient := false
			format := "json"
			command := buildDomainCommand(findAssetDomain(t), func() (*client.APIClient, error) {
				createdClient = true
				return nil, nil
			}, logging.New(logging.LevelError), &format, &ExportConfig{})
			command.SilenceErrors = true
			command.SilenceUsage = true
			command.SetArgs(args)
			if err := command.Execute(); err == nil {
				t.Error("invalid detach arguments must be refused")
			}
			if len(args) == 3 && createdClient {
				t.Error("invalid detach arguments must be refused before creating a client")
			}
		})
	}
}
