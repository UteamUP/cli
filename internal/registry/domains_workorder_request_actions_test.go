package registry

import (
	"net/http"
	"testing"

	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestWorkorderRequesterActionsUseOnlyGuidAndJWTTransport(t *testing.T) {
	const guid = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	for _, name := range []string{"withdraw-request", "request-closure"} {
		t.Run(name, func(t *testing.T) {
			requests, err := runRegisteredDomainCommand(t, "workorder", name, guid)
			if err != nil {
				t.Fatal(err)
			}
			if len(requests) != 1 {
				t.Fatalf("expected one request, got %d", len(requests))
			}
			request := requests[0]
			if request.method != http.MethodPost || request.path != "/api/workorder/by-guid/"+guid+"/"+name {
				t.Fatalf("unexpected route %s %s", request.method, request.path)
			}
			if len(request.body) != 0 || len(request.query) != 0 {
				t.Fatal("requester action must not accept an owner, tenant or status payload")
			}
			if request.header.Get("Authorization") != "Bearer registered-domain-token" {
				t.Fatal("must use the signed-in JWT")
			}
			if request.header.Get("X-Requested-With") != "XMLHttpRequest" {
				t.Fatal("missing mutation CSRF header")
			}
			wantTool := "UteamupWorkorderWithdrawRequest"
			if name == "request-closure" {
				wantTool = "UteamupWorkorderRequestClosure"
			}
			for _, action := range findDomain("workorder").Actions {
				if action.Name == name && action.ToolName != wantTool {
					t.Fatalf("tool = %s, want %s", action.ToolName, wantTool)
				}
			}
		})
	}
}

func TestWorkorderRequesterActionsRejectNonPublicGuids(t *testing.T) {
	for _, name := range []string{"withdraw-request", "request-closure"} {
		for _, guid := range []string{"42", "00000000-0000-0000-0000-000000000000"} {
			t.Run(name+guid, func(t *testing.T) {
				createdClient := false
				format := "json"
				command := buildDomainCommand(findDomain("workorder"), func() (*client.APIClient, error) {
					createdClient = true
					return nil, nil
				}, logging.New(logging.LevelError), &format, &ExportConfig{})
				command.SilenceErrors, command.SilenceUsage = true, true
				command.SetArgs([]string{name, guid})
				if err := command.Execute(); err == nil {
					t.Fatal("invalid GUID must fail")
				}
				if createdClient {
					t.Fatal("invalid GUID must fail before request client creation")
				}
			})
		}
	}
}
